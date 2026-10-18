import Foundation
import Testing

@testable import NuvroModels
@testable import NuvroTravel

@Suite("TravelModeModel")
@MainActor
struct TravelModeModelTests {
    private var calendar: Calendar {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: "UTC")!
        return calendar
    }

    private func date(hour: Int, minute: Int) -> Date {
        var components = DateComponents()
        components.year = 2026; components.month = 10; components.day = 14; components.hour = hour; components.minute = minute
        return calendar.date(from: components)!
    }

    private let garden = ItineraryActivity(time: "15:00", title: "Botanical Garden", detail: "Outdoor · 25 min by car", status: .weather, travelToNext: "25 min by car", isOutdoor: true)
    private let museum = ItineraryActivity(time: "15:00", title: "Madeira Story Centre", detail: "Indoor · 8 min walk", travelToNext: "8 min walk")

    private func day() -> ItineraryDay {
        ItineraryDay(dayNumber: 3, title: "Day 3", summary: "", departureTravel: "12 min walk", activities: [
            ItineraryActivity(time: "10:00", title: "Funchal Cable Car", detail: "To Monte · 15 min ride", status: .ticket, travelToNext: "12 min walk"),
            ItineraryActivity(time: "12:30", title: "Lunch at Alto Monte", detail: "Reservation for 2", status: .reserved, travelToNext: "10 min walk"),
            garden,
            ItineraryActivity(time: "19:30", title: "Dinner in Old Town", detail: "Table for 2", status: .reserved)
        ])
    }

    private func model(at time: Date) -> TravelModeModel {
        TravelModeModel(
            day: day(), dayNumber: 3, destinationName: "Madeira", now: time, weather: WeatherSummary(condition: "Light clouds", temperatureCelsius: 23),
            pendingAdjustment: PlanAdjustment(trigger: "Rain", dayNumber: 3, before: garden, after: museum, relocationNote: "moved", reasoning: "", preserved: ["Dinner"]),
            calendar: calendar
        )
    }

    @Test("the next activity is the first one not yet started")
    func nextActivity() {
        #expect(model(at: date(hour: 9, minute: 26)).nextActivity?.title == "Funchal Cable Car")
        #expect(model(at: date(hour: 11, minute: 0)).nextActivity?.title == "Lunch at Alto Monte")
        #expect(model(at: date(hour: 21, minute: 0)).nextActivity == nil)
    }

    @Test("leave-in time subtracts the walk from the gap to the next activity")
    func leaveIn() {
        #expect(model(at: date(hour: 9, minute: 26)).leaveInMinutes == 22)
        #expect(model(at: date(hour: 9, minute: 55)).leaveInMinutes == 0)
        #expect(model(at: date(hour: 11, minute: 0)).leaveInMinutes == 78)
    }

    @Test("ride durations in the activity detail never count as the walk")
    func detailIsNotTravel() {
        let travel = TravelModeModel(
            day: ItineraryDay(dayNumber: 1, title: "", summary: "", activities: [ItineraryActivity(time: "10:00", title: "Cable Car", detail: "15 min ride")]),
            dayNumber: 1, destinationName: "Madeira", now: date(hour: 9, minute: 26), weather: WeatherSummary(condition: "", temperatureCelsius: 20),
            pendingAdjustment: PlanAdjustment(trigger: "", dayNumber: 1, before: garden, after: museum, relocationNote: "", reasoning: "", preserved: []), calendar: calendar
        )

        #expect(travel.leaveInMinutes == 34)
    }

    @Test("greeting and date label follow the prototype clock")
    func greetingAndDate() {
        let morning = model(at: date(hour: 9, minute: 26))

        #expect(morning.greeting == "Good morning")
        #expect(morning.dateLabel.contains("Wednesday"))
        #expect(morning.dateLabel.contains("Oct 14"))
        #expect(model(at: date(hour: 15, minute: 0)).greeting == "Good afternoon")
    }

    @Test("applying the pending adjustment swaps the activity and clears the weather prompt")
    func apply() {
        let travel = model(at: date(hour: 9, minute: 26))
        #expect(travel.hasPendingWeatherChange)

        travel.beginApplying()
        #expect(travel.isApplying)
        travel.applyPendingAdjustment()

        #expect(!travel.isApplying)
        #expect(!travel.hasPendingWeatherChange)
        #expect(travel.day.activities.map(\.title) == ["Funchal Cable Car", "Lunch at Alto Monte", "Madeira Story Centre", "Dinner in Old Town"])
        #expect(travel.wasUpdated(museum))
    }

    @Test("keeping the original leaves the day untouched")
    func keepOriginal() {
        let travel = model(at: date(hour: 9, minute: 26))
        travel.beginApplying()
        travel.keepOriginal()

        #expect(!travel.isApplying)
        #expect(travel.day.activities.contains(garden))
        #expect(travel.hasPendingWeatherChange)
    }
}
