import Foundation

public enum PaperworkPreference: String, CaseIterable, Sendable, Codable {
    case keepSimple = "keep_simple"
    case evisaFine = "evisa_fine"
    case anyDestination = "any_destination"
}

public enum TravelPace: String, CaseIterable, Sendable, Codable {
    case relaxed, balanced, full
}

public enum WeatherPreference: String, CaseIterable, Sendable, Codable {
    case warm, mild, cool
    case cold = "cold_snow"
    case noPreference = "no_preference"
}

public enum CrowdPreference: String, CaseIterable, Sendable, Codable {
    case quiet, balanced
    case fine = "crowds_fine"
}

public enum FlightPreference: String, CaseIterable, Sendable, Codable {
    case fewerStops = "fewer_stops"
    case avoidOvernight = "avoid_overnight"
    case shortest
    case lowestPrice = "lowest_price"
    case noPreference = "no_preference"
}

public enum StayPreference: String, CaseIterable, Sendable, Codable {
    case quietNeighbourhood = "quiet_neighbourhood"
    case central
    case nearNightlife = "near_nightlife"
    case nearBeach = "near_beach"
    case hotel, apartment, resort
    case budgetFriendly = "budget_friendly"
    case premium
}

public enum TravelGroup: String, CaseIterable, Sendable, Codable {
    case solo, partner, friends, family
    case changes = "it_changes"
}

public enum AccessibilityPreference: String, CaseIterable, Sendable, Codable {
    case none = "no_requirements"
    case mobility
    case wheelchair = "wheelchair_access"
    case reducedWalking = "reduced_walking"
    case vision, hearing, other
}

public struct TravelProfile: Sendable, Hashable, Codable {
    public var countryOfResidence: String
    public var passportCountries: [String]
    public var hasDeclinedPassport: Bool
    public var residencePermits: [String]
    public var preferredDepartureAirport: String?
    public var preferredCurrency: String
    public var paperworkPreference: PaperworkPreference
    public var interests: Set<TravelPreference>
    public var pace: TravelPace
    public var weather: Set<WeatherPreference>
    public var crowds: CrowdPreference
    public var flights: Set<FlightPreference>
    public var stays: Set<StayPreference>
    public var group: TravelGroup
    public var accessibility: Set<AccessibilityPreference>

    public init(countryOfResidence: String, passportCountries: [String], hasDeclinedPassport: Bool = false, residencePermits: [String] = [], preferredDepartureAirport: String?, preferredCurrency: String, paperworkPreference: PaperworkPreference, interests: Set<TravelPreference>, pace: TravelPace, weather: Set<WeatherPreference>, crowds: CrowdPreference, flights: Set<FlightPreference>, stays: Set<StayPreference>, group: TravelGroup, accessibility: Set<AccessibilityPreference>) {
        self.countryOfResidence = countryOfResidence
        self.passportCountries = passportCountries
        self.hasDeclinedPassport = hasDeclinedPassport
        self.residencePermits = residencePermits
        self.preferredDepartureAirport = preferredDepartureAirport
        self.preferredCurrency = preferredCurrency
        self.paperworkPreference = paperworkPreference
        self.interests = interests
        self.pace = pace
        self.weather = weather
        self.crowds = crowds
        self.flights = flights
        self.stays = stays
        self.group = group
        self.accessibility = accessibility
    }

    public var hasPassportContext: Bool { !passportCountries.isEmpty }

    private enum CodingKeys: String, CodingKey {
        case countryOfResidence = "country_of_residence"
        case passportCountries = "passport_countries"
        case hasDeclinedPassport = "has_declined_passport"
        case residencePermits = "residence_permits"
        case preferredDepartureAirport = "preferred_departure_airport"
        case preferredCurrency = "preferred_currency"
        case paperworkPreference = "travel_paperwork_preference"
        case interests = "travel_interests"
        case pace = "travel_pace"
        case weather = "weather_preference"
        case crowds = "crowd_preference"
        case flights = "flight_preference"
        case stays = "stay_preference"
        case group = "typical_travel_group"
        case accessibility = "accessibility_preferences"
    }
}

public enum TravelPreference: String, CaseIterable, Sendable, Codable, Identifiable {
    case beach = "Beach"
    case food = "Food"
    case culture = "Culture"
    case nature = "Nature"
    case adventure = "Adventure"
    case nightlife = "Nightlife"
    case relaxation = "Relaxation"
    case shopping = "Shopping"
    case family = "Family"
    case history = "History"

    public var id: String { rawValue }
    public var title: String { rawValue }

    public var systemImage: String {
        switch self {
        case .beach: "beach.umbrella.fill"
        case .food: "fork.knife"
        case .culture: "building.columns.fill"
        case .nature: "leaf.fill"
        case .nightlife: "moon.stars.fill"
        case .adventure: "figure.hiking"
        case .relaxation: "sun.horizon.fill"
        case .family: "figure.2.and.child.holdinghands"
        case .shopping: "bag.fill"
        case .history: "building.2.fill"
        }
    }
}
