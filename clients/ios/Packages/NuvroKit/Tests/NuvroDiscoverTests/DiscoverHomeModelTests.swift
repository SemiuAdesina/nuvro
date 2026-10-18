import Foundation
import Testing

@testable import NuvroDiscover
@testable import NuvroModels

@Suite("DiscoverHomeModel")
@MainActor
struct DiscoverHomeModelTests {
    private func madeira() throws -> DestinationMatch {
        try DestinationMatch(
            id: "madeira", name: "Madeira", country: "Portugal", region: "Atlantic island", photoReference: "madeira",
            score: MatchScore(percent: 94),
            estimate: TripEstimate(low: Money(amountMinor: 264000, currency: "USD"), high: Money(amountMinor: 294000, currency: "USD")),
            weather: WeatherSummary(condition: "Warm", temperatureCelsius: 23),
            flight: FlightSummary(stops: 1, approximateHours: 9),
            durationDays: 7, pace: "Easygoing", tags: ["Coastal", "Food"],
            reasons: [MatchReason(text: "Longer flight than preferred", verdict: .tradeoff), MatchReason(text: "Within your budget", verdict: .fits)]
        )
    }

    @Test("greeting follows the hour of day")
    func greeting() throws {
        let model = DiscoverHomeModel(suggestions: [try madeira()])

        #expect(model.greeting(hour: 6) == "Good morning")
        #expect(model.greeting(hour: 11) == "Good morning")
        #expect(model.greeting(hour: 12) == "Good afternoon")
        #expect(model.greeting(hour: 17) == "Good afternoon")
        #expect(model.greeting(hour: 18) == "Good evening")
        #expect(model.greeting(hour: 23) == "Good evening")
    }

    @Test("inspiration themes toggle independently")
    func themes() throws {
        let model = DiscoverHomeModel(suggestions: [try madeira()])

        model.toggle(.beach)
        model.toggle(.food)
        #expect(model.isSelected(.beach) && model.isSelected(.food))
        model.toggle(.beach)
        #expect(!model.isSelected(.beach) && model.isSelected(.food))
    }

    @Test("saving a destination is reversible")
    func saving() throws {
        let destination = try madeira()
        let model = DiscoverHomeModel(suggestions: [destination])

        model.toggleSaved(destination)
        #expect(model.isSaved(destination))
        model.toggleSaved(destination)
        #expect(!model.isSaved(destination))
    }

    @Test("the card meta is derived from the match so prices agree across screens")
    func meta() throws {
        let model = DiscoverHomeModel(suggestions: [try madeira()], locale: Locale(identifier: "en_US"))

        #expect(model.meta(for: try madeira()) == "Warm · Coastal · From $2,640")
    }

    @Test("the card reason is the first positive match reason")
    func reason() throws {
        let model = DiscoverHomeModel(suggestions: [try madeira()])

        #expect(model.reason(for: try madeira()) == "Within your budget")
    }
}
