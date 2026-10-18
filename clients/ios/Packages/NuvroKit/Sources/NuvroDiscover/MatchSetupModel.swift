import Foundation
import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import Observation

public enum DateChoice: String, CaseIterable, Sendable {
    case exact
    case flexible
}

public enum SetupPresentation: String, Identifiable, Sendable {
    case airport
    case dates
    case travelers

    public var id: String { rawValue }
}

public enum Climate: String, CaseIterable, Sendable { case warm = "Warm", mild = "Mild", cold = "Cold" }
public enum Pace: String, CaseIterable, Sendable { case quiet = "Quiet", balanced = "Balanced", lively = "Lively" }
public enum Tier: String, CaseIterable, Sendable { case budget = "Budget", balanced = "Balanced", premium = "Premium" }
public enum Scope: String, CaseIterable, Sendable { case domestic = "Domestic", international = "International", anywhere = "Anywhere" }

public struct BudgetScale: Sendable, Equatable {
    public let range: ClosedRange<Double>
    public let step: Double
    public let initial: Double

    public init(range: ClosedRange<Double>, step: Double, initial: Double) {
        self.range = range
        self.step = step
        self.initial = initial
    }

    public static let usDollars = BudgetScale(range: 800...8000, step: 100, initial: 3000)
}

@Observable
@MainActor
public final class MatchSetupModel {
    public static let travelerRange: ClosedRange<Int> = 1...8
    public let budgetRange: ClosedRange<Double>
    public let budgetStep: Double

    public var departureAirport: Airport?
    public let airports: [Airport]
    public var budgetMajor: Double
    public var isBudgetFlexible = true
    public var dateSelection: DateSelection
    public var travelers = TravelerCounts()
    public var presentation: SetupPresentation?
    public private(set) var preferences: Set<TravelPreference> = [.beach, .food, .relaxation]
    public var isFineTuneExpanded = false
    public var climate: Climate = .warm
    public var pace: Pace = .quiet
    public var tier: Tier = .balanced
    public var scope: Scope = .anywhere
    public var prefersDirectFlights = false
    public var prefersEasierEntry = true
    public let copy: MatchSetupCopy
    public let pickerCopy: PickerCopy
    public let currency: String
    public let locale: Locale
    public let today: Date

    public init(airports: [Airport] = [], departureCode: String? = nil, language: Language = .english, currency: String = "USD", locale: Locale = .current, budgetScale: BudgetScale = .usDollars, interests: Set<TravelPreference>? = nil, today: Date = .now, calendar: Calendar = .current) {
        self.airports = airports
        self.budgetRange = budgetScale.range
        self.budgetStep = budgetScale.step
        self.budgetMajor = budgetScale.initial
        if let interests, !interests.isEmpty { self.preferences = interests }
        self.departureAirport = airports.first { $0.code == departureCode } ?? airports.first
        self.copy = .resolve(language)
        self.pickerCopy = .resolve(language)
        self.currency = currency
        self.locale = locale
        self.today = today
        let nextMonth = calendar.date(byAdding: .month, value: 1, to: today) ?? today
        self.dateSelection = .flexible(month: calendar.date(from: calendar.dateComponents([.year, .month], from: nextMonth)) ?? nextMonth, nights: 7)
    }

    public var departure: String {
        departureAirport?.shortLabel ?? ""
    }

    public var dateChoice: DateChoice {
        get { dateSelection.isFlexible ? .flexible : .exact }
        set {
            guard newValue != dateChoice else { return }
            switch newValue {
            case .flexible:
                dateSelection = .flexible(month: dateSelection.exactStart ?? today, nights: dateSelection.nights)
            case .exact:
                let start = dateSelection.flexibleMonth ?? today
                dateSelection = .exact(start: start, end: Calendar.current.date(byAdding: .day, value: dateSelection.nights, to: start) ?? start)
            }
        }
    }

    public var adults: Int {
        get { travelers.adults }
        set { travelers.adults = max(1, newValue) }
    }

    public var budget: Money? {
        try? Money(amountMinor: Int64(budgetMajor * 100), currency: currency)
    }

    public var budgetLabel: String {
        budget?.formattedWhole(locale: locale) ?? ""
    }

    public var budgetHint: String {
        let fraction = (budgetMajor - budgetRange.lowerBound) / max(budgetRange.upperBound - budgetRange.lowerBound, 1)
        switch fraction {
        case ..<0.14: return "Short trips and nearby escapes"
        case ..<0.38: return "Most of Europe and the Caribbean fit"
        case ..<0.66: return "Long-haul and premium stays open up"
        default: return "Almost anywhere, in comfort"
        }
    }

    public var travelersLabel: String {
        travelers.label(adultWord: pickerCopy.adult, adultsWord: pickerCopy.adults.lowercased(), childWord: pickerCopy.child, childrenWord: pickerCopy.children.lowercased(), infantWord: pickerCopy.infant, infantsWord: pickerCopy.infants.lowercased())
    }

    public var datesLabel: String {
        dateSelection.label(locale: locale, flexibleWord: pickerCopy.anytimeIn, aboutWord: pickerCopy.about, nightsWord: pickerCopy.nights.lowercased())
    }

    public var travelerCounts: [String: Int] {
        get { ["adults": travelers.adults, "children": travelers.children, "infants": travelers.infants] }
        set { travelers = TravelerCounts(adults: newValue["adults"] ?? travelers.adults, children: newValue["children"] ?? travelers.children, infants: newValue["infants"] ?? travelers.infants) }
    }

    public var travelerLines: [CountPickerLine] {
        [
            CountPickerLine(id: "adults", title: pickerCopy.adults, detail: pickerCopy.adultsDetail, range: 1...8),
            CountPickerLine(id: "children", title: pickerCopy.children, detail: pickerCopy.childrenDetail, range: 0...6),
            CountPickerLine(id: "infants", title: pickerCopy.infants, detail: pickerCopy.infantsDetail, range: 0...4)
        ]
    }

    public var dateSheetCopy: DateSelectionCopy {
        DateSelectionCopy(title: pickerCopy.dates, exact: copy.exactDates, flexible: copy.flexibleDates, start: pickerCopy.startDate, end: pickerCopy.endDate, month: pickerCopy.month, nights: pickerCopy.nights, done: stateCopy.done)
    }

    public var stateCopy: StateCopy { .resolve(.english) }

    public var preferenceCountLabel: String {
        "\(preferences.count) \(copy.selected)"
    }

    public var fineTuneSummary: String {
        [climate.rawValue, pace.rawValue, tier.rawValue, scope.rawValue].joined(separator: " · ")
    }

    public func toggle(_ preference: TravelPreference) {
        if preferences.contains(preference) {
            preferences.remove(preference)
        } else {
            preferences.insert(preference)
        }
    }

    public func isSelected(_ preference: TravelPreference) -> Bool {
        preferences.contains(preference)
    }

    public var canSubmit: Bool {
        !preferences.isEmpty && departureAirport != nil
    }
}
