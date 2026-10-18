package retention

import "time"

const PolicyVersion = "retention-v1"

type Class string

const (
	PersonalizationSignals Class = "personalization_signals"
	CommerceHandoffs       Class = "commerce_handoffs"
	ProviderWebhookEvents  Class = "provider_webhook_events"
	TripAlerts             Class = "trip_alerts"
	ReplanProposals        Class = "replan_proposals"
	ExpiredPriceWatches    Class = "expired_price_watches"
)

type Rule struct {
	Class     Class
	MaxAge    time.Duration
	Rationale string
}

func (r Rule) CutoffFrom(now time.Time) time.Time { return now.Add(-r.MaxAge).UTC() }

const day = 24 * time.Hour

func Policy() []Rule {
	return []Rule{
		{Class: PersonalizationSignals, MaxAge: 730 * day,
			Rationale: "inference reads at most two years of behaviour, so older signals influence nothing"},
		{Class: CommerceHandoffs, MaxAge: 400 * day,
			Rationale: "the longest partner attribution window, after which a click can no longer be credited"},
		{Class: ProviderWebhookEvents, MaxAge: 180 * day,
			Rationale: "replay protection only has to outlive a partner's redelivery window"},
		{Class: TripAlerts, MaxAge: 180 * day,
			Rationale: "alerts describe a moment during a trip and are not a travel record"},
		{Class: ReplanProposals, MaxAge: 180 * day,
			Rationale: "a proposal is a transient offer to change a plan, not part of the plan"},
		{Class: ExpiredPriceWatches, MaxAge: 90 * day,
			Rationale: "a watch that has expired has no monitoring left to do"},
	}
}

func RuleFor(class Class) (Rule, bool) {
	for _, rule := range Policy() {
		if rule.Class == class {
			return rule, true
		}
	}
	return Rule{}, false
}
