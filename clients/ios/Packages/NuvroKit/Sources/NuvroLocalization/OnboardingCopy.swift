import Foundation

public struct OnboardingCopy: Equatable, Sendable {
    public let tagline: String
    public let headline1: String
    public let supporting1: String
    public let body1: String
    public let callToAction1: String
    public let headline2: String
    public let body2: String
    public let callToAction2: String
    public let headline3: String
    public let supporting3: String
    public let body3: String
    public let callToAction3: String
    public let secondaryAction3: String
    public let skip: String
    public let languageSheetTitle: String
    public let languageSearch: String
    public let match: String
    public let you: String
    public let lunch: String
    public let outdoor: String
    public let indoor: String
    public let plansUpdated: String
    public let plansUpdatedDetail: String
    public let translation: String

    public static func resolve(_ language: Language) -> OnboardingCopy {
        translated[language] ?? english
    }

    public static var supported: [Language] {
        Language.allCases.filter { translated.keys.contains($0) }
    }

    static let translated: [Language: OnboardingCopy] = [
        .english: english,
        .spanish: spanish,
        .french: french,
        .portuguese: portuguese,
        .german: german,
        .japanese: japanese
    ]
}
