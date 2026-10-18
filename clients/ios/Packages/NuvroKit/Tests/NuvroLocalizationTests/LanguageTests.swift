import Testing

@testable import NuvroLocalization

@Suite("Language")
struct LanguageTests {
    @Test("the picker offers the ten languages from the design in order")
    func pickerOrder() {
        #expect(Language.allCases.map(\.rawValue) == ["en", "es", "fr", "pt", "de", "it", "ar", "zh", "ja", "ko"])
    }

    @Test("native names are shown, never English exonyms")
    func nativeNames() {
        #expect(Language.spanish.nativeName == "Español")
        #expect(Language.japanese.nativeName == "日本語")
        #expect(Language.arabic.nativeName == "العربية")
    }

    @Test("only Arabic lays out right to left")
    func rightToLeft() {
        #expect(Language.arabic.isRightToLeft)
        #expect(Language.allCases.filter(\.isRightToLeft) == [.arabic])
    }

    @Test("device language detection honours preference order and region tags")
    func detection() {
        #expect(Language.detect(from: ["pt-BR", "en-GB"]) == .portuguese)
        #expect(Language.detect(from: ["nl-NL", "de-DE"]) == .german)
        #expect(Language.detect(from: ["JA"]) == .japanese)
    }

    @Test("unsupported device languages fall back to English")
    func detectionFallback() {
        #expect(Language.detect(from: ["nl-NL", "sv-SE"]) == .english)
        #expect(Language.detect(from: []) == .english)
    }

    @Test("search matches the native name or the code, case insensitively")
    func matching() {
        #expect(Language.matching("esp") == [.spanish])
        #expect(Language.matching("JA") == [.japanese])
        #expect(Language.matching("  ") == Language.allCases)
        #expect(Language.matching("xyz").isEmpty)
    }

    @Test("six languages carry onboarding translations")
    func translations() {
        let translated = Language.allCases.filter(\.hasOnboardingTranslation)

        #expect(translated == [.english, .spanish, .french, .portuguese, .german, .japanese])
        #expect(!Language.korean.hasOnboardingTranslation)
    }
}
