import Foundation

public struct SavedDestination: Sendable, Hashable, Codable, Identifiable {
    public let id: String
    public let name: String
    public let country: String
    public let matchPercent: Int
    public let reason: String
    public let note: String
    public let media: DestinationMedia

    public init(id: String, name: String, country: String, matchPercent: Int, reason: String, note: String, media: DestinationMedia) {
        self.id = id
        self.name = name
        self.country = country
        self.matchPercent = matchPercent
        self.reason = reason
        self.note = note
        self.media = media
    }

    private enum CodingKeys: String, CodingKey {
        case id, name, country, reason, note, media
        case matchPercent = "match_percent"
    }
}

public struct SavedPlace: Sendable, Hashable, Codable, Identifiable {
    public let name: String
    public let subtitle: String
    public let media: DestinationMedia

    public init(name: String, subtitle: String, media: DestinationMedia) {
        self.name = name
        self.subtitle = subtitle
        self.media = media
    }

    public var id: String { name }
}

public struct ComparisonColumn: Sendable, Hashable, Codable, Identifiable {
    public let name: String
    public let matchPercent: Int
    public let gist: String
    public let media: DestinationMedia

    public init(name: String, matchPercent: Int, gist: String, media: DestinationMedia) {
        self.name = name
        self.matchPercent = matchPercent
        self.gist = gist
        self.media = media
    }

    public var id: String { name }

    private enum CodingKeys: String, CodingKey {
        case name, gist, media
        case matchPercent = "match_percent"
    }
}

public struct ComparisonMetric: Sendable, Hashable, Codable, Identifiable {
    public enum Winner: String, Sendable, Codable {
        case left, right, none
    }

    public let key: String
    public let left: String
    public let right: String
    public let winner: Winner

    public init(key: String, left: String, right: String, winner: Winner) {
        self.key = key
        self.left = left
        self.right = right
        self.winner = winner
    }

    public var id: String { key }
}

public struct ComparisonDifference: Sendable, Hashable, Codable, Identifiable {
    public let name: String
    public let text: String

    public init(name: String, text: String) {
        self.name = name
        self.text = text
    }

    public var id: String { name }
}

public struct DestinationComparison: Sendable, Hashable, Codable {
    public let subtitle: String
    public let columns: [ComparisonColumn]
    public let metrics: [ComparisonMetric]
    public let differences: [ComparisonDifference]

    public init(subtitle: String, columns: [ComparisonColumn], metrics: [ComparisonMetric], differences: [ComparisonDifference]) {
        self.subtitle = subtitle
        self.columns = columns
        self.metrics = metrics
        self.differences = differences
    }
}
