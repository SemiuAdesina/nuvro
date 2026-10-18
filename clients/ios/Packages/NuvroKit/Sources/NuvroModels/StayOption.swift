import Foundation

public struct StayOption: Sendable, Hashable, Codable, Identifiable {
    public enum Recommendation: String, Sendable, Codable, CaseIterable {
        case bestOverall = "best_overall"
        case bestValue = "best_value"
        case bestLocation = "best_location"
        case bestPremium = "best_premium"
    }

    public let id: String
    public let name: String
    public let area: String
    public let distanceNote: String
    public let nightlyRate: Money
    public let nights: Int
    public let rating: Double
    public let minutesToPlans: Int
    public let recommendation: Recommendation
    public let photoReference: String
    public let reasons: [String]
    public let coordinate: GeoCoordinate

    public init(id: String, name: String, area: String, distanceNote: String, nightlyRate: Money, nights: Int, rating: Double, minutesToPlans: Int, recommendation: Recommendation, photoReference: String, reasons: [String], coordinate: GeoCoordinate) {
        self.id = id
        self.name = name
        self.area = area
        self.distanceNote = distanceNote
        self.nightlyRate = nightlyRate
        self.nights = nights
        self.rating = rating
        self.minutesToPlans = minutesToPlans
        self.recommendation = recommendation
        self.photoReference = photoReference
        self.reasons = reasons
        self.coordinate = coordinate
    }

    private enum CodingKeys: String, CodingKey {
        case id, name, area, nights, rating, recommendation, reasons, coordinate
        case distanceNote = "distance_note"
        case nightlyRate = "nightly_rate"
        case minutesToPlans = "minutes_to_plans"
        case photoReference = "photo_reference"
    }

    public var stayTotal: Money? {
        try? Money(amountMinor: nightlyRate.amountMinor * Int64(nights), currency: nightlyRate.currency)
    }
}
