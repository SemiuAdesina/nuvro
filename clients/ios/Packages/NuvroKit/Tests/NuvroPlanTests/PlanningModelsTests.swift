import Foundation
import Testing

@testable import NuvroModels
@testable import NuvroPlan

@Suite("Planning models")
@MainActor
struct PlanningModelsTests {
    @Test("checklist progress counts done items across groups")
    func checklist() {
        let model = ChecklistModel(groups: [
            ChecklistGroup(name: "Documents", items: [ChecklistItem(name: "Passport", isDone: true), ChecklistItem(name: "Insurance")]),
            ChecklistGroup(name: "Pack", items: [ChecklistItem(name: "Shoes", isDone: true)])
        ], insight: "")

        #expect(model.readyLabel == "2 of 3 ready")
        #expect(abs(model.fraction - 2.0 / 3.0) < 0.001)
        model.toggle(model.groups[0].items[1], in: model.groups[0])
        #expect(model.readyLabel == "3 of 3 ready")
        #expect(model.countLabel(model.groups[0]) == "2 of 2")
    }

    @Test("voting adds and removes my vote and updates the leader")
    func voting() {
        let decision = GroupDecision(question: "Where?", meta: "", options: [DecisionOption(name: "A", votes: 2, hasMyVote: true), DecisionOption(name: "B", votes: 1, hasMyVote: false)])
        let model = GroupModel(members: [], decisions: [decision], changes: [], budgetLabel: "", budgetNote: "")

        model.vote(decision.options[1], in: decision)
        #expect(model.decisions[0].options[1].votes == 2)
        #expect(model.decisions[0].options[1].hasMyVote)
        model.vote(decision.options[0], in: decision)
        #expect(model.decisions[0].options[0].votes == 1)
        #expect(model.decisions[0].leadingVotes == 2)
        #expect(model.votesLabel(model.decisions[0].options[0]) == "1 vote")
    }

    @Test("flights expose the selected tier and the remaining options")
    func flights() throws {
        let price = try Money(amountMinor: 74200, currency: "USD")
        let model = FlightsModel(options: [
            FlightOption(tier: .bestOverall, airline: "TAP", departs: "", arrives: "", duration: "", stopsLabel: "", stopCount: 1, price: price, reason: "", baggage: "", checkedLabel: ""),
            FlightOption(tier: .lowestPrice, airline: "Iberia", departs: "", arrives: "", duration: "", stopsLabel: "", stopCount: 2, price: price, reason: "", baggage: "", checkedLabel: "")
        ], locale: Locale(identifier: "en_US"))

        #expect(model.selected?.airline == "TAP")
        #expect(model.others.map(\.airline) == ["Iberia"])
        model.selectedTier = FlightOption.Tier.lowestPrice
        #expect(model.selected?.airline == "Iberia")
        #expect(model.price(model.options[0]) == "$742")
    }

    @Test("practical sections open one at a time and transport preselects the plan's best option")
    func practicalAndTransport() {
        let sections = [InfoSection(name: "Money", gist: "", rows: [], source: "", placeholder: "gold"), InfoSection(name: "Safety", gist: "", rows: [], source: "", placeholder: "gold")]
        let practical = PracticalModel(entryCheckedLabel: "", entryRows: [], sections: sections, emergency: [], openSection: "Money")
        #expect(practical.isOpen(sections[0]))
        practical.toggle(sections[1])
        #expect(practical.isOpen(sections[1]) && !practical.isOpen(sections[0]))
        practical.toggle(sections[1])
        #expect(practical.openSection == nil)

        let guide = TransportGuide(recommendationTitle: "", recommendationBody: "", arrivals: [ArrivalOption(name: "Transfer", meta: "", costLabel: "", tag: .recommended, placeholder: "aqua"), ArrivalOption(name: "Car", meta: "", costLabel: "", tag: .bestForPlan, placeholder: "sky")], transferLabel: "", airport: GeoCoordinate(latitude: 0, longitude: 0), hotel: GeoCoordinate(latitude: 0, longitude: 0), local: [])
        let transport = TransportModel(guide: guide)
        #expect(transport.selectedArrival == "Car")
    }
}
