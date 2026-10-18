import Foundation
import NuvroLocalization
import NuvroModels

public struct DestinationFact: Equatable, Sendable, Identifiable {
    public let key: String
    public let value: String

    public var id: String { key }
}

public enum DestinationFacts {
    public static func glance(for match: DestinationMatch, estimate: Money?, copy: DestinationCopy, locale: Locale) -> [DestinationFact] {
        var facts = [DestinationFact(key: copy.length, value: "\(match.durationDays) days")]
        if let estimate {
            facts.append(DestinationFact(key: copy.trip, value: "~\(estimate.formattedWhole(locale: locale))"))
        }
        if let weather = match.weather { facts.append(DestinationFact(key: copy.climate, value: weather.condition)) }
        if let flight = match.flight { facts.append(DestinationFact(key: copy.flight, value: stopsLabel(flight, copy: copy))) }
        facts.append(DestinationFact(key: copy.pace, value: match.pace))
        return facts
    }

    public static func reasons(for match: DestinationMatch, budgetCeiling: Money?, estimate: Money?, copy: DestinationCopy, locale: Locale) -> [DestinationFact] {
        var facts: [DestinationFact] = []
        if let budgetCeiling, let estimate, budgetCeiling.currency == estimate.currency,
           let margin = try? Money(amountMinor: budgetCeiling.amountMinor - estimate.amountMinor, currency: estimate.currency),
           margin.amountMinor >= 0 {
            facts.append(DestinationFact(key: copy.budget, value: "\(margin.formattedWhole(locale: locale)) \(copy.under)"))
        }
        if let weather = match.weather { facts.append(DestinationFact(key: copy.weather, value: "\(weather.temperatureCelsius)° · \(weather.condition)")) }
        facts.append(DestinationFact(key: copy.styleMatch, value: match.tags.prefix(2).joined(separator: " + ")))
        if let flight = match.flight { facts.append(DestinationFact(key: copy.flight, value: "\(flight.approximateHours)h · \(stopsLabel(flight, copy: copy))")) }
        return facts
    }

    public static func summary(for match: DestinationMatch) -> String {
        if let summary = match.summary { return summary }
        let styles = match.tags.prefix(2).map { $0.lowercased() }.joined(separator: " and ")
        let weather = match.weather.map { "\($0.condition.lowercased())-weather preference, " } ?? ""
        return "\(match.name) matches your budget, \(weather)\(styles) interests, and \(match.pace.lowercased()) travel style."
    }

    static func stopsLabel(_ flight: FlightSummary, copy: DestinationCopy) -> String {
        "\(flight.stops) \(flight.stops == 1 ? copy.stop : copy.stops)"
    }
}
