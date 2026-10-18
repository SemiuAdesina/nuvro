import Foundation
import NuvroLocalization
import NuvroModels
import Observation

@Observable
@MainActor
public final class TripSummaryModel {
    public let summary: TripSummary
    public private(set) var loved: Set<String>
    public private(set) var notForMe: Set<String>
    public let copy: SummaryCopy
    public let locale: Locale

    public init(summary: TripSummary, loved: Set<String>, notForMe: Set<String>, language: Language = .english, locale: Locale = .current) {
        self.summary = summary
        self.loved = loved
        self.notForMe = notForMe
        self.copy = .resolve(language)
        self.locale = locale
    }

    public var totalLabel: String { summary.total.formattedWhole(locale: locale) }

    public var budgetLabel: String? {
        guard let difference = summary.budgetDifference, difference.amountMinor != 0,
              let magnitude = try? Money(amountMinor: abs(difference.amountMinor), currency: difference.currency) else { return nil }
        return "\(magnitude.formattedWhole(locale: locale)) \(difference.amountMinor > 0 ? copy.underBudget : copy.overBudget)"
    }

    public var facts: [(value: String, label: String)] {
        [("\(summary.days)", copy.days), ("\(summary.placesCount)", copy.places), ("\(summary.activitiesCount)", copy.activities), ("\(summary.kilometresExplored) km", copy.explored)]
    }

    public func toggleLoved(_ option: String) {
        if loved.contains(option) { loved.remove(option) } else { loved.insert(option) }
    }

    public func toggleNotForMe(_ option: String) {
        if notForMe.contains(option) { notForMe.remove(option) } else { notForMe.insert(option) }
    }
}
