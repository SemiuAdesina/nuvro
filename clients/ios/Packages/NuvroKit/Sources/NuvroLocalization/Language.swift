import Foundation

public enum Language: String, CaseIterable, Sendable, Identifiable, Codable {
    case english = "en"
    case spanish = "es"
    case french = "fr"
    case portuguese = "pt"
    case german = "de"
    case italian = "it"
    case arabic = "ar"
    case chinese = "zh"
    case japanese = "ja"
    case korean = "ko"

    public var id: String { rawValue }

    public var nativeName: String {
        switch self {
        case .english: "English"
        case .spanish: "Español"
        case .french: "Français"
        case .portuguese: "Português"
        case .german: "Deutsch"
        case .italian: "Italiano"
        case .arabic: "العربية"
        case .chinese: "中文"
        case .japanese: "日本語"
        case .korean: "한국어"
        }
    }

    public var isRightToLeft: Bool {
        self == .arabic
    }

    public var hasOnboardingTranslation: Bool {
        OnboardingCopy.translated.keys.contains(self)
    }

    public static func detect(from preferredLanguages: [String]) -> Language {
        for identifier in preferredLanguages {
            let code = identifier.split(separator: "-").first.map(String.init) ?? identifier
            if let match = Language(rawValue: code.lowercased()) {
                return match
            }
        }
        return .english
    }

    public static func matching(_ query: String) -> [Language] {
        let needle = query.trimmingCharacters(in: .whitespaces).lowercased()
        guard !needle.isEmpty else { return allCases }
        return allCases.filter {
            $0.nativeName.lowercased().contains(needle) || $0.rawValue.contains(needle)
        }
    }
}
