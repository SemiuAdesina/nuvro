import Foundation
import NuvroModels
import Testing

@testable import NuvroPlan

@Suite("Flights states")
@MainActor
struct FlightsStateTests {
    private func option(_ tier: FlightOption.Tier) throws -> FlightOption {
        try FlightOption(tier: tier, airline: "TAP", departs: "09:00", arrives: "21:00", duration: "9h", stopsLabel: "1 stop", stopCount: 1, price: Money(amountMinor: 76000, currency: "USD"), reason: "", baggage: "", checkedLabel: "")
    }

    @Test("a failed refresh keeps fares visible but marks them stale")
    func staleAfterFailure() async throws {
        let model = FlightsModel(options: try [option(.bestOverall)], source: { throw URLError(.timedOut) })
        await model.refresh()
        #expect(model.options.count == 1)
        #expect(model.freshness == .stale)
        #expect(model.state.failure == .offline)
    }

    @Test("a successful refresh restores fresh fares")
    func freshAfterSuccess() async throws {
        let fresh = try [option(.lowestPrice)]
        let model = FlightsModel(options: [], freshness: .stale, source: { fresh })
        await model.refresh()
        #expect(model.options == fresh)
        #expect(model.freshness == .fresh)
    }

    @Test("an empty result and a loading state expose no selection")
    func emptyAndLoading() {
        #expect(FlightsModel(options: []).selected == nil)
        #expect(FlightsModel(state: .loading).options.isEmpty)
    }
}
