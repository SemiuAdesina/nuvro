import Foundation
import NuvroLocalization
import NuvroModels
import NuvroTravelSetup

public extension TravelerProfile {
    static func make(from travel: TravelProfile, name: String?, airports: [Airport], language: Language = .english, locale: Locale = .current) -> TravelerProfile {
        let copy = ProfileCopy.resolve(language)
        let summary = TravelProfileSummary(copy: .resolve(language), locale: locale)
        let displayName = name?.trimmingCharacters(in: .whitespaces).isEmpty == false ? name! : copy.travellerFallback
        let airport = airports.first { $0.code == travel.preferredDepartureAirport }
        let home = [airport?.shortLabel, TravelCatalog.regionName(travel.countryOfResidence, locale: locale)].compactMap { $0 }.joined(separator: " · ")
        let interests = TravelPreference.allCases.filter { travel.interests.contains($0) }
        let chips = interests.prefix(4).enumerated().map { index, interest in
            StyleChip(name: interest.title, tone: [StyleChip.Tone.aqua, .gold, .twilight, .sky][index % 4])
        }
        let lines = summary.lines(for: TravelSetupDraft(profile: travel), airports: airports)
        let preferenceKeys: [(TravelSetupStep, String)] = [(.interests, "Interests"), (.pace, "Travel pace"), (.weather, "Weather"), (.crowds, "Crowds"), (.flights, "Flight preferences"), (.stays, "Stay preferences"), (.group, "Travel group"), (.accessibility, "Accessibility")]
        let placeholders = ["curacao", "madeira", "gold", "twilight", "sky", "aqua", "guanacaste", "lisbon"]
        let preferences = preferenceKeys.enumerated().compactMap { index, entry -> PreferenceRow? in
            guard let line = lines.first(where: { $0.step == entry.0 }) else { return nil }
            return PreferenceRow(name: entry.1, value: line.value, placeholder: placeholders[index % placeholders.count])
        }
        return TravelerProfile(name: displayName, initial: String(displayName.prefix(1)).uppercased(), homeLabel: home, styleChips: Array(chips), preferences: preferences, learned: [], usesTripHistory: true)
    }
}

public struct TravelAccessRow: Identifiable, Sendable {
    public let step: TravelSetupStep
    public let title: String
    public let value: String

    public var id: Int { step.rawValue }

    public static func rows(for travel: TravelProfile, airports: [Airport], language: Language = .english, locale: Locale = .current) -> [TravelAccessRow] {
        let copy = ProfileCopy.resolve(language)
        let summary = TravelProfileSummary(copy: .resolve(language), locale: locale)
        let lines = summary.lines(for: TravelSetupDraft(profile: travel), airports: airports)
        func value(_ step: TravelSetupStep) -> String { lines.first { $0.step == step }?.value ?? "" }
        return [
            TravelAccessRow(step: .residence, title: copy.residence, value: value(.residence)),
            TravelAccessRow(step: .passports, title: copy.passport, value: value(.passports)),
            TravelAccessRow(step: .airport, title: copy.departureAirport, value: value(.airport)),
            TravelAccessRow(step: .paperwork, title: copy.paperwork, value: value(.paperwork))
        ]
    }
}
