import Foundation

public struct PlaceMemory: Sendable, Hashable, Codable, Identifiable {
    public let name: String
    public let media: DestinationMedia

    public init(name: String, media: DestinationMedia) {
        self.name = name
        self.media = media
    }

    public var id: String { name }
}

public struct NextTripSuggestion: Sendable, Hashable, Codable {
    public let name: String
    public let country: String
    public let reason: String
    public let media: DestinationMedia

    public init(name: String, country: String, reason: String, media: DestinationMedia) {
        self.name = name
        self.country = country
        self.reason = reason
        self.media = media
    }
}

public struct TripSummary: Sendable, Hashable, Codable {
    public let destinationName: String
    public let dateRangeLabel: String
    public let days: Int
    public let placesCount: Int
    public let activitiesCount: Int
    public let kilometresExplored: Int
    public let hero: DestinationMedia
    public let places: [PlaceMemory]
    public let total: Money
    public let plannedTotal: Money
    public let lovedOptions: [String]
    public let notForMeOptions: [String]
    public let suggestion: NextTripSuggestion

    public init(destinationName: String, dateRangeLabel: String, days: Int, placesCount: Int, activitiesCount: Int, kilometresExplored: Int, hero: DestinationMedia, places: [PlaceMemory], total: Money, plannedTotal: Money, lovedOptions: [String], notForMeOptions: [String], suggestion: NextTripSuggestion) {
        self.destinationName = destinationName
        self.dateRangeLabel = dateRangeLabel
        self.days = days
        self.placesCount = placesCount
        self.activitiesCount = activitiesCount
        self.kilometresExplored = kilometresExplored
        self.hero = hero
        self.places = places
        self.total = total
        self.plannedTotal = plannedTotal
        self.lovedOptions = lovedOptions
        self.notForMeOptions = notForMeOptions
        self.suggestion = suggestion
    }

    public var budgetDifference: Money? {
        try? Money(amountMinor: plannedTotal.amountMinor - total.amountMinor, currency: total.currency)
    }

    private enum CodingKeys: String, CodingKey {
        case days, hero, places, total, suggestion
        case destinationName = "destination_name"
        case dateRangeLabel = "date_range_label"
        case placesCount = "places_count"
        case activitiesCount = "activities_count"
        case kilometresExplored = "kilometres_explored"
        case plannedTotal = "planned_total"
        case lovedOptions = "loved_options"
        case notForMeOptions = "not_for_me_options"
    }
}
