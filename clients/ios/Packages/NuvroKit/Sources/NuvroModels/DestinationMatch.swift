import Foundation

public struct WeatherSummary: Sendable, Hashable, Codable {
    public let condition: String
    public let temperatureCelsius: Int

    public init(condition: String, temperatureCelsius: Int) {
        self.condition = condition
        self.temperatureCelsius = temperatureCelsius
    }

    private enum CodingKeys: String, CodingKey {
        case condition
        case temperatureCelsius = "temperature_celsius"
    }
}

public struct FlightSummary: Sendable, Hashable, Codable {
    public let stops: Int
    public let approximateHours: Int

    public init(stops: Int, approximateHours: Int) {
        self.stops = stops
        self.approximateHours = approximateHours
    }

    private enum CodingKeys: String, CodingKey {
        case stops
        case approximateHours = "approximate_hours"
    }
}

public struct TripEstimate: Sendable, Hashable, Codable {
    public let low: Money
    public let high: Money

    public init(low: Money, high: Money) {
        self.low = low
        self.high = high
    }
}

public struct DestinationMatch: Sendable, Hashable, Codable, Identifiable {
    public let id: String
    public let name: String
    public let country: String
    public let region: String
    public let photoReference: String
    public let photoURL: URL?
    public let score: MatchScore
    public let estimate: TripEstimate?
    public let weather: WeatherSummary?
    public let flight: FlightSummary?
    public let durationDays: Int
    public let pace: String
    public let tags: [String]
    public let reasons: [MatchReason]
    public let summary: String?

    public init(id: String, name: String, country: String, region: String, photoReference: String, photoURL: URL? = nil, score: MatchScore, estimate: TripEstimate?, weather: WeatherSummary?, flight: FlightSummary?, durationDays: Int, pace: String, tags: [String], reasons: [MatchReason], summary: String? = nil) {
        self.id = id
        self.name = name
        self.country = country
        self.region = region
        self.photoReference = photoReference
        self.photoURL = photoURL
        self.score = score
        self.estimate = estimate
        self.weather = weather
        self.flight = flight
        self.durationDays = durationDays
        self.pace = pace
        self.tags = tags
        self.reasons = reasons
        self.summary = summary
    }

    private enum CodingKeys: String, CodingKey {
        case id, name, country, region, score, estimate, weather, flight, pace, tags, reasons, summary
        case photoReference = "photo_reference"
        case photoURL = "photo_url"
        case durationDays = "duration_days"
    }

    public var displayName: String { "\(name), \(country)" }
}
