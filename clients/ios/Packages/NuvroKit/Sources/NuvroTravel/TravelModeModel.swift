import Foundation
import NuvroLocalization
import NuvroModels
import Observation

@Observable
@MainActor
public final class TravelModeModel {
    public private(set) var day: ItineraryDay
    public private(set) var appliedAdjustment: PlanAdjustment?
    public private(set) var isApplying = false
    public var weatherFreshness: Freshness = .fresh
    public let destinationName: String
    public let dayNumber: Int
    public let now: Date
    public let weather: WeatherSummary
    public let pendingAdjustment: PlanAdjustment
    public let copy: TravelCopy
    public let replanCopy: ReplanCopy
    private let calendar: Calendar

    public init(day: ItineraryDay, dayNumber: Int, destinationName: String, now: Date, weather: WeatherSummary, pendingAdjustment: PlanAdjustment, language: Language = .english, calendar: Calendar = .current) {
        self.day = day
        self.dayNumber = dayNumber
        self.destinationName = destinationName
        self.now = now
        self.weather = weather
        self.pendingAdjustment = pendingAdjustment
        self.copy = .resolve(language)
        self.replanCopy = .resolve(language)
        self.calendar = calendar
    }

    public var greeting: String {
        switch calendar.component(.hour, from: now) {
        case ..<12: copy.goodMorning
        case 12..<18: copy.goodAfternoon
        default: copy.goodEvening
        }
    }

    public var dateLabel: String {
        now.formatted(.dateTime.weekday(.wide).month(.abbreviated).day().locale(Locale(identifier: "en_US")))
    }

    public var nextActivity: ItineraryActivity? {
        day.activities.first { minutes(of: $0.time) >= minutesNow }
    }

    public var leaveInMinutes: Int? {
        guard let next = nextActivity else { return nil }
        return max(0, minutes(of: next.time) - walkMinutes(next) - minutesNow)
    }

    public var hasPendingWeatherChange: Bool {
        appliedAdjustment == nil
    }

    public func isNext(_ activity: ItineraryActivity) -> Bool {
        activity.id == nextActivity?.id
    }

    public func wasUpdated(_ activity: ItineraryActivity) -> Bool {
        appliedAdjustment?.after.id == activity.id
    }

    public func beginApplying() {
        isApplying = true
    }

    public func applyPendingAdjustment() {
        day = pendingAdjustment.apply(to: day)
        appliedAdjustment = pendingAdjustment
        isApplying = false
    }

    public func keepOriginal() {
        isApplying = false
    }

    private var minutesNow: Int {
        calendar.component(.hour, from: now) * 60 + calendar.component(.minute, from: now)
    }

    private func minutes(of time: String) -> Int {
        let parts = time.split(separator: ":").compactMap { Int($0) }
        guard parts.count == 2 else { return 0 }
        return parts[0] * 60 + parts[1]
    }

    private func walkMinutes(_ activity: ItineraryActivity) -> Int {
        guard let travel = day.travel(to: activity) else { return 0 }
        let parts = travel.split(separator: " ")
        guard let unitIndex = parts.firstIndex(where: { $0.hasPrefix("min") }), unitIndex > 0 else { return 0 }
        return Int(parts[unitIndex - 1]) ?? 0
    }
}
