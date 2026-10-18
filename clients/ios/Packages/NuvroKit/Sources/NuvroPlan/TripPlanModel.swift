import Foundation
import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import Observation

public enum TripEditPresentation: String, Identifiable, Sendable {
    case menu
    case dates
    case travelers
    case budget

    public var id: String { rawValue }
}

@Observable
@MainActor
public final class TripPlanModel {
    public private(set) var plan: TripPlan
    public var askText = ""
    public var editing: TripEditPresentation?
    public var dateSelection: DateSelection
    public var travelerCounts: [String: Int]
    public var budgetMajor: Double
    public private(set) var lastEvent: String?

    public func clearEvent() {
        lastEvent = nil
    }
    private var applied: (dates: DateSelection, travelers: [String: Int], budget: Double)
    public let copy: TripPlanCopy
    public let pickerCopy: PickerCopy
    public let locale: Locale

    public init(plan: TripPlan, language: Language = .english, locale: Locale = .current) {
        self.plan = plan
        self.copy = .resolve(language)
        self.pickerCopy = .resolve(language)
        self.locale = locale
        let dates = DateSelection.exact(start: plan.startDate, end: Calendar.current.date(byAdding: .day, value: plan.destination.durationDays, to: plan.startDate) ?? plan.startDate)
        let counts = ["adults": 2, "children": 0, "infants": 0]
        let budget = Double(plan.budgetCeiling.amountMinor) / 100
        self.dateSelection = dates
        self.travelerCounts = counts
        self.budgetMajor = budget
        self.applied = (dates, counts, budget)
    }

    public var travelerLines: [CountPickerLine] {
        [
            CountPickerLine(id: "adults", title: pickerCopy.adults, detail: pickerCopy.adultsDetail, range: 1...8),
            CountPickerLine(id: "children", title: pickerCopy.children, detail: pickerCopy.childrenDetail, range: 0...6),
            CountPickerLine(id: "infants", title: pickerCopy.infants, detail: pickerCopy.infantsDetail, range: 0...4)
        ]
    }

    public var dateSheetCopy: DateSelectionCopy {
        DateSelectionCopy(title: pickerCopy.dates, exact: pickerCopy.exactDates, flexible: pickerCopy.flexibleDates, start: pickerCopy.startDate, end: pickerCopy.endDate, month: pickerCopy.month, nights: pickerCopy.nights, done: copy.done)
    }

    public func applyEdits() {
        let travelers = TravelerCounts(adults: travelerCounts["adults"] ?? 2, children: travelerCounts["children"] ?? 0, infants: travelerCounts["infants"] ?? 0)
        let travelersLabel = travelers.label(adultWord: pickerCopy.adult, adultsWord: pickerCopy.adults.lowercased(), childWord: pickerCopy.child, childrenWord: pickerCopy.children.lowercased(), infantWord: pickerCopy.infant, infantsWord: pickerCopy.infants.lowercased())
        let dateLabel = dateSelection.label(locale: locale, flexibleWord: pickerCopy.anytimeIn, aboutWord: pickerCopy.about, nightsWord: pickerCopy.nights.lowercased())
        guard let ceiling = try? Money(amountMinor: Int64(budgetMajor * 100), currency: plan.budgetCeiling.currency) else { return }
        let changed = dateSelection != applied.dates || travelerCounts != applied.travelers || budgetMajor != applied.budget
        guard changed else { return }
        applied = (dateSelection, travelerCounts, budgetMajor)
        plan = TripPlan(
            id: plan.id, destination: plan.destination, dateRangeLabel: dateLabel, travelersLabel: travelersLabel, startDate: dateSelection.exactStart ?? plan.startDate,
            budgetCeiling: ceiling, readiness: plan.readiness, budgetLines: plan.budgetLines, days: plan.days
        )
        lastEvent = copy.tripUpdated
    }

    public var readinessLabel: String {
        "\(plan.readyCount) of \(plan.readiness.count) \(copy.ready)"
    }

    public var subtitle: String {
        "\(plan.dateRangeLabel) · \(plan.destination.durationDays) days · \(plan.travelersLabel)"
    }

    public var estimateLabel: String {
        plan.estimatedTotal?.formattedWhole(locale: locale) ?? ""
    }

    public var budgetStatus: String? {
        guard let remaining = plan.remainingBudget else { return nil }
        let magnitude = try? Money(amountMinor: abs(remaining.amountMinor), currency: remaining.currency)
        guard let magnitude else { return nil }
        return "\(magnitude.formattedWhole(locale: locale)) \(remaining.amountMinor >= 0 ? copy.underBudget : copy.overBudget)"
    }

    public var isWithinBudget: Bool {
        (plan.remainingBudget?.amountMinor ?? 0) >= 0
    }

    public func share(of line: BudgetLine) -> Double {
        guard let total = plan.estimatedTotal, total.amountMinor > 0 else { return 0 }
        return Double(line.amount.amountMinor) / Double(total.amountMinor)
    }

    public func toggleReadiness(_ item: ReadinessItem) {
        let updated = plan.readiness.map { $0.id == item.id ? ReadinessItem(name: $0.name, isReady: !$0.isReady) : $0 }
        plan = TripPlan(
            id: plan.id, destination: plan.destination, dateRangeLabel: plan.dateRangeLabel, travelersLabel: plan.travelersLabel, startDate: plan.startDate,
            budgetCeiling: plan.budgetCeiling, readiness: updated, budgetLines: plan.budgetLines, days: plan.days
        )
    }

    public func apply(_ adjustment: PlanAdjustment) {
        guard let day = plan.day(adjustment.dayNumber) else { return }
        plan = plan.replacing(day: adjustment.apply(to: day))
    }

    public func suggest(_ text: String) {
        askText = text
    }
}
