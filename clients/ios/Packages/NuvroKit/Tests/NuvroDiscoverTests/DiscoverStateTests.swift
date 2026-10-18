import Foundation
import NuvroModels
import Testing

@testable import NuvroDiscover

@Suite("Discover suggestion states")
@MainActor
struct DiscoverStateTests {
    private func match(_ id: String) throws -> DestinationMatch {
        try DestinationMatch(
            id: id, name: id.capitalized, country: "Portugal", region: "", photoReference: "madeira", score: MatchScore(percent: 90),
            estimate: TripEstimate(low: Money(amountMinor: 100, currency: "USD"), high: Money(amountMinor: 200, currency: "USD")),
            weather: WeatherSummary(condition: "Warm", temperatureCelsius: 22), flight: FlightSummary(stops: 0, approximateHours: 2), durationDays: 5, pace: "Easy", tags: [], reasons: []
        )
    }

    @Test("a refresh keeps existing picks visible and replaces them on success")
    func refreshSuccess() async throws {
        let fresh = try [match("lisbon")]
        let model = DiscoverHomeModel(suggestions: try [match("madeira")], source: { fresh })
        await model.refresh()
        #expect(model.suggestions.map(\.id) == ["lisbon"])
        #expect(model.suggestionsState == .loaded(fresh))
    }

    @Test("a failed refresh keeps the cached picks and records the failure")
    func refreshFailure() async throws {
        let cached = try [match("madeira")]
        let model = DiscoverHomeModel(suggestions: cached, source: { throw URLError(.notConnectedToInternet) })
        await model.refresh()
        #expect(model.suggestions == cached)
        #expect(model.suggestionsState.failure == .offline)
    }

    @Test("without a source refresh is a no-op and saving reports feedback")
    func noSourceAndFeedback() async throws {
        let picks = try [match("madeira")]
        let model = DiscoverHomeModel(suggestions: picks)
        await model.refresh()
        #expect(model.suggestionsState == .loaded(picks))
        model.toggleSaved(picks[0])
        #expect(model.lastEvent == "Saved")
        model.toggleSaved(picks[0])
        #expect(model.lastEvent == "Removed from saved")
    }
}
