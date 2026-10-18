import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import SwiftUI

public struct MatchResultsView: View {
    @Environment(\.designMetrics) private var metrics
    @State private var isRevealed = false
    @State private var isComparing = false
    @State private var isHeroSaved = false
    private let state: LoadState<[DestinationMatch]>
    private let copy: MatchResultsCopy
    private let stateCopy: StateCopy
    private let locale: Locale
    private let heroNamespace: Namespace.ID
    private let onEditPreferences: () -> Void
    private let onExplore: (DestinationMatch) -> Void
    private let onRetry: () -> Void

    public init(state: LoadState<[DestinationMatch]>, language: Language = .english, locale: Locale = .current, heroNamespace: Namespace.ID, onEditPreferences: @escaping () -> Void, onExplore: @escaping (DestinationMatch) -> Void, onRetry: @escaping () -> Void = {}) {
        self.state = state
        self.copy = .resolve(language)
        self.stateCopy = .resolve(language)
        self.locale = locale
        self.heroNamespace = heroNamespace
        self.onEditPreferences = onEditPreferences
        self.onExplore = onExplore
        self.onRetry = onRetry
    }

    public init(matches: [DestinationMatch], language: Language = .english, locale: Locale = .current, heroNamespace: Namespace.ID, onEditPreferences: @escaping () -> Void, onExplore: @escaping (DestinationMatch) -> Void) {
        self.init(state: .loaded(matches), language: language, locale: locale, heroNamespace: heroNamespace, onEditPreferences: onEditPreferences, onExplore: onExplore)
    }

    private var matches: [DestinationMatch] { state.value ?? [] }

    @ViewBuilder
    private var resultStates: some View {
        switch state {
        case .idle, .loading:
            NuvroLoadingState(copy.ranking)
        case .failed(let failure, let cached) where cached == nil:
            NuvroErrorState(
                symbol: failure == .offline ? "wifi.slash" : "exclamationmark.triangle",
                title: failure == .offline ? stateCopy.offlineTitle : stateCopy.unavailableTitle,
                message: failure == .offline ? stateCopy.offlineBody : stateCopy.unavailableBody,
                retry: (stateCopy.retry, onRetry)
            )
        case .loaded(let value) where value.isEmpty:
            NuvroEmptyState(symbol: "sparkles", title: copy.noMatchesTitle, message: copy.noMatchesBody, action: (copy.adjustPreferences, onEditPreferences))
        default:
            EmptyView()
        }
    }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(18)) {
                VStack(alignment: .leading, spacing: metrics.points(6)) {
                    Text(copy.title).typeStyle(Typography.headline).foregroundStyle(Palette.label)
                    Text(copy.body).typeStyle(Typography.body).foregroundStyle(Palette.labelAlpha(0.62))
                }
                resultStates
                if let hero = matches.first {
                    HeroMatchCard(
                        match: hero, copy: copy, locale: locale, isRevealed: isRevealed, isSaved: isHeroSaved, isComparing: isComparing,
                        onExplore: { onExplore(hero) },
                        onSave: { isHeroSaved.toggle() },
                        onCompare: { isComparing.toggle() }
                    )
                    .matchedTransitionSource(id: hero.id, in: heroNamespace)
                }
                if matches.count > 1 {
                    VStack(alignment: .leading, spacing: metrics.points(10)) {
                        Text(copy.otherMatches).typeStyle(Typography.cardHeading).foregroundStyle(Palette.label).padding(.top, metrics.points(4))
                        ForEach(matches.dropFirst()) { match in
                            OtherMatchRow(match: match, locale: locale) { onExplore(match) }
                                .matchedTransitionSource(id: match.id, in: heroNamespace)
                        }
                    }
                }
            }
            .padding(.horizontal, metrics.points(20))
            .padding(.top, metrics.points(8))
            .padding(.bottom, metrics.points(60))
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button(action: onEditPreferences) {
                    Text(copy.editPreferences)
                        .typeStyle(Typography.control)
                        .foregroundStyle(Palette.label)
                        .padding(.horizontal, metrics.points(4))
                        .fixedSize()
                }
            }
        }
        .task { isRevealed = true }
    }
}

struct OtherMatchRow: View {
    @Environment(\.designMetrics) private var metrics
    let match: DestinationMatch
    let locale: Locale
    let onOpen: () -> Void

    private var meta: String {
        var parts = Array(match.tags.prefix(2))
        if let estimate = match.estimate { parts.append("~\(estimate.low.formattedWhole(locale: locale))") }
        return parts.joined(separator: " · ")
    }

    var body: some View {
        Button(action: onOpen) {
            HStack(spacing: metrics.points(12)) {
                DynamicDestinationImage(media: MediaDescriptor(url: match.photoURL, placeholder: match.photoReference, altText: match.displayName), labelSize: 8)
                    .frame(width: metrics.points(72), height: metrics.points(64))
                    .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.thumbnail), style: .continuous))
                VStack(alignment: .leading, spacing: metrics.points(3)) {
                    Text(match.displayName).typeStyle(Typography.cardTitleLarge).foregroundStyle(Palette.label)
                    Text(meta).typeStyle(Typography.caption).foregroundStyle(Palette.labelAlpha(0.6))
                }
                Spacer(minLength: 0)
                CompactMatchBadge(percent: match.score.percent)
            }
            .padding(metrics.points(10))
            .surfaceCard(radius: Radius.cardCompact)
        }
        .buttonStyle(PressScaleButtonStyle())
        .accessibilityElement(children: .combine)
    }
}
