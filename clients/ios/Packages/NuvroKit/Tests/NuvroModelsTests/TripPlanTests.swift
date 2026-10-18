import Foundation
import Testing

@testable import NuvroModels

@Suite("TripPlan")
struct TripPlanTests {
    private func money(_ minor: Int64) throws -> Money { try Money(amountMinor: minor, currency: "USD") }

    private func destination() throws -> DestinationMatch {
        try DestinationMatch(
            id: "madeira", name: "Madeira", country: "Portugal", region: "Atlantic island", photoReference: "madeira",
            score: MatchScore(percent: 94),
            estimate: TripEstimate(low: money(264000), high: money(294000)),
            weather: WeatherSummary(condition: "Warm", temperatureCelsius: 23),
            flight: FlightSummary(stops: 1, approximateHours: 9),
            durationDays: 7, pace: "Easygoing", tags: ["Coastal"], reasons: []
        )
    }

    private func plan() throws -> TripPlan {
        try TripPlan(
            id: "trip-1", destination: destination(), dateRangeLabel: "Oct 12–19", travelersLabel: "2 adults", startDate: Date(timeIntervalSince1970: 1_791_763_200),
            budgetCeiling: money(300000),
            readiness: [
                ReadinessItem(name: "Destination", isReady: true), ReadinessItem(name: "Dates", isReady: true),
                ReadinessItem(name: "Stay", isReady: true), ReadinessItem(name: "Basic itinerary", isReady: true),
                ReadinessItem(name: "Budget", isReady: true), ReadinessItem(name: "Flight", isReady: false),
                ReadinessItem(name: "Activities", isReady: false), ReadinessItem(name: "Stay connected", isReady: false)
            ],
            budgetLines: [
                BudgetLine(category: "Flight", amount: money(64000)), BudgetLine(category: "Stay", amount: money(119000)),
                BudgetLine(category: "Food", amount: money(42000)), BudgetLine(category: "Activities", amount: money(33000)),
                BudgetLine(category: "Transport", amount: money(20000))
            ],
            days: [ItineraryDay(dayNumber: 1, title: "Day 1 — Arrival", summary: "Arrive · Funchal · Easy evening")]
        )
    }

    @Test("readiness is counted from the items, never hardcoded")
    func readiness() throws {
        let plan = try plan()

        #expect(plan.readyCount == 5)
        #expect(abs(plan.readinessFraction - 0.625) < 0.0001)
    }

    @Test("the estimated total is the sum of the budget lines in minor units")
    func estimatedTotal() throws {
        let plan = try plan()

        #expect(plan.estimatedTotal?.amountMinor == 278000)
        #expect(plan.estimatedTotal?.currency == "USD")
    }

    @Test("remaining budget is the ceiling minus the estimate")
    func remainingBudget() throws {
        #expect(try plan().remainingBudget?.amountMinor == 22000)
    }

    @Test("an empty plan has no total and zero readiness")
    func emptyPlan() throws {
        let plan = try TripPlan(id: "t", destination: destination(), dateRangeLabel: "", travelersLabel: "", startDate: Date(timeIntervalSince1970: 1_791_763_200), budgetCeiling: money(1), readiness: [], budgetLines: [], days: [])

        #expect(plan.estimatedTotal == nil)
        #expect(plan.readinessFraction == 0)
    }

    @Test("round trips through the contract keys")
    func roundTrips() throws {
        let plan = try plan()
        let encoded = try JSONEncoder().encode(plan)
        let text = String(decoding: encoded, as: UTF8.self)

        #expect(text.contains("budget_ceiling"))
        #expect(text.contains("is_ready"))
        #expect(text.contains("day_number"))
        #expect(text.contains("start_date"))
        #expect(try JSONDecoder().decode(TripPlan.self, from: encoded) == plan)
    }

    @Test("the budget fraction compares the estimate to the ceiling")
    func budgetFraction() throws {
        #expect(abs(try plan().budgetFraction - 2780.0 / 3000.0) < 0.0001)
    }

    @Test("day dates count forward from the start date")
    func dayDates() throws {
        let plan = try plan()
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: "UTC")!

        #expect(calendar.component(.day, from: plan.date(forDay: 1, calendar: calendar)) == 12)
        #expect(calendar.component(.day, from: plan.date(forDay: 3, calendar: calendar)) == 14)
    }

    @Test("a plan adjustment swaps one activity and leaves the rest of the day intact")
    func adjustment() throws {
        let garden = ItineraryActivity(time: "15:00", title: "Botanical Garden", detail: "Outdoor", isOutdoor: true)
        let museum = ItineraryActivity(time: "15:00", title: "Madeira Story Centre", detail: "Indoor")
        let dinner = ItineraryActivity(time: "19:30", title: "Dinner", detail: "Reserved", status: .reserved)
        let day = ItineraryDay(dayNumber: 3, title: "Day 3", summary: "", activities: [garden, dinner])
        let adjustment = PlanAdjustment(trigger: "Rain", dayNumber: 3, before: garden, after: museum, relocationNote: "", reasoning: "", preserved: [])

        let adjusted = adjustment.apply(to: day)

        #expect(adjusted.activities.map(\.title) == ["Madeira Story Centre", "Dinner"])
        let original = try plan()
        let withDay = TripPlan(id: original.id, destination: original.destination, dateRangeLabel: original.dateRangeLabel, travelersLabel: original.travelersLabel, startDate: original.startDate, budgetCeiling: original.budgetCeiling, readiness: original.readiness, budgetLines: original.budgetLines, days: [day])
        #expect(withDay.replacing(day: adjusted).day(3)?.activities.first?.title == "Madeira Story Centre")
        #expect(withDay.replacing(day: adjusted).days.count == 1)
    }

    @Test("a stay total multiplies the nightly rate by the nights")
    func stayTotal() throws {
        let stay = try StayOption(id: "q", name: "Quinta", area: "Funchal", distanceNote: "", nightlyRate: money(14000), nights: 7, rating: 4.7, minutesToPlans: 9, recommendation: .bestOverall, photoReference: "gold", reasons: [], coordinate: GeoCoordinate(latitude: 0, longitude: 0))

        #expect(stay.stayTotal?.amountMinor == 98000)
    }
}

@Suite("Companion models")
struct CompanionModelContractTests {
    @Test("trip documents mask all but the last four characters")
    func masking() {
        let document = TripDocument(id: "x", category: .flight, title: "", subtitle: "", whenLabel: "", travelers: 1, isAvailableOffline: false, reference: "TP1691-8842")
        #expect(document.maskedReference == "••••••-••••")
    }

    @Test("free time fits a place when the walk there and back plus a visit fit")
    func freeTime() {
        let window = FreeTimeWindow(minutes: 90, untilLabel: "dinner")
        let media = DestinationMedia(placeholder: "gold", altText: "")
        let near = NearbyPlace(id: "a", name: "", category: .food, categoryLabel: "", walkMinutes: 10, travelLabel: "", isOpen: true, hoursLabel: "", reason: "", media: media, coordinate: GeoCoordinate(latitude: 0, longitude: 0), markerLetter: "A")
        let far = NearbyPlace(id: "b", name: "", category: .food, categoryLabel: "", walkMinutes: 20, travelLabel: "", isOpen: true, hoursLabel: "", reason: "", media: media, coordinate: GeoCoordinate(latitude: 0, longitude: 0), markerLetter: "B")

        #expect(window.fits(near))
        #expect(!window.fits(far))
    }

    @Test("destination media round trips through contract keys")
    func mediaKeys() throws {
        let media = DestinationMedia(url: URL(string: "https://example.test/photo.jpg"), placeholder: "gold", altText: "Market", attribution: "Photo", provider: "places")
        let text = String(decoding: try JSONEncoder().encode(media), as: UTF8.self)

        #expect(text.contains("alt_text"))
        #expect(media.hasPhoto)
        #expect(try JSONDecoder().decode(DestinationMedia.self, from: Data(text.utf8)) == media)
    }
}
