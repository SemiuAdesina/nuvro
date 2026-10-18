import NuvroDesignSystem
import NuvroModels
import SwiftUI

private enum DiscoverPreviewFixtures {
    static let madeira: DestinationMatch = {
        let low = try! Money(amountMinor: 264000, currency: "USD")
        let high = try! Money(amountMinor: 294000, currency: "USD")
        return try! DestinationMatch(
            id: "madeira", name: "Madeira", country: "Portugal", region: "Atlantic island", photoReference: "madeira", score: MatchScore(percent: 94),
            estimate: TripEstimate(low: low, high: high), weather: WeatherSummary(condition: "Warm", temperatureCelsius: 23), flight: FlightSummary(stops: 1, approximateHours: 9),
            durationDays: 7, pace: "Easygoing", tags: ["Coastal"], reasons: [MatchReason(text: "Within your budget", verdict: .fits)]
        )
    }()

    @MainActor
    static func home(_ state: LoadState<[DestinationMatch]>) -> some View {
        NavigationStack {
            DiscoverHomeView(model: DiscoverHomeModel(state: state, locale: Locale(identifier: "en_US")), onStartSearch: {}, onExploreByBudget: {}, onSeeAll: {}, onOpen: { _ in })
        }
    }
}

#Preview("Discover loading") {
    DesignCanvas(fullBleed: false) { _ in DiscoverPreviewFixtures.home(.loading) }
}

#Preview("Discover offline, nothing cached") {
    DesignCanvas(fullBleed: false) { _ in DiscoverPreviewFixtures.home(.failed(.offline, cached: nil)) }
}

#Preview("Discover empty") {
    DesignCanvas(fullBleed: false) { _ in DiscoverPreviewFixtures.home(.loaded([])) }
}

#Preview("Discover refreshing with cached picks") {
    DesignCanvas(fullBleed: false) { _ in DiscoverPreviewFixtures.home(.refreshing([DiscoverPreviewFixtures.madeira])) }
}

#Preview("Match results ranking") {
    struct Host: View {
        @Namespace private var namespace
        var body: some View {
            NavigationStack { MatchResultsView(state: .loading, locale: Locale(identifier: "en_US"), heroNamespace: namespace, onEditPreferences: {}, onExplore: { _ in }) }
        }
    }
    return DesignCanvas(fullBleed: false) { _ in Host() }
}

#Preview("Match results empty, large text") {
    struct Host: View {
        @Namespace private var namespace
        var body: some View {
            NavigationStack { MatchResultsView(state: .loaded([]), locale: Locale(identifier: "en_US"), heroNamespace: namespace, onEditPreferences: {}, onExplore: { _ in }) }
        }
    }
    return DesignCanvas(fullBleed: false) { _ in Host() }.dynamicTypeSize(.accessibility2)
}
