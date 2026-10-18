package replanning

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/itinerary"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/money"
	"github.com/SemiuAdesina/nuvro/backend/internal/domain/replan"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

const AlgorithmVersion = "replan-v1"
const ProposalLifetime = 2 * time.Hour
const maximumChanges = 4

type Engine struct {
	places      PlaceFinder
	routes      RouteFinder
	weather     WeatherReader
	interpreter Interpreter
	timer       clock.Clock
}

func New(places PlaceFinder, routes RouteFinder, weatherReader WeatherReader, interpreter Interpreter, timer clock.Clock) *Engine {
	return &Engine{places: places, routes: routes, weather: weatherReader, interpreter: interpreter, timer: timer}
}

func (e *Engine) Propose(ctx context.Context, request Request) replan.Proposal {
	intent := e.intent(ctx, request)
	if intent.Trigger.Valid() && request.Trigger == replan.Manual {
		request.Trigger = intent.Trigger
	}
	if intent.LowerPace {
		request.ReducedWalking = request.ReducedWalking || intent.PreferNear
	}
	now := e.timer.Now().UTC()
	proposal := replan.Proposal{
		TripID: request.TripID, Trigger: request.Trigger,
		ItineraryVersion: request.Itinerary.Version,
		Preserved:        preservedConstraints(request),
		CreatedAt:        now, ExpiresAt: now.Add(ProposalLifetime),
	}
	for _, blocked := range e.conflicts(request) {
		if len(proposal.Changes) == maximumChanges {
			break
		}
		change := e.changeFor(ctx, request, blocked, intent)
		proposal.Changes = append(proposal.Changes, change)
		proposal.Reasons = append(proposal.Reasons, change.Reasons...)
	}
	proposal.Reasons = unique(proposal.Reasons)
	if len(proposal.Changes) == 0 {
		proposal.Reasons = []replan.Reason{replan.NoAlternativeFound}
		return proposal
	}
	proposal.Violations = e.validate(request, proposal)
	proposal.BudgetDelta, proposal.HasBudgetDelta = budgetDelta(request, proposal)
	proposal.Summary = summarise(proposal)
	return proposal
}

func (e *Engine) intent(ctx context.Context, request Request) Intent {
	if e.interpreter == nil || strings.TrimSpace(request.Message) == "" {
		return Intent{Trigger: request.Trigger, LowerPace: request.Trigger == replan.Tired}
	}
	intent, err := e.interpreter.Interpret(ctx, request.Message)
	if err != nil || !intent.Trigger.Valid() {
		return Intent{Trigger: request.Trigger, LowerPace: request.Trigger == replan.Tired}
	}
	for _, category := range intent.Categories {
		if !category.Valid() {
			return Intent{Trigger: request.Trigger, LowerPace: request.Trigger == replan.Tired}
		}
	}
	return intent
}

func (e *Engine) validate(request Request, proposal replan.Proposal) []itinerary.Violation {
	plan := Apply(request.Itinerary, proposal)
	shape := itinerary.Constraints{BufferMinutes: 15, Locked: lockedItems(request.Itinerary)}
	if request.Itinerary.Timezone != "" {
		plan.Timezone = request.Itinerary.Timezone
	}
	return itinerary.Validate(plan, shape)
}

func Apply(plan itinerary.Itinerary, proposal replan.Proposal) itinerary.Itinerary {
	items := slices.Clone(plan.Items)
	for _, change := range proposal.Changes {
		switch change.Kind {
		case replan.Remove:
			items = withoutItem(items, change.Original.ID)
		case replan.Replace, replan.Move:
			items = withoutItem(items, change.Original.ID)
			items = append(items, *change.Proposed)
		case replan.Add:
			items = append(items, *change.Proposed)
		}
	}
	slices.SortStableFunc(items, func(left, right itinerary.Item) int { return left.StartAt.Compare(right.StartAt) })
	plan.Items = items
	return plan
}

func withoutItem(items []itinerary.Item, id string) []itinerary.Item {
	remaining := make([]itinerary.Item, 0, len(items))
	for _, item := range items {
		if item.ID != id {
			remaining = append(remaining, item)
		}
	}
	return remaining
}

func lockedItems(plan itinerary.Itinerary) map[string]itinerary.Item {
	locked := make(map[string]itinerary.Item)
	for _, item := range plan.LockedItems() {
		locked[item.ID] = item
	}
	return locked
}

func preservedConstraints(request Request) []replan.Constraint {
	preserved := []replan.Constraint{replan.LockedItems, replan.TripDates}
	if len(request.Itinerary.LockedItems()) > 0 {
		preserved = append(preserved, replan.Reservations)
	}
	if request.BudgetCeiling != nil {
		preserved = append(preserved, replan.BudgetCeiling)
	}
	if request.ReducedWalking {
		preserved = append(preserved, replan.Accessibility)
	}
	if request.Pace != "" {
		preserved = append(preserved, replan.TravelPace)
	}
	return preserved
}

func budgetDelta(request Request, proposal replan.Proposal) (money.Amount, bool) {
	if request.Currency == "" {
		return money.Amount{}, false
	}
	currency, err := money.ParseCurrency(request.Currency)
	if err != nil {
		return money.Amount{}, false
	}
	delta := int64(0)
	for _, change := range proposal.Changes {
		if change.Original != nil {
			delta -= change.Original.EstimatedCostMinor
		}
		if change.Proposed != nil {
			delta += change.Proposed.EstimatedCostMinor
		}
	}
	amount, err := money.New(delta, currency)
	if err != nil {
		return money.Amount{}, false
	}
	return amount, true
}

func unique(reasons []replan.Reason) []replan.Reason {
	seen := make(map[replan.Reason]struct{}, len(reasons))
	ordered := make([]replan.Reason, 0, len(reasons))
	for _, reason := range reasons {
		if _, duplicate := seen[reason]; duplicate {
			continue
		}
		seen[reason] = struct{}{}
		ordered = append(ordered, reason)
	}
	return ordered
}

func summarise(proposal replan.Proposal) string {
	switch {
	case len(proposal.Changes) == 1 && proposal.Changes[0].Kind == replan.Replace:
		return "Replaces " + proposal.Changes[0].Original.Title + " with " + proposal.Changes[0].Proposed.Title
	case len(proposal.Changes) == 1 && proposal.Changes[0].Kind == replan.Move:
		return "Moves " + proposal.Changes[0].Original.Title + " to a later slot"
	case len(proposal.Changes) == 1:
		return "Removes " + proposal.Changes[0].Original.Title
	default:
		return "Adjusts the affected part of the plan"
	}
}
