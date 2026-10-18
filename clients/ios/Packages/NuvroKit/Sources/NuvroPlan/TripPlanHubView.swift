import NuvroDesignSystem
import NuvroModels
import SwiftUI

public struct TripPlanHubView: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(TransientFeedback.self) private var feedback: TransientFeedback?
    @Bindable private var model: TripPlanModel
    private let onViewItinerary: () -> Void
    private let onViewBudget: () -> Void
    private let onSelectModule: (PlanningModule) -> Void

    private let onTravelers: () -> Void

    public init(model: TripPlanModel, onViewItinerary: @escaping () -> Void = {}, onViewBudget: @escaping () -> Void = {}, onSelectModule: @escaping (PlanningModule) -> Void = { _ in }, onTravelers: @escaping () -> Void = {}) {
        self.model = model
        self.onViewItinerary = onViewItinerary
        self.onViewBudget = onViewBudget
        self.onSelectModule = onSelectModule
        self.onTravelers = onTravelers
    }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(18)) {
                Button { model.editing = .menu } label: {
                    TripIdentityHeader(destination: model.plan.destination, subtitle: model.subtitle)
                }
                .buttonStyle(.plain)
                .accessibilityHint(model.copy.editTrip)
                ReadinessCard(
                    title: model.copy.yourTrip,
                    label: model.readinessLabel,
                    fraction: model.plan.readinessFraction,
                    items: model.plan.readiness,
                    onToggle: model.toggleReadiness
                )
                BudgetSummaryCard(
                    title: model.copy.estimatedTrip,
                    estimate: model.estimateLabel,
                    status: model.budgetStatus,
                    isWithinBudget: model.isWithinBudget,
                    lines: model.plan.budgetLines,
                    share: model.share,
                    viewBudget: model.copy.viewBudget,
                    locale: model.locale,
                    onViewBudget: onViewBudget
                )
                ItineraryPreviewCard(title: model.copy.yourWeek, dayLabel: model.copy.day, viewItinerary: model.copy.viewItinerary, days: Array(model.plan.days.prefix(3)), onViewItinerary: onViewItinerary)
                PlanningModulesGrid(onSelect: onSelectModule)
                AskNuvroCard(
                    title: model.copy.askNuvro,
                    placeholder: model.copy.askPlaceholder,
                    suggestions: model.copy.askSuggestions,
                    text: $model.askText,
                    onSuggest: model.suggest,
                    onSend: {}
                )
            }
            .padding(.horizontal, metrics.points(20))
            .padding(.top, metrics.points(8))
            .padding(.bottom, metrics.points(40))
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .confirmationDialog(model.copy.editTrip, isPresented: Binding(get: { model.editing == .menu }, set: { if !$0, model.editing == .menu { model.editing = nil } }), titleVisibility: .visible) {
            Button(model.copy.changeDates) { model.editing = .dates }
            Button(model.copy.changeTravelers) { model.editing = .travelers }
            Button(model.copy.changeBudget) { model.editing = .budget }
            Button(model.copy.travellingTogether) { model.editing = nil; onTravelers() }
            Button(model.copy.cancel, role: .cancel) { model.editing = nil }
        }
        .sheet(item: Binding(get: { model.editing == .menu ? nil : model.editing }, set: { model.editing = $0 }), onDismiss: model.applyEdits) { presentation in
            switch presentation {
            case .dates:
                DateSelectionSheet(selection: $model.dateSelection, copy: model.dateSheetCopy, today: model.plan.startDate)
            case .travelers:
                CountPickerSheet(title: model.pickerCopy.travelers, doneLabel: model.copy.done, lines: model.travelerLines, counts: $model.travelerCounts)
            case .budget:
                TripBudgetSheet(model: model)
            case .menu:
                EmptyView()
            }
        }
        .onChange(of: model.lastEvent) { _, event in if let event { feedback?.show(event); model.clearEvent() } }
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                HStack(spacing: metrics.points(6)) {
                    Circle().fill(Palette.gold).frame(width: metrics.points(6), height: metrics.points(6))
                    Text(model.copy.planning).typeStyle(Typography.statusPill).foregroundStyle(Palette.amber)
                }
                .padding(.horizontal, metrics.points(6))
                .fixedSize()
                .accessibilityElement(children: .combine)
            }
        }
    }
}
