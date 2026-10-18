import NuvroDesignSystem
import NuvroModels
import SwiftUI

private enum PlanPreviewFixtures {
    static let fare: FlightOption = {
        let price = try! Money(amountMinor: 76000, currency: "USD")
        return FlightOption(tier: .bestOverall, airline: "TAP Air Portugal", departs: "21:40", arrives: "09:10", duration: "9h 30m", stopsLabel: "1 stop · Lisbon", stopCount: 1, price: price, reason: "Shortest total journey inside your budget", baggage: "1 checked bag", checkedLabel: "Fares checked 2h ago")
    }()
}

#Preview("Flights loading") {
    DesignCanvas(fullBleed: false) { _ in
        NavigationStack { FlightsView(model: FlightsModel(state: .loading, locale: Locale(identifier: "en_US")), routeLabel: "New York → Madeira", travelers: "2 adults") }
    }
}

#Preview("Flights provider unavailable") {
    DesignCanvas(fullBleed: false) { _ in
        NavigationStack { FlightsView(model: FlightsModel(state: .failed(.unavailable, cached: nil), locale: Locale(identifier: "en_US")), routeLabel: "New York → Madeira", travelers: "2 adults") }
    }
}

#Preview("Flights stale fares") {
    DesignCanvas(fullBleed: false) { _ in
        NavigationStack { FlightsView(model: FlightsModel(options: [PlanPreviewFixtures.fare], freshness: .stale, locale: Locale(identifier: "en_US")), routeLabel: "New York → Madeira", travelers: "2 adults") }
    }
}

#Preview("Flights no results") {
    DesignCanvas(fullBleed: false) { _ in
        NavigationStack { FlightsView(model: FlightsModel(options: [], locale: Locale(identifier: "en_US")), routeLabel: "New York → Madeira", travelers: "2 adults") }
    }
}

#Preview("Trips library empty") {
    DesignCanvas(fullBleed: false) { _ in
        NavigationStack { TripsLibraryView(model: TripsLibraryModel(entries: []), onContinueTrip: {}, onOpenPlan: {}, onOpenPast: {}, onPlanTrip: {}) }
    }
}
