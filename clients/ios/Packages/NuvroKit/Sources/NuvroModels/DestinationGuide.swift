import Foundation

public struct GeoCoordinate: Sendable, Hashable, Codable {
    public let latitude: Double
    public let longitude: Double

    public init(latitude: Double, longitude: Double) {
        self.latitude = latitude
        self.longitude = longitude
    }
}

public struct StayArea: Sendable, Hashable, Codable, Identifiable {
    public let name: String
    public let summary: String
    public let highlight: String
    public let photoReference: String

    public init(name: String, summary: String, highlight: String, photoReference: String) {
        self.name = name
        self.summary = summary
        self.highlight = highlight
        self.photoReference = photoReference
    }

    public var id: String { name }

    private enum CodingKeys: String, CodingKey {
        case name, summary, highlight
        case photoReference = "photo_reference"
    }
}

public struct Activity: Sendable, Hashable, Codable, Identifiable {
    public let name: String
    public let summary: String
    public let photoReference: String

    public init(name: String, summary: String, photoReference: String) {
        self.name = name
        self.summary = summary
        self.photoReference = photoReference
    }

    public var id: String { name }

    private enum CodingKeys: String, CodingKey {
        case name, summary
        case photoReference = "photo_reference"
    }
}

public struct MapPlace: Sendable, Hashable, Codable, Identifiable {
    public enum Kind: String, Sendable, Codable {
        case stay
        case attraction
        case restaurant
        case saved
    }

    public let name: String
    public let kind: Kind
    public let coordinate: GeoCoordinate

    public init(name: String, kind: Kind, coordinate: GeoCoordinate) {
        self.name = name
        self.kind = kind
        self.coordinate = coordinate
    }

    public var id: String { name }
}

public struct DestinationGuide: Sendable, Hashable, Codable {
    public let destinationID: String
    public let centre: GeoCoordinate
    public let areas: [StayArea]
    public let activities: [Activity]
    public let places: [MapPlace]

    public init(destinationID: String, centre: GeoCoordinate, areas: [StayArea], activities: [Activity], places: [MapPlace]) {
        self.destinationID = destinationID
        self.centre = centre
        self.areas = areas
        self.activities = activities
        self.places = places
    }

    private enum CodingKeys: String, CodingKey {
        case centre, areas, activities, places
        case destinationID = "destination_id"
    }
}
