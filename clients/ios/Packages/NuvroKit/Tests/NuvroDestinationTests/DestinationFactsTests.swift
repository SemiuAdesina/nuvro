import Foundation
import Testing

@testable import NuvroDestination
@testable import NuvroLocalization
@testable import NuvroModels

@Suite("DestinationFacts")
struct DestinationFactsTests {
    private let locale = Locale(identifier: "en_US")

    private func madeira() throws -> DestinationMatch {
        try DestinationMatch(
            id: "madeira", name: "Madeira", country: "Portugal", region: "Atlantic island", photoReference: "madeira",
            score: MatchScore(percent: 94),
            estimate: TripEstimate(low: Money(amountMinor: 264000, currency: "USD"), high: Money(amountMinor: 294000, currency: "USD")),
            weather: WeatherSummary(condition: "Warm", temperatureCelsius: 23),
            flight: FlightSummary(stops: 1, approximateHours: 9),
            durationDays: 7, pace: "Easygoing", tags: ["Coastal", "Food"], reasons: []
        )
    }

    @Test("glance facts read length, trip, climate, flight and pace in order")
    func glance() throws {
        let facts = DestinationFacts.glance(for: try madeira(), estimate: try Money(amountMinor: 278000, currency: "USD"), copy: .english, locale: locale)

        #expect(facts.map(\.key) == ["Length", "Trip", "Climate", "Flight", "Pace"])
        #expect(facts.map(\.value) == ["7 days", "~$2,780", "Warm", "1 stop", "Easygoing"])
    }

    @Test("the trip fact is omitted when no estimate exists")
    func glanceWithoutEstimate() throws {
        let facts = DestinationFacts.glance(for: try madeira(), estimate: nil, copy: .english, locale: locale)

        #expect(!facts.map(\.key).contains("Trip"))
    }

    @Test("reasons compute the budget margin from the ceiling and estimate")
    func reasons() throws {
        let facts = DestinationFacts.reasons(
            for: try madeira(),
            budgetCeiling: try Money(amountMinor: 300000, currency: "USD"),
            estimate: try Money(amountMinor: 278000, currency: "USD"),
            copy: .english, locale: locale
        )

        #expect(facts.first == DestinationFact(key: "Budget", value: "$220 under"))
        #expect(facts.map(\.key) == ["Budget", "Weather", "Style match", "Flight"])
        #expect(facts[2].value == "Coastal + Food")
        #expect(facts[3].value == "9h · 1 stop")
    }

    @Test("an over-budget destination shows no budget reason rather than a false one")
    func overBudget() throws {
        let facts = DestinationFacts.reasons(
            for: try madeira(),
            budgetCeiling: try Money(amountMinor: 200000, currency: "USD"),
            estimate: try Money(amountMinor: 278000, currency: "USD"),
            copy: .english, locale: locale
        )

        #expect(!facts.map(\.key).contains("Budget"))
    }

    @Test("plural stops are labelled correctly")
    func stops() {
        #expect(DestinationFacts.stopsLabel(FlightSummary(stops: 2, approximateHours: 12), copy: .english) == "2 stops")
        #expect(DestinationFacts.stopsLabel(FlightSummary(stops: 0, approximateHours: 3), copy: .english) == "0 stops")
    }

    @Test("the summary sentence is built from the match, not typed by hand")
    func summary() throws {
        #expect(DestinationFacts.summary(for: try madeira()) == "Madeira matches your budget, warm-weather preference, coastal and food interests, and easygoing travel style.")
    }
}
