import Foundation

public enum DateSelection: Equatable, Sendable, Hashable, Codable {
    case exact(start: Date, end: Date)
    case flexible(month: Date, nights: Int)

    public var isFlexible: Bool {
        if case .flexible = self { return true }
        return false
    }

    public var exactStart: Date? {
        if case .exact(let start, _) = self { return start }
        return nil
    }

    public var flexibleMonth: Date? {
        if case .flexible(let month, _) = self { return month }
        return nil
    }

    public var nights: Int {
        switch self {
        case .exact(let start, let end): max(1, Calendar(identifier: .gregorian).dateComponents([.day], from: start, to: end).day ?? 1)
        case .flexible(_, let nights): nights
        }
    }

    public func label(locale: Locale, flexibleWord: String, aboutWord: String, nightsWord: String) -> String {
        switch self {
        case .exact(let start, let end):
            let range = start..<end
            return range.formatted(.interval.day().month(.abbreviated).locale(locale))
        case .flexible(let month, let nights):
            let monthName = month.formatted(.dateTime.month(.wide).locale(locale))
            return "\(flexibleWord) \(monthName) · \(aboutWord) \(nights) \(nightsWord)"
        }
    }

    public func normalised() -> DateSelection {
        switch self {
        case .exact(let start, let end) where end <= start:
            return .exact(start: start, end: Calendar(identifier: .gregorian).date(byAdding: .day, value: 1, to: start) ?? start)
        default:
            return self
        }
    }
}

public struct TravelerCounts: Equatable, Sendable, Hashable, Codable {
    public var adults: Int
    public var children: Int
    public var infants: Int

    public init(adults: Int = 2, children: Int = 0, infants: Int = 0) {
        self.adults = max(1, adults)
        self.children = max(0, children)
        self.infants = max(0, infants)
    }

    public var total: Int { adults + children + infants }

    public func label(adultWord: String, adultsWord: String, childWord: String, childrenWord: String, infantWord: String, infantsWord: String) -> String {
        var parts = ["\(adults) \(adults == 1 ? adultWord : adultsWord)"]
        if children > 0 { parts.append("\(children) \(children == 1 ? childWord : childrenWord)") }
        if infants > 0 { parts.append("\(infants) \(infants == 1 ? infantWord : infantsWord)") }
        return parts.joined(separator: ", ")
    }
}

public struct Airport: Sendable, Hashable, Codable, Identifiable {
    public let code: String
    public let city: String
    public let name: String
    public let country: String

    public init(code: String, city: String, name: String, country: String) {
        self.code = code
        self.city = city
        self.name = name
        self.country = country
    }

    public var id: String { code }
    public var shortLabel: String { "\(city) · \(code)" }

    public func matches(_ query: String) -> Bool {
        let needle = query.trimmingCharacters(in: .whitespaces).lowercased()
        guard !needle.isEmpty else { return true }
        return [code, city, name, country].contains { $0.lowercased().contains(needle) }
    }
}
