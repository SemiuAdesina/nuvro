import Foundation

public struct ItineraryActivity: Sendable, Hashable, Codable, Identifiable {
    public enum Status: String, Sendable, Codable {
        case booked
        case reserved
        case ticket
        case weather
    }

    public let time: String
    public let title: String
    public let detail: String
    public let photoReference: String?
    public let status: Status?
    public let travelToNext: String?
    public let isOutdoor: Bool

    public init(time: String, title: String, detail: String, photoReference: String? = nil, status: Status? = nil, travelToNext: String? = nil, isOutdoor: Bool = false) {
        self.time = time
        self.title = title
        self.detail = detail
        self.photoReference = photoReference
        self.status = status
        self.travelToNext = travelToNext
        self.isOutdoor = isOutdoor
    }

    public var id: String { "\(time) \(title)" }

    private enum CodingKeys: String, CodingKey {
        case time, title, detail, status
        case photoReference = "photo_reference"
        case travelToNext = "travel_to_next"
        case isOutdoor = "is_outdoor"
    }
}

public struct ItineraryDay: Sendable, Hashable, Codable, Identifiable {
    public let dayNumber: Int
    public let title: String
    public let summary: String
    public let departureTravel: String?
    public let activities: [ItineraryActivity]

    public init(dayNumber: Int, title: String, summary: String, departureTravel: String? = nil, activities: [ItineraryActivity] = []) {
        self.dayNumber = dayNumber
        self.title = title
        self.summary = summary
        self.departureTravel = departureTravel
        self.activities = activities
    }

    public var id: Int { dayNumber }

    private enum CodingKeys: String, CodingKey {
        case dayNumber = "day_number"
        case departureTravel = "departure_travel"
        case title, summary, activities
    }

    public func travel(to activity: ItineraryActivity) -> String? {
        guard let index = activities.firstIndex(where: { $0.id == activity.id }) else { return nil }
        return index == 0 ? departureTravel : activities[index - 1].travelToNext
    }
}

public struct PlanAdjustment: Sendable, Hashable, Codable {
    public let trigger: String
    public let dayNumber: Int
    public let before: ItineraryActivity
    public let after: ItineraryActivity
    public let relocationNote: String
    public let reasoning: String
    public let preserved: [String]

    public init(trigger: String, dayNumber: Int, before: ItineraryActivity, after: ItineraryActivity, relocationNote: String, reasoning: String, preserved: [String]) {
        self.trigger = trigger
        self.dayNumber = dayNumber
        self.before = before
        self.after = after
        self.relocationNote = relocationNote
        self.reasoning = reasoning
        self.preserved = preserved
    }

    private enum CodingKeys: String, CodingKey {
        case trigger, before, after, reasoning, preserved
        case dayNumber = "day_number"
        case relocationNote = "relocation_note"
    }

    public func apply(to day: ItineraryDay) -> ItineraryDay {
        let activities = day.activities.map { $0.id == before.id ? after : $0 }
        return ItineraryDay(dayNumber: day.dayNumber, title: day.title, summary: day.summary, departureTravel: day.departureTravel, activities: activities)
    }
}
