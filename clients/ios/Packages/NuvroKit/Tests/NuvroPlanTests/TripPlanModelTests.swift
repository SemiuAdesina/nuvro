import Foundation
import Testing

@testable import NuvroModels
@testable import NuvroPlan

@Suite("TripPlanModel")
@MainActor
struct TripPlanModelTests {
    private func money(_ minor: Int64) throws -> Money { try Money(amountMinor: minor, currency: "USD") }

    private func plan(ceiling: Int64 = 300000) throws -> TripPlan {
        let destination = try DestinationMatch(
            id: "madeira", name: "Madeira", country: "Portugal", region: "Atlantic island", photoReference: "madeira",
            score: MatchScore(percent: 94), estimate: TripEstimate(low: money(264000), high: money(294000)),
            weather: WeatherSummary(condition: "Warm", temperatureCelsius: 23), flight: FlightSummary(stops: 1, approximateHours: 9),
            durationDays: 7, pace: "Easygoing", tags: ["Coastal"], reasons: []
        )
        return try TripPlan(
            id: "t", destination: destination, dateRangeLabel: "Oct 12–19", travelersLabel: "2 adults", startDate: Date(timeIntervalSince1970: 1_791_763_200), budgetCeiling: money(ceiling),
            readiness: [ReadinessItem(name: "Destination", isReady: true), ReadinessItem(name: "Flight", isReady: false)],
            budgetLines: [BudgetLine(category: "Flight", amount: money(64000)), BudgetLine(category: "Stay", amount: money(119000)), BudgetLine(category: "Food", amount: money(42000)), BudgetLine(category: "Activities", amount: money(33000)), BudgetLine(category: "Transport", amount: money(20000))],
            days: []
        )
    }

    @Test("labels are derived from the plan")
    func labels() throws {
        let model = TripPlanModel(plan: try plan(), locale: Locale(identifier: "en_US"))

        #expect(model.readinessLabel == "1 of 2 ready")
        #expect(model.subtitle == "Oct 12–19 · 7 days · 2 adults")
        #expect(model.estimateLabel == "$2,780")
        #expect(model.budgetStatus == "$220 under budget")
        #expect(model.isWithinBudget)
    }

    @Test("over budget is reported honestly")
    func overBudget() throws {
        let model = TripPlanModel(plan: try plan(ceiling: 250000), locale: Locale(identifier: "en_US"))

        #expect(model.budgetStatus == "$280 over budget")
        #expect(!model.isWithinBudget)
    }

    @Test("budget shares sum to one")
    func shares() throws {
        let model = TripPlanModel(plan: try plan())
        let total = model.plan.budgetLines.reduce(0.0) { $0 + model.share(of: $1) }

        #expect(abs(total - 1) < 0.0001)
        #expect(abs(model.share(of: model.plan.budgetLines[1]) - 1190.0 / 2780.0) < 0.0001)
    }

    @Test("toggling readiness updates the count")
    func readiness() throws {
        let model = TripPlanModel(plan: try plan())

        model.toggleReadiness(model.plan.readiness[1])
        #expect(model.readinessLabel == "2 of 2 ready")
        model.toggleReadiness(model.plan.readiness[0])
        #expect(model.readinessLabel == "1 of 2 ready")
    }

    @Test("a suggestion fills the ask field")
    func suggestion() throws {
        let model = TripPlanModel(plan: try plan())
        model.suggest("Make Day 3 more relaxing.")

        #expect(model.askText == "Make Day 3 more relaxing.")
    }
}
