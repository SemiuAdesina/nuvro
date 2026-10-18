import Foundation
import NuvroModels

public struct TravelSetupDraft: Sendable, Hashable, Codable {
    public var stepIndex: Int = 0
    public var countryOfResidence: String?
    public var passportCountries: [String] = []
    public var hasDeclinedPassport = false
    public var permitAnswer: PermitAnswer?
    public var residencePermits: [String] = []
    public var departureAirport: String?
    public var currency: String?
    public var paperwork: PaperworkPreference?
    public var interests: Set<TravelPreference> = []
    public var pace: TravelPace?
    public var weather: Set<WeatherPreference> = []
    public var crowds: CrowdPreference?
    public var flights: Set<FlightPreference> = []
    public var stays: Set<StayPreference> = []
    public var group: TravelGroup?
    public var accessibility: Set<AccessibilityPreference> = []

    public enum PermitAnswer: String, Sendable, Codable {
        case yes, no, skipped
    }

    public init() {}

    public init(profile: TravelProfile) {
        countryOfResidence = profile.countryOfResidence
        passportCountries = profile.passportCountries
        hasDeclinedPassport = profile.hasDeclinedPassport
        permitAnswer = profile.residencePermits.isEmpty ? .no : .yes
        residencePermits = profile.residencePermits
        departureAirport = profile.preferredDepartureAirport
        currency = profile.preferredCurrency
        paperwork = profile.paperworkPreference
        interests = profile.interests
        pace = profile.pace
        weather = profile.weather
        crowds = profile.crowds
        flights = profile.flights
        stays = profile.stays
        group = profile.group
        accessibility = profile.accessibility
    }

    public func profile() -> TravelProfile? {
        guard let countryOfResidence, let currency, let paperwork, let pace, let crowds, let group, !interests.isEmpty, !weather.isEmpty else { return nil }
        guard !passportCountries.isEmpty || hasDeclinedPassport else { return nil }
        return TravelProfile(
            countryOfResidence: countryOfResidence, passportCountries: passportCountries, hasDeclinedPassport: hasDeclinedPassport, residencePermits: residencePermits,
            preferredDepartureAirport: departureAirport, preferredCurrency: currency, paperworkPreference: paperwork, interests: interests, pace: pace,
            weather: weather, crowds: crowds, flights: flights, stays: stays, group: group, accessibility: accessibility
        )
    }
}

public protocol TravelSetupStore: Sendable {
    func loadDraft() -> TravelSetupDraft?
    func saveDraft(_ draft: TravelSetupDraft)
    func clearDraft()
    func loadProfile() -> TravelProfile?
    func saveProfile(_ profile: TravelProfile)
}

public final class InMemoryTravelSetupStore: TravelSetupStore, @unchecked Sendable {
    private let lock = NSLock()
    private var draft: TravelSetupDraft?
    private var profile: TravelProfile?

    public init(draft: TravelSetupDraft? = nil, profile: TravelProfile? = nil) {
        self.draft = draft
        self.profile = profile
    }

    public func loadDraft() -> TravelSetupDraft? { lock.withLock { draft } }
    public func saveDraft(_ draft: TravelSetupDraft) { lock.withLock { self.draft = draft } }
    public func clearDraft() { lock.withLock { draft = nil } }
    public func loadProfile() -> TravelProfile? { lock.withLock { profile } }
    public func saveProfile(_ profile: TravelProfile) { lock.withLock { self.profile = profile } }
}

public struct UserDefaultsTravelSetupStore: TravelSetupStore, @unchecked Sendable {
    private let defaults: UserDefaults
    private let draftKey = "travelSetupDraft"
    private let profileKey = "travelProfile"

    public init(defaults: UserDefaults = .standard) {
        self.defaults = defaults
    }

    public func loadDraft() -> TravelSetupDraft? {
        defaults.data(forKey: draftKey).flatMap { try? JSONDecoder().decode(TravelSetupDraft.self, from: $0) }
    }

    public func saveDraft(_ draft: TravelSetupDraft) {
        if let data = try? JSONEncoder().encode(draft) { defaults.set(data, forKey: draftKey) }
    }

    public func clearDraft() {
        defaults.removeObject(forKey: draftKey)
    }

    public func loadProfile() -> TravelProfile? {
        defaults.data(forKey: profileKey).flatMap { try? JSONDecoder().decode(TravelProfile.self, from: $0) }
    }

    public func saveProfile(_ profile: TravelProfile) {
        if let data = try? JSONEncoder().encode(profile) { defaults.set(data, forKey: profileKey) }
    }
}
