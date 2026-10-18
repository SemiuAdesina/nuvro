import Testing

@testable import NuvroLocalization

@Suite("OnboardingCopy")
struct OnboardingCopyTests {
    @Test("English is the canonical onboarding story")
    func englishStory() {
        let copy = OnboardingCopy.resolve(.english)

        #expect(copy.tagline == "The world, made personal.")
        #expect(copy.headline1 == "Not sure where to go?")
        #expect(copy.supporting1 == "Your budget. Your dates. Your vibe.")
        #expect(copy.headline2 == "Pick a place. We handle the rest.")
        #expect(copy.headline3 == "Your trip doesn’t stop when you arrive.")
        #expect(copy.callToAction3 == "Find my vacation")
        #expect(copy.secondaryAction3 == "I already have a trip")
    }

    @Test("languages without a translation fall back to English")
    func fallback() {
        #expect(OnboardingCopy.resolve(.korean) == OnboardingCopy.resolve(.english))
        #expect(OnboardingCopy.resolve(.arabic) == OnboardingCopy.resolve(.english))
        #expect(OnboardingCopy.resolve(.italian) == OnboardingCopy.resolve(.english))
    }

    @Test("translated languages differ from English")
    func translationsAreDistinct() {
        for language in OnboardingCopy.supported where language != .english {
            #expect(OnboardingCopy.resolve(language) != OnboardingCopy.resolve(.english), "\(language) is identical to English")
        }
    }

    @Test("every translated field is populated")
    func noEmptyFields() {
        for language in OnboardingCopy.supported {
            let copy = OnboardingCopy.resolve(language)
            let fields = [
                copy.tagline, copy.headline1, copy.supporting1, copy.body1, copy.callToAction1,
                copy.headline2, copy.body2, copy.callToAction2,
                copy.headline3, copy.supporting3, copy.body3, copy.callToAction3, copy.secondaryAction3,
                copy.skip, copy.languageSheetTitle, copy.languageSearch, copy.match, copy.you, copy.lunch,
                copy.outdoor, copy.indoor, copy.plansUpdated, copy.plansUpdatedDetail, copy.translation
            ]
            #expect(fields.allSatisfy { !$0.isEmpty }, "\(language) has an empty field")
        }
    }

    @Test("Japanese translation carries the design's copy")
    func japanese() {
        let copy = OnboardingCopy.resolve(.japanese)

        #expect(copy.tagline == "世界を、あなたのものに。")
        #expect(copy.callToAction3 == "私の休暇を見つける")
    }
}
