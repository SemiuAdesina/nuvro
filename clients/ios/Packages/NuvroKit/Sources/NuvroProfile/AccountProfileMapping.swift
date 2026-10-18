import Foundation
import NuvroAPI
import NuvroModels

extension ProfileModel {
    public func load(from gateway: any AccountGateway, name: String?) async {
        do {
            async let profile = gateway.profile()
            async let preferences = gateway.preferences()
            async let personalization = gateway.personalization()
            let snapshot = try await AccountSnapshot(profile: profile, preferences: preferences, personalization: personalization)
            apply(TravelerProfile.make(from: snapshot, name: name))
            usesTripHistory = snapshot.personalization.historyLearningEnabled
            loadFailure = nil
        } catch {
            loadFailure = LoadFailure(error: error)
        }
    }

    public func syncHistoryLearning(_ enabled: Bool, with gateway: any AccountGateway) async {
        guard enabled != profile.usesTripHistory else { return }
        do {
            let updated = try await gateway.setHistoryLearning(enabled)
            apply(profile.replacingLearning(enabled: updated.historyLearningEnabled, insights: updated.insights.map(LearnedInsight.init)))
            usesTripHistory = updated.historyLearningEnabled
        } catch {
            usesTripHistory = profile.usesTripHistory
            loadFailure = LoadFailure(error: error)
        }
    }
}

public struct AccountSnapshot: Sendable, Hashable {
    public let profile: AccountProfile
    public let preferences: AccountPreferences
    public let personalization: AccountPersonalization

    public init(profile: AccountProfile, preferences: AccountPreferences, personalization: AccountPersonalization) {
        self.profile = profile
        self.preferences = preferences
        self.personalization = personalization
    }
}

extension TravelerProfile {
    public static func make(from snapshot: AccountSnapshot, name: String?) -> TravelerProfile {
        let displayName = name?.trimmingCharacters(in: .whitespaces).nonEmpty ?? "Traveller"
        let preferences = snapshot.preferences
        return TravelerProfile(
            name: displayName, initial: String(displayName.prefix(1)).uppercased(),
            homeLabel: homeLabel(snapshot.profile),
            styleChips: styleChips(preferences),
            preferences: preferenceRows(preferences, currency: snapshot.profile.preferredCurrency),
            learned: snapshot.personalization.insights.map(LearnedInsight.init),
            usesTripHistory: snapshot.personalization.historyLearningEnabled
        )
    }

    func replacingLearning(enabled: Bool, insights: [LearnedInsight]) -> TravelerProfile {
        TravelerProfile(name: name, initial: initial, homeLabel: homeLabel, styleChips: styleChips, preferences: preferences, learned: insights, usesTripHistory: enabled)
    }

    private static func homeLabel(_ profile: AccountProfile) -> String {
        let country = profile.homeCountry.flatMap { Locale.current.localizedString(forRegionCode: $0) }
        return [country, profile.departureAirport].compactMap { $0 }.joined(separator: " · ").nonEmpty ?? "Add where you travel from"
    }

    private static func styleChips(_ preferences: AccountPreferences) -> [StyleChip] {
        var chips = preferences.interests.prefix(3).map { StyleChip(name: Vocabulary.interest($0), tone: .aqua) }
        chips.append(StyleChip(name: Vocabulary.pace(preferences.pace), tone: .sky))
        if preferences.weather != "any" { chips.append(StyleChip(name: Vocabulary.weather(preferences.weather), tone: .gold)) }
        return chips
    }

    private static func preferenceRows(_ preferences: AccountPreferences, currency: String) -> [PreferenceRow] {
        [
            PreferenceRow(name: "Interests", value: preferences.interests.map(Vocabulary.interest).joined(separator: ", "), placeholder: "curacao"),
            PreferenceRow(name: "Travel pace", value: Vocabulary.pace(preferences.pace), placeholder: "madeira"),
            PreferenceRow(name: "Weather", value: Vocabulary.weather(preferences.weather), placeholder: "gold"),
            PreferenceRow(name: "Crowds", value: Vocabulary.crowds(preferences.crowds), placeholder: "twilight"),
            PreferenceRow(name: "Flight time", value: preferences.maxFlightMinutes.map { "Up to \($0 / 60) h" } ?? "Any", placeholder: "sky"),
            PreferenceRow(name: "Travel group", value: preferences.travelingWithChildren ? "With children" : "No children", placeholder: "curacao"),
            PreferenceRow(name: "Currency", value: currency, placeholder: "gold")
        ]
    }
}

extension LearnedInsight {
    init(_ preference: LearnedPreference) {
        self.init(text: Vocabulary.insight(preference.summaryCode), source: "From \(preference.evidenceCount) \(preference.evidenceCount == 1 ? "signal" : "signals")")
    }
}

enum Vocabulary {
    static func interest(_ raw: String) -> String {
        switch raw {
        case "beach", "coastal": "Beach"
        case "food": "Food"
        case "culture": "Culture"
        case "nature": "Nature"
        case "nightlife": "Nightlife"
        default: raw.capitalized
        }
    }

    static func pace(_ raw: String) -> String {
        switch raw {
        case "relaxed": "Relaxed"
        case "packed": "Packed"
        default: "Balanced"
        }
    }

    static func weather(_ raw: String) -> String {
        switch raw {
        case "warm": "Warm"
        case "mild": "Mild"
        case "cold": "Cold"
        default: "Any weather"
        }
    }

    static func crowds(_ raw: String) -> String {
        switch raw {
        case "quiet": "Avoid very busy places"
        case "lively": "Lively places"
        default: "No preference"
        }
    }

    static func insight(_ code: String) -> String {
        switch code {
        case "insight.interest.coastal": "You often choose coastal destinations"
        case "insight.interest.food": "Food is usually part of your trips"
        case "insight.interest.culture": "You lean towards culture and history"
        case "insight.interest.nature": "You seek out nature and the outdoors"
        case "insight.interest.nightlife": "Nightlife tends to feature in your trips"
        case "insight.stay.quiet_neighbourhood": "You usually prefer quieter neighbourhoods"
        case "insight.stay.central_neighbourhood": "You usually stay somewhere central"
        case "insight.flight.fewer_stops": "You prefer flights with fewer stops"
        case "insight.pace.relaxed": "You like a relaxed pace"
        case "insight.pace.packed": "You like a full schedule"
        case "insight.transfer.low_tolerance": "Long transfers wear on you"
        default: "NUVRO noticed a pattern in your trips"
        }
    }
}

private extension String {
    var nonEmpty: String? { isEmpty ? nil : self }
}
