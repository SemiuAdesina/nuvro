import Foundation
import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import Observation

@Observable
@MainActor
public final class ItineraryModel {
    public var selectedDay: Int
    public var pendingRemoval: ItineraryActivity?
    public var moving: ItineraryActivity?
    public private(set) var removedIDs: Set<String> = []
    public private(set) var moves: [String: Int] = [:]
    public private(set) var lastEvent: String?

    public func clearEvent() {
        lastEvent = nil
    }
    public let copy: ItineraryCopy
    private let calendar: Calendar
    private let plan: () -> TripPlan

    public init(plan: @escaping @MainActor () -> TripPlan, selectedDay: Int = 2, language: Language = .english, calendar: Calendar = .current) {
        self.plan = plan
        self.selectedDay = selectedDay
        self.copy = .resolve(language)
        self.calendar = calendar
    }

    public var currentPlan: TripPlan { plan() }

    public var days: [(number: Int, label: String, date: String)] {
        currentPlan.days.map { day in
            let date = currentPlan.date(forDay: day.dayNumber, calendar: calendar)
            return (day.dayNumber, "\(copy.day) \(day.dayNumber)", "\(calendar.component(.day, from: date))")
        }
    }

    public var day: ItineraryDay? {
        guard let stored = currentPlan.day(selectedDay) else { return nil }
        let kept = stored.activities.filter { !removedIDs.contains($0.id) && (moves[$0.id] ?? selectedDay) == selectedDay }
        let movedIn = currentPlan.days.filter { $0.dayNumber != selectedDay }.flatMap(\.activities).filter { moves[$0.id] == selectedDay && !removedIDs.contains($0.id) }
        let activities = (kept + movedIn).sorted { $0.time < $1.time }
        return ItineraryDay(dayNumber: stored.dayNumber, title: stored.title, summary: stored.summary, activities: activities)
    }

    public func requestRemove(_ activity: ItineraryActivity) {
        if activity.status == nil {
            remove(activity)
        } else {
            pendingRemoval = activity
        }
    }

    public func remove(_ activity: ItineraryActivity) {
        removedIDs.insert(activity.id)
        pendingRemoval = nil
        lastEvent = "\(copy.removed) \(copy.day) \(selectedDay)"
    }

    public func cancelRemoval() {
        pendingRemoval = nil
    }

    public func move(_ activity: ItineraryActivity, toDay dayNumber: Int) {
        guard currentPlan.day(dayNumber) != nil else { return }
        moves[activity.id] = dayNumber
        moving = nil
        lastEvent = "\(copy.movedTo) \(copy.day) \(dayNumber)"
    }

    public var dayOptions: [PickerOption] {
        currentPlan.days.filter { $0.dayNumber != selectedDay }.map { PickerOption(id: "\($0.dayNumber)", label: $0.title, detail: $0.summary) }
    }

    public var daySubtitle: String {
        guard let day else { return "" }
        let outdoor = day.activities.filter(\.isOutdoor).count
        let stops = "\(day.activities.count) stops"
        return outdoor > 0 ? "\(stops) · \(outdoor) outdoor" : stops
    }

    public func statusLabel(_ status: ItineraryActivity.Status) -> String {
        switch status {
        case .booked: copy.booked
        case .reserved: copy.reserved
        case .ticket: copy.ticket
        case .weather: copy.weather
        }
    }
}
