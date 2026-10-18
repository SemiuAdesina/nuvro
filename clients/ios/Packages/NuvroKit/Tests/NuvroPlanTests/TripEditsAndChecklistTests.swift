import Foundation
import NuvroModels
import Testing

@testable import NuvroPlan

@Suite("Trip edits")
@MainActor
struct TripEditTests {
    private func plan() throws -> TripPlan {
        let destination = try DestinationMatch(
            id: "madeira", name: "Madeira", country: "Portugal", region: "Atlantic", photoReference: "madeira", score: MatchScore(percent: 94),
            estimate: TripEstimate(low: Money(amountMinor: 1, currency: "USD"), high: Money(amountMinor: 2, currency: "USD")),
            weather: WeatherSummary(condition: "Warm", temperatureCelsius: 23), flight: FlightSummary(stops: 1, approximateHours: 9), durationDays: 7, pace: "Easy", tags: [], reasons: []
        )
        return try TripPlan(id: "t", destination: destination, dateRangeLabel: "Oct 12–19", travelersLabel: "2 adults", startDate: Date(timeIntervalSince1970: 1_791_763_200), budgetCeiling: Money(amountMinor: 300000, currency: "USD"), readiness: [], budgetLines: [], days: [])
    }

    @Test("editing travellers and budget rewrites the plan and reports once")
    func edits() throws {
        let model = TripPlanModel(plan: try plan(), locale: Locale(identifier: "en_US"))
        model.applyEdits()
        #expect(model.lastEvent == nil)
        model.travelerCounts = ["adults": 2, "children": 1, "infants": 0]
        model.budgetMajor = 3500
        model.applyEdits()
        #expect(model.plan.travelersLabel == "2 adults, 1 child")
        #expect(model.plan.budgetCeiling.amountMinor == 350000)
        #expect(model.lastEvent == "Trip updated")
    }

    @Test("flexible dates rewrite the date label")
    func dates() throws {
        let model = TripPlanModel(plan: try plan(), locale: Locale(identifier: "en_US"))
        model.dateSelection = .flexible(month: Date(timeIntervalSince1970: 1_791_763_200), nights: 5)
        model.applyEdits()
        #expect(model.plan.dateRangeLabel.contains("Anytime in"))
        #expect(model.plan.dateRangeLabel.contains("5 nights"))
    }
}

@Suite("Checklist items")
@MainActor
struct ChecklistItemTests {
    private func model() -> ChecklistModel {
        ChecklistModel(groups: [ChecklistGroup(name: "Documents", items: [ChecklistItem(name: "Passport", isDone: true), ChecklistItem(name: "Insurance")])], insight: "")
    }

    @Test("custom items are added once and can be removed")
    func addAndRemove() {
        let checklist = model()
        let group = checklist.groups[0]
        checklist.beginAdding(to: group)
        checklist.newItemName = " Adapter "
        checklist.commitNewItem()
        #expect(checklist.groups[0].items.map(\.name) == ["Passport", "Insurance", "Adapter"])
        #expect(checklist.addingTo == nil)
        checklist.beginAdding(to: checklist.groups[0])
        checklist.newItemName = "adapter"
        checklist.commitNewItem()
        #expect(checklist.groups[0].items.count == 3)
        checklist.remove(checklist.groups[0].items[2], from: checklist.groups[0])
        #expect(checklist.groups[0].items.count == 2)
    }

    @Test("empty names are ignored and completion is derived")
    func completion() {
        let checklist = model()
        checklist.beginAdding(to: checklist.groups[0])
        checklist.newItemName = "   "
        checklist.commitNewItem()
        #expect(checklist.groups[0].items.count == 2)
        #expect(!checklist.isComplete(checklist.groups[0]))
        #expect(!checklist.isAllDone)
        checklist.toggle(checklist.groups[0].items[1], in: checklist.groups[0])
        #expect(checklist.isComplete(checklist.groups[0]))
        #expect(checklist.isAllDone)
    }
}
