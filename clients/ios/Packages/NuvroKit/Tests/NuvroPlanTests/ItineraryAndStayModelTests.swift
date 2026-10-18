import Foundation
import Testing

@testable import NuvroModels
@testable import NuvroPlan

@Suite("ItineraryModel and StayModel")
@MainActor
struct ItineraryAndStayModelTests {
    private func money(_ minor: Int64) throws -> Money { try Money(amountMinor: minor, currency: "USD") }

    private var calendar: Calendar {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: "UTC")!
        return calendar
    }

    private func plan() throws -> TripPlan {
        let destination = try DestinationMatch(
            id: "madeira", name: "Madeira", country: "Portugal", region: "Atlantic island", photoReference: "madeira",
            score: MatchScore(percent: 94), estimate: TripEstimate(low: money(264000), high: money(294000)),
            weather: WeatherSummary(condition: "Warm", temperatureCelsius: 23), flight: FlightSummary(stops: 1, approximateHours: 9),
            durationDays: 7, pace: "Easygoing", tags: ["Coastal"], reasons: []
        )
        return try TripPlan(
            id: "t", destination: destination, dateRangeLabel: "Oct 12–19", travelersLabel: "2 adults", startDate: Date(timeIntervalSince1970: 1_791_763_200),
            budgetCeiling: money(300000), readiness: [], budgetLines: [],
            days: [
                ItineraryDay(dayNumber: 1, title: "Day 1 — Arrival", summary: "", activities: [ItineraryActivity(time: "16:40", title: "Land", detail: "")]),
                ItineraryDay(dayNumber: 2, title: "Day 2 — Funchal", summary: "", activities: [
                    ItineraryActivity(time: "10:00", title: "Old Town", detail: "", isOutdoor: true),
                    ItineraryActivity(time: "12:30", title: "Market", detail: ""),
                    ItineraryActivity(time: "19:30", title: "Dinner", detail: "", status: .reserved)
                ])
            ]
        )
    }

    @Test("day tabs are labelled from the plan's start date")
    func dayTabs() throws {
        let plan = try plan()
        let model = ItineraryModel(plan: { plan }, calendar: calendar)

        #expect(model.days.map(\.label) == ["Day 1", "Day 2"])
        #expect(model.days.map(\.date) == ["12", "13"])
    }

    @Test("the selected day drives the title and subtitle")
    func selectedDay() throws {
        let plan = try plan()
        let model = ItineraryModel(plan: { plan }, calendar: calendar)

        #expect(model.day?.title == "Day 2 — Funchal")
        #expect(model.daySubtitle == "3 stops · 1 outdoor")
        model.selectedDay = 1
        #expect(model.day?.title == "Day 1 — Arrival")
        #expect(model.daySubtitle == "1 stops")
    }

    @Test("status labels come from copy")
    func statusLabels() throws {
        let plan = try plan()
        let model = ItineraryModel(plan: { plan })

        #expect(model.statusLabel(.reserved) == "Reserved")
        #expect(model.statusLabel(.ticket) == "Ticket")
    }

    private func stays() throws -> [StayOption] {
        [
            StayOption(id: "q", name: "Quinta", area: "Funchal", distanceNote: "", nightlyRate: try money(14000), nights: 7, rating: 4.7, minutesToPlans: 9, recommendation: .bestOverall, photoReference: "gold", reasons: [], coordinate: GeoCoordinate(latitude: 0, longitude: 0)),
            StayOption(id: "c", name: "Carmo", area: "Funchal", distanceNote: "", nightlyRate: try money(9600), nights: 7, rating: 4.3, minutesToPlans: 6, recommendation: .bestValue, photoReference: "gold", reasons: [], coordinate: GeoCoordinate(latitude: 0, longitude: 0))
        ]
    }

    @Test("the hero follows the active recommendation filter and the others exclude it")
    func heroAndOthers() throws {
        let model = StayModel(options: try stays(), locale: Locale(identifier: "en_US"))

        #expect(model.hero?.name == "Quinta")
        #expect(model.others.map(\.name) == ["Carmo"])
        model.filter = .bestValue
        #expect(model.hero?.name == "Carmo")
        #expect(model.others.map(\.name) == ["Quinta"])
    }

    @Test("stay totals and nightly labels are money formatted")
    func labels() throws {
        let model = StayModel(options: try stays(), locale: Locale(identifier: "en_US"))

        #expect(model.nightlyLabel(model.options[0]) == "$140")
        #expect(model.totalLabel(model.options[0]) == "$980")
    }

    @Test("saving a stay is reversible")
    func saving() throws {
        let model = StayModel(options: try stays())
        model.toggleSaved(model.options[0])
        #expect(model.isSaved(model.options[0]))
        model.toggleSaved(model.options[0])
        #expect(!model.isSaved(model.options[0]))
    }
}
