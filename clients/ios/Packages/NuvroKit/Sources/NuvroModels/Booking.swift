import Foundation

public struct FlightOption: Sendable, Hashable, Codable, Identifiable {
    public enum Tier: String, Sendable, Codable, CaseIterable {
        case bestOverall = "best_overall"
        case lowestPrice = "lowest_price"
        case easiestJourney = "easiest_journey"
    }

    public let tier: Tier
    public let airline: String
    public let departs: String
    public let arrives: String
    public let duration: String
    public let stopsLabel: String
    public let stopCount: Int
    public let price: Money
    public let reason: String
    public let baggage: String
    public let checkedLabel: String

    public init(tier: Tier, airline: String, departs: String, arrives: String, duration: String, stopsLabel: String, stopCount: Int, price: Money, reason: String, baggage: String, checkedLabel: String) {
        self.tier = tier
        self.airline = airline
        self.departs = departs
        self.arrives = arrives
        self.duration = duration
        self.stopsLabel = stopsLabel
        self.stopCount = stopCount
        self.price = price
        self.reason = reason
        self.baggage = baggage
        self.checkedLabel = checkedLabel
    }

    public var id: String { tier.rawValue }

    private enum CodingKeys: String, CodingKey {
        case tier, airline, departs, arrives, duration, price, reason, baggage
        case stopsLabel = "stops_label"
        case stopCount = "stop_count"
        case checkedLabel = "checked_label"
    }
}

public struct HeroExperience: Sendable, Hashable, Codable {
    public let name: String
    public let subtitle: String
    public let durationLabel: String
    public let fromPriceLabel: String
    public let rating: Double
    public let fitTitle: String
    public let fitBody: String
    public let media: DestinationMedia

    public init(name: String, subtitle: String, durationLabel: String, fromPriceLabel: String, rating: Double, fitTitle: String, fitBody: String, media: DestinationMedia) {
        self.name = name
        self.subtitle = subtitle
        self.durationLabel = durationLabel
        self.fromPriceLabel = fromPriceLabel
        self.rating = rating
        self.fitTitle = fitTitle
        self.fitBody = fitBody
        self.media = media
    }

    private enum CodingKeys: String, CodingKey {
        case name, subtitle, rating, media
        case durationLabel = "duration_label"
        case fromPriceLabel = "from_price_label"
        case fitTitle = "fit_title"
        case fitBody = "fit_body"
    }
}

public struct Experience: Sendable, Hashable, Codable, Identifiable {
    public let name: String
    public let meta: String
    public let fit: String
    public let media: DestinationMedia

    public init(name: String, meta: String, fit: String, media: DestinationMedia) {
        self.name = name
        self.meta = meta
        self.fit = fit
        self.media = media
    }

    public var id: String { name }
}

public struct ArrivalOption: Sendable, Hashable, Codable, Identifiable {
    public enum Tag: String, Sendable, Codable {
        case recommended
        case bestForPlan = "best_for_plan"
    }

    public let name: String
    public let meta: String
    public let costLabel: String
    public let tag: Tag?
    public let placeholder: String

    public init(name: String, meta: String, costLabel: String, tag: Tag? = nil, placeholder: String) {
        self.name = name
        self.meta = meta
        self.costLabel = costLabel
        self.tag = tag
        self.placeholder = placeholder
    }

    public var id: String { name }

    private enum CodingKeys: String, CodingKey {
        case name, meta, tag, placeholder
        case costLabel = "cost_label"
    }
}

public struct LocalTransport: Sendable, Hashable, Codable, Identifiable {
    public let name: String
    public let note: String
    public let placeholder: String

    public init(name: String, note: String, placeholder: String) {
        self.name = name
        self.note = note
        self.placeholder = placeholder
    }

    public var id: String { name }
}

public struct TransportGuide: Sendable, Hashable, Codable {
    public let recommendationTitle: String
    public let recommendationBody: String
    public let arrivals: [ArrivalOption]
    public let transferLabel: String
    public let airport: GeoCoordinate
    public let hotel: GeoCoordinate
    public let local: [LocalTransport]

    public init(recommendationTitle: String, recommendationBody: String, arrivals: [ArrivalOption], transferLabel: String, airport: GeoCoordinate, hotel: GeoCoordinate, local: [LocalTransport]) {
        self.recommendationTitle = recommendationTitle
        self.recommendationBody = recommendationBody
        self.arrivals = arrivals
        self.transferLabel = transferLabel
        self.airport = airport
        self.hotel = hotel
        self.local = local
    }

    private enum CodingKeys: String, CodingKey {
        case arrivals, airport, hotel, local
        case recommendationTitle = "recommendation_title"
        case recommendationBody = "recommendation_body"
        case transferLabel = "transfer_label"
    }
}
