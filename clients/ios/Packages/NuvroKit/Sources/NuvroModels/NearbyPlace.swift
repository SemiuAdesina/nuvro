import Foundation

public struct NearbyPlace: Sendable, Hashable, Codable, Identifiable {
    public enum Category: String, Sendable, Codable, CaseIterable {
        case food, coffee, thingsToDo = "things_to_do", views, shopping, indoor, saved
    }

    public let id: String
    public let name: String
    public let category: Category
    public let categoryLabel: String
    public let walkMinutes: Int
    public let travelLabel: String
    public let isOpen: Bool
    public let hoursLabel: String
    public let reason: String
    public let media: DestinationMedia
    public let coordinate: GeoCoordinate
    public let markerLetter: String

    public init(id: String, name: String, category: Category, categoryLabel: String, walkMinutes: Int, travelLabel: String, isOpen: Bool, hoursLabel: String, reason: String, media: DestinationMedia, coordinate: GeoCoordinate, markerLetter: String) {
        self.id = id
        self.name = name
        self.category = category
        self.categoryLabel = categoryLabel
        self.walkMinutes = walkMinutes
        self.travelLabel = travelLabel
        self.isOpen = isOpen
        self.hoursLabel = hoursLabel
        self.reason = reason
        self.media = media
        self.coordinate = coordinate
        self.markerLetter = markerLetter
    }

    private enum CodingKeys: String, CodingKey {
        case id, name, category, reason, media, coordinate
        case categoryLabel = "category_label"
        case walkMinutes = "walk_minutes"
        case travelLabel = "travel_label"
        case isOpen = "is_open"
        case hoursLabel = "hours_label"
        case markerLetter = "marker_letter"
    }
}

public struct FreeTimeWindow: Sendable, Hashable, Codable {
    public let minutes: Int
    public let untilLabel: String

    public init(minutes: Int, untilLabel: String) {
        self.minutes = minutes
        self.untilLabel = untilLabel
    }

    public func fits(_ place: NearbyPlace, visitMinutes: Int = 60) -> Bool {
        place.walkMinutes * 2 + visitMinutes <= minutes
    }

    private enum CodingKeys: String, CodingKey {
        case minutes
        case untilLabel = "until_label"
    }
}
