import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import SwiftUI

public struct DiscoverHomeView: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(\.calendar) private var calendar
    @Environment(TransientFeedback.self) private var feedback: TransientFeedback?
    private let model: DiscoverHomeModel
    private let stateCopy = StateCopy.resolve(.english)
    private let onStartSearch: () -> Void
    private let onExploreByBudget: () -> Void
    private let onSeeAll: () -> Void
    private let onOpen: (DestinationMatch) -> Void

    public init(
        model: DiscoverHomeModel,
        onStartSearch: @escaping () -> Void,
        onExploreByBudget: @escaping () -> Void,
        onSeeAll: @escaping () -> Void,
        onOpen: @escaping (DestinationMatch) -> Void
    ) {
        self.model = model
        self.onStartSearch = onStartSearch
        self.onExploreByBudget = onExploreByBudget
        self.onSeeAll = onSeeAll
        self.onOpen = onOpen
    }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(22)) {
                header
                ConversationalEntry(placeholder: model.copy.searchPlaceholder, onTap: onStartSearch, onVoice: onStartSearch)
                SignatureHeroCard(title: model.copy.heroTitle, subtitle: model.copy.heroBody, action: model.copy.heroAction, onAction: onExploreByBudget)
                InspirationChips(themes: InspirationTheme.allCases, isSelected: model.isSelected, onToggle: model.toggle)
                    .padding(.horizontal, -metrics.points(20))
                recommended
            }
            .padding(.horizontal, metrics.points(20))
            .padding(.top, metrics.points(8))
            .padding(.bottom, metrics.points(24))
        }
        .background(Palette.background.ignoresSafeArea())
        .refreshable { await model.refresh() }
        .onChange(of: model.lastEvent) { _, event in if let event { feedback?.show(event); model.clearEvent() } }
        .scrollIndicators(.hidden)
    }

    private var header: some View {
        VStack(alignment: .leading, spacing: metrics.points(4)) {
            Text(model.greeting(hour: calendar.component(.hour, from: .now)))
                .typeStyle(Typography.greeting)
                .foregroundStyle(Palette.labelAlpha(0.6))
            Text(model.copy.headline)
                .typeStyle(Typography.headline)
                .foregroundStyle(Palette.label)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    private var recommended: some View {
        VStack(alignment: .leading, spacing: metrics.points(14)) {
            SectionHeader(model.copy.recommendedTitle, action: model.copy.seeAll, onAction: onSeeAll)
            switch model.suggestionsState {
            case .loading, .idle:
                NuvroLoadingState(stateCopy.loading)
            case .failed(let failure, let cached) where cached == nil:
                NuvroErrorState(
                    symbol: failure == .offline ? "wifi.slash" : "exclamationmark.triangle",
                    title: failure == .offline ? stateCopy.offlineTitle : stateCopy.unavailableTitle,
                    message: failure == .offline ? stateCopy.offlineBody : stateCopy.unavailableBody,
                    retry: (stateCopy.retry, { Task { await model.refresh() } })
                )
            case .loaded(let value) where value.isEmpty:
                NuvroEmptyState(symbol: "sparkles", title: model.copy.noSuggestionsTitle, message: model.copy.noSuggestionsBody, action: (model.copy.heroAction, onStartSearch))
            default:
                EmptyView()
            }
            if model.suggestionsState.isLoading, !model.suggestions.isEmpty {
                NuvroLoadingState(stateCopy.loading, isCompact: true)
            }
            ForEach(model.suggestions) { destination in
                DestinationPickCard(
                    destination: destination,
                    meta: model.meta(for: destination),
                    reason: model.reason(for: destination),
                    isSaved: model.isSaved(destination),
                    onOpen: { onOpen(destination) },
                    onSave: { model.toggleSaved(destination) }
                )
            }
            Text(model.copy.priceDisclaimer)
                .typeStyle(Typography.small)
                .foregroundStyle(Palette.labelAlpha(0.45))
                .frame(maxWidth: .infinity)
        }
    }
}
