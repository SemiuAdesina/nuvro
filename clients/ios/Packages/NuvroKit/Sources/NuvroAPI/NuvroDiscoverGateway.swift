import Foundation
import NuvroModels

public struct DiscoverBrief: Sendable, Hashable {
    public let budgetMinor: Int64
    public let currency: String
    public let startDate: Date
    public let endDate: Date
    public let adults: Int
    public let departureAirport: String?
    public let residenceCountry: String?
    public let passportCountries: [String]
    public let interests: [String]

    public init(budgetMinor: Int64, currency: String, startDate: Date, endDate: Date, adults: Int, departureAirport: String? = nil, residenceCountry: String? = nil, passportCountries: [String] = [], interests: [String] = []) {
        self.budgetMinor = budgetMinor
        self.currency = currency
        self.startDate = startDate
        self.endDate = endDate
        self.adults = adults
        self.departureAirport = departureAirport
        self.residenceCountry = residenceCountry
        self.passportCountries = passportCountries
        self.interests = interests
    }

    public var nights: Int {
        max(1, Calendar(identifier: .gregorian).dateComponents([.day], from: startDate, to: endDate).day ?? 1)
    }
}

public protocol DiscoverGateway: Sendable {
    func matches(_ brief: DiscoverBrief) async throws -> [DestinationMatch]
}

public struct NuvroDiscoverGateway: DiscoverGateway {
    private let client: any APIProtocol

    public init(client: any APIProtocol) {
        self.client = client
    }

    public func matches(_ brief: DiscoverBrief) async throws -> [DestinationMatch] {
        let request = Components.Schemas.MatchRequest(
            adults: brief.adults,
            budget: Components.Schemas.Money(amountMinor: brief.budgetMinor, currency: brief.currency),
            departureAirport: brief.departureAirport,
            endDate: Self.day(brief.endDate),
            interests: brief.interests.isEmpty ? nil : brief.interests,
            passportCountries: brief.passportCountries.isEmpty ? nil : brief.passportCountries,
            residenceCountry: brief.residenceCountry,
            startDate: Self.day(brief.startDate)
        )
        switch try await client.discoverMatches(.init(body: .json(request))) {
        case .ok(let ok): return try ok.body.json.recommendations.map { DestinationMatch($0, nights: brief.nights) }
        case .default(let statusCode, let problem): throw APIFailure.from(problem, statusCode: statusCode)
        }
    }

    static func day(_ date: Date) -> String {
        date.formatted(.iso8601.year().month().day().dateSeparator(.dash))
    }
}

extension DestinationMatch {
    init(_ recommendation: Components.Schemas.Recommendation, nights: Int) {
        let destination = recommendation.destination
        let components = recommendation.personalMatch.components
        self.init(
            id: destination.id, name: destination.name,
            country: Locale.current.localizedString(forRegionCode: destination.countryCode) ?? destination.countryCode,
            region: destination.region ?? destination.name, photoReference: destination.id,
            photoURL: destination.media.flatMap { URL(string: $0.url) },
            score: (try? MatchScore(percent: min(100, max(0, recommendation.personalMatch.score)))) ?? MatchScore.floor,
            estimate: recommendation.affordability.estimatedTrip.flatMap(TripEstimate.init),
            weather: nil, flight: FlightSummary(recommendation.flight),
            durationDays: nights,
            pace: components.first { $0.key == "pace" }.map { Vocabulary.pace($0) } ?? "Balanced",
            tags: components.filter { $0.assessment.rawValue == "strong" }.map { "\($0.label) fit" },
            reasons: recommendation.reasons.compactMap(Vocabulary.reason),
            summary: recommendation.summary
        )
    }
}

extension MatchScore {
    static var floor: MatchScore { try! MatchScore(percent: 0) }
}

extension TripEstimate {
    init?(_ range: Components.Schemas.CostRange) {
        guard let low = try? Money(amountMinor: range.minimum.amountMinor, currency: range.minimum.currency),
              let high = try? Money(amountMinor: range.maximum.amountMinor, currency: range.maximum.currency) else { return nil }
        self.init(low: low, high: high)
    }
}

extension FlightSummary {
    init?(_ flight: Components.Schemas.FlightSummary) {
        guard let stops = flight.stops, let minutes = flight.durationMinutes else { return nil }
        self.init(stops: stops, approximateHours: max(1, Int((Double(minutes) / 60).rounded())))
    }
}

enum Vocabulary {
    static func pace(_ component: Components.Schemas.MatchComponent) -> String {
        switch component.explanation {
        case "profile_match": "Matches your pace"
        case "profile_mismatch": "Faster than you like"
        default: "Balanced"
        }
    }

    static func reason(_ code: String) -> MatchReason? {
        switch code {
        case "weather_fit": MatchReason(text: "Weather fits the dates you chose", verdict: .fits)
        case "crowds_fit": MatchReason(text: "Crowds match what you prefer", verdict: .fits)
        case "pace_fit": MatchReason(text: "The pace suits how you travel", verdict: .fits)
        case "interests_fit": MatchReason(text: "Strong overlap with your interests", verdict: .fits)
        case "budget_fit", "within_budget": MatchReason(text: "Within your budget", verdict: .fits)
        case "entry_evidence_unknown": MatchReason(text: "Entry requirements not yet checked", verdict: .tradeoff)
        case "cost_evidence_unknown": MatchReason(text: "Trip cost not yet estimated", verdict: .tradeoff)
        case "flight_evidence_unknown": MatchReason(text: "Flights not yet checked", verdict: .tradeoff)
        case "weather_mismatch": MatchReason(text: "Weather may not be what you asked for", verdict: .tradeoff)
        case "over_budget": MatchReason(text: "Likely over your budget", verdict: .tradeoff)
        default: nil
        }
    }
}
