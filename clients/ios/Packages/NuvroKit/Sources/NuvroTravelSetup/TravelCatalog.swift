import Foundation
import NuvroDesignSystem

public struct RegionOption: Identifiable, Hashable, Sendable {
    public let code: String
    public let name: String

    public init(code: String, name: String) {
        self.code = code
        self.name = name
    }

    public var id: String { code }

    public var flag: String {
        code.unicodeScalars.compactMap { UnicodeScalar(127397 + $0.value).map(String.init) }.joined()
    }

    public var label: String { name }
}

public struct CurrencyOption: Identifiable, Hashable, Sendable {
    public let code: String
    public let name: String

    public init(code: String, name: String) {
        self.code = code
        self.name = name
    }

    public var id: String { code }
    public var label: String { "\(code) — \(name)" }
}

public enum TravelCatalog {
    public static let permitZones: [PickerOption] = [PickerOption(id: "SCHENGEN", label: "Schengen area")]

    public static func regions(locale: Locale = .current) -> [RegionOption] {
        Locale.Region.isoRegions
            .filter { $0.identifier.count == 2 && $0.subRegions.isEmpty }
            .compactMap { region in locale.localizedString(forRegionCode: region.identifier).map { RegionOption(code: region.identifier, name: $0) } }
            .sorted { $0.name.localizedCaseInsensitiveCompare($1.name) == .orderedAscending }
    }

    public static func currencies(locale: Locale = .current) -> [CurrencyOption] {
        Locale.Currency.isoCurrencies
            .filter { $0.isISOCurrency }
            .compactMap { currency in locale.localizedString(forCurrencyCode: currency.identifier).map { CurrencyOption(code: currency.identifier, name: $0) } }
            .sorted { $0.name.localizedCaseInsensitiveCompare($1.name) == .orderedAscending }
    }

    public static func suggestedCurrency(forRegion code: String?) -> String? {
        guard let code else { return nil }
        return Locale(identifier: "und_\(code)").currency?.identifier
    }

    public static func regionName(_ code: String, locale: Locale = .current) -> String {
        locale.localizedString(forRegionCode: code) ?? code
    }

    public static func currencyName(_ code: String, locale: Locale = .current) -> String {
        locale.localizedString(forCurrencyCode: code) ?? code
    }

    public static func matches(_ region: RegionOption, _ query: String) -> Bool {
        wordPrefixMatch(query, in: [region.name, region.code])
    }

    public static func matches(_ currency: CurrencyOption, _ query: String) -> Bool {
        wordPrefixMatch(query, in: [currency.name, currency.code])
    }

    public static func wordPrefixMatch(_ query: String, in fields: [String]) -> Bool {
        let needle = query.trimmingCharacters(in: .whitespaces).folding(options: [.caseInsensitive, .diacriticInsensitive], locale: nil)
        guard !needle.isEmpty else { return true }
        return fields.contains { field in
            let folded = field.folding(options: [.caseInsensitive, .diacriticInsensitive], locale: nil)
            return folded.hasPrefix(needle) || folded.split(whereSeparator: { $0 == " " || $0 == "-" || $0 == "(" }).contains { $0.hasPrefix(needle) }
        }
    }
}
