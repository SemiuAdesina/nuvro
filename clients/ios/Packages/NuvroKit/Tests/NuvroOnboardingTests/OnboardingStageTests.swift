import Testing

@testable import NuvroOnboarding

@Suite("OnboardingStage")
struct OnboardingStageTests {
    @Test("the splash is not a screen and has no page")
    func splash() {
        #expect(!OnboardingStage.splash.isScreen)
        #expect(OnboardingStage.splash.pageIndex == nil)
        #expect(OnboardingStage.splash.previousScreen == nil)
        #expect(OnboardingStage.splash.next == .discover)
    }

    @Test("the three screens are paged zero through two")
    func pages() {
        #expect(OnboardingStage.screens.map(\.pageIndex) == [0, 1, 2])
        #expect(OnboardingStage.travel.next == nil)
        #expect(OnboardingStage.plan.previousScreen == .discover)
        #expect(OnboardingStage.discover.previousScreen == nil)
    }

    @Test("stages order chronologically")
    func ordering() {
        #expect(OnboardingStage.splash < .discover)
        #expect(OnboardingStage.plan < .travel)
        #expect(min(OnboardingStage.travel, .discover) == .discover)
    }
}
