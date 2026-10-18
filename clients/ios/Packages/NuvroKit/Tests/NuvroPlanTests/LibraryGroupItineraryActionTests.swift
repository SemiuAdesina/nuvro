import Foundation
import NuvroModels
import Testing

@testable import NuvroPlan

@Suite("Trips library actions")
@MainActor
struct TripsLibraryModelTests {
    private func entry(_ id: String, phase: TripLibraryEntry.Phase) -> TripLibraryEntry {
        TripLibraryEntry(id: id, destinationName: id.capitalized, dateLabel: "Oct", phase: phase, progress: 0.5, readinessLabel: nil, media: DestinationMedia(placeholder: "gold", altText: id), statusLabel: nil, weatherLabel: nil, nextLabel: nil)
    }

    @Test("archiving and deleting remove entries and report the event")
    func archiveAndDelete() {
        let model = TripsLibraryModel(entries: [entry("madeira", phase: .active), entry("lisbon", phase: .upcoming)])
        model.archive(model.entries[1])
        #expect(model.entries(in: .upcoming).isEmpty)
        #expect(model.lastEvent == "Trip archived")
        model.requestDelete(model.entries[0])
        #expect(model.pendingDelete?.id == "madeira")
        model.cancelDelete()
        #expect(model.entries.count == 1)
        model.requestDelete(model.entries[0])
        model.confirmDelete()
        #expect(model.isEmpty)
        #expect(model.lastEvent == "Trip deleted")
    }

    @Test("downloading marks the trip offline once")
    func download() {
        let model = TripsLibraryModel(entries: [entry("madeira", phase: .active)])
        #expect(!model.isDownloaded(model.entries[0]))
        model.download(model.entries[0])
        #expect(model.isDownloaded(model.entries[0]))
        #expect(model.lastEvent == "Downloaded for offline")
    }
}

@Suite("Group planning actions")
@MainActor
struct GroupModelActionTests {
    private func model() -> GroupModel {
        GroupModel(
            members: [GroupMember(name: "Alex", subtitle: "alex@example.com", role: .owner, initial: "A", placeholder: "sky"), GroupMember(name: "Sam", subtitle: "sam@example.com", role: .editor, initial: "S", placeholder: "gold")],
            decisions: [], changes: [], budgetLabel: "$4,820", budgetNote: ""
        )
    }

    @Test("owners cannot be removed, others need confirmation")
    func removal() {
        let group = model()
        group.requestRemove(group.members[0])
        #expect(group.presentation == nil)
        group.requestRemove(group.members[1])
        #expect(group.presentation == .removeConfirmation(group.members[1]))
        group.confirmRemove()
        #expect(group.members.count == 1)
        #expect(group.lastEvent == "Traveller removed")
    }

    @Test("roles change for non-owners and leaving is confirmed")
    func rolesAndLeaving() {
        let group = model()
        group.setRole(.viewer, for: group.members[1])
        #expect(group.members[1].role == .viewer)
        group.setRole(.viewer, for: group.members[0])
        #expect(group.members[0].role == .owner)
        group.requestLeave()
        #expect(group.presentation == .leaveConfirmation)
        group.confirmLeave()
        #expect(group.hasLeft)
    }

    @Test("invitations need a plausible new address and add a pending member")
    func invite() {
        let group = model()
        group.inviteEmail = "sam@example.com"
        #expect(!group.canInvite)
        group.inviteEmail = "jo@example.com"
        #expect(group.canInvite)
        group.sendInvite()
        #expect(group.members.last?.subtitle == "Invited · pending")
        #expect(group.members.last?.role == .viewer)
        #expect(group.inviteEmail.isEmpty)
        #expect(group.lastEvent == "Invitation sent")
    }
}

@Suite("Itinerary actions")
@MainActor
struct ItineraryActionTests {
    private func plan() throws -> TripPlan {
        let destination = try DestinationMatch(
            id: "madeira", name: "Madeira", country: "Portugal", region: "Atlantic", photoReference: "madeira", score: MatchScore(percent: 94),
            estimate: TripEstimate(low: Money(amountMinor: 1, currency: "USD"), high: Money(amountMinor: 2, currency: "USD")),
            weather: WeatherSummary(condition: "Warm", temperatureCelsius: 23), flight: FlightSummary(stops: 1, approximateHours: 9), durationDays: 7, pace: "Easy", tags: [], reasons: []
        )
        let lunch = ItineraryActivity(time: "12:30", title: "Lunch", detail: "", status: .reserved)
        let walk = ItineraryActivity(time: "10:00", title: "Walk", detail: "")
        return try TripPlan(
            id: "t", destination: destination, dateRangeLabel: "", travelersLabel: "", startDate: Date(timeIntervalSince1970: 1_791_763_200), budgetCeiling: Money(amountMinor: 300000, currency: "USD"),
            readiness: [], budgetLines: [],
            days: [ItineraryDay(dayNumber: 1, title: "Day 1", summary: "", activities: [walk, lunch]), ItineraryDay(dayNumber: 2, title: "Day 2", summary: "", activities: [])]
        )
    }

    @Test("free activities remove immediately, reservations ask first")
    func remove() throws {
        let plan = try plan()
        let model = ItineraryModel(plan: { plan }, selectedDay: 1)
        model.requestRemove(plan.days[0].activities[0])
        #expect(model.day?.activities.map(\.title) == ["Lunch"])
        #expect(model.lastEvent == "Removed from Day 1")
        model.requestRemove(plan.days[0].activities[1])
        #expect(model.pendingRemoval?.title == "Lunch")
        #expect(model.day?.activities.count == 1)
        model.cancelRemoval()
        #expect(model.pendingRemoval == nil)
        model.requestRemove(plan.days[0].activities[1])
        model.remove(plan.days[0].activities[1])
        #expect(model.day?.activities.isEmpty == true)
    }

    @Test("moving an activity relocates it to the chosen day")
    func move() throws {
        let plan = try plan()
        let model = ItineraryModel(plan: { plan }, selectedDay: 1)
        #expect(model.dayOptions.map(\.id) == ["2"])
        model.move(plan.days[0].activities[0], toDay: 2)
        #expect(model.day?.activities.map(\.title) == ["Lunch"])
        model.selectedDay = 2
        #expect(model.day?.activities.map(\.title) == ["Walk"])
        #expect(model.lastEvent == "Moved to Day 2")
    }
}
