import Foundation
import NuvroLocalization
import NuvroModels
import Observation

public enum StayPresentation: String, CaseIterable, Sendable {
    case list
    case map
}

@Observable
@MainActor
public final class StayModel {
    public let options: [StayOption]
    public var filter: StayOption.Recommendation = .bestOverall
    public var presentation: StayPresentation = .list
    public private(set) var savedIDs: Set<String> = []
    public let copy: StayCopy
    public let locale: Locale

    public init(options: [StayOption], language: Language = .english, locale: Locale = .current) {
        self.options = options
        self.copy = .resolve(language)
        self.locale = locale
    }

    public var hero: StayOption? {
        options.first { $0.recommendation == filter } ?? options.first
    }

    public var others: [StayOption] {
        options.filter { $0.id != hero?.id }
    }

    public func label(for recommendation: StayOption.Recommendation) -> String {
        switch recommendation {
        case .bestOverall: copy.bestOverall
        case .bestValue: copy.bestValue
        case .bestLocation: copy.bestLocation
        case .bestPremium: copy.bestPremium
        }
    }

    public func toggleSaved(_ option: StayOption) {
        if savedIDs.contains(option.id) { savedIDs.remove(option.id) } else { savedIDs.insert(option.id) }
    }

    public func isSaved(_ option: StayOption) -> Bool {
        savedIDs.contains(option.id)
    }

    public func nightlyLabel(_ option: StayOption) -> String {
        option.nightlyRate.formattedWhole(locale: locale)
    }

    public func totalLabel(_ option: StayOption) -> String {
        option.stayTotal?.formattedWhole(locale: locale) ?? ""
    }
}
