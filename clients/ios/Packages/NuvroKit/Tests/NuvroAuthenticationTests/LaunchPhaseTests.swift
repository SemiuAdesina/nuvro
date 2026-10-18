import Testing

@testable import NuvroAuthentication

@Suite("LaunchPhase")
struct LaunchPhaseTests {
    @Test("a new signed-in traveller goes to travel setup before the story")
    func newUser() {
        #expect(LaunchPhase(authentication: .signedIn(.fixture), hasCompletedTravelSetup: false, hasFinishedOnboarding: false) == .travelSetup)
    }

    @Test("a traveller who completed setup but not the story sees the story, and a returning traveller lands in the app")
    func returningUser() {
        #expect(LaunchPhase(authentication: .signedIn(.fixture), hasCompletedTravelSetup: true, hasFinishedOnboarding: false) == .onboarding)
        #expect(LaunchPhase(authentication: .signedIn(.fixture), hasCompletedTravelSetup: true, hasFinishedOnboarding: true) == .main)
    }

    @Test("authentication always precedes setup and restoration precedes authentication")
    func gates() {
        #expect(LaunchPhase(authentication: .signedOut(.none), hasCompletedTravelSetup: false, hasFinishedOnboarding: false) == .authentication)
        #expect(LaunchPhase(authentication: .expired, hasCompletedTravelSetup: true, hasFinishedOnboarding: true) == .authentication)
        #expect(LaunchPhase(authentication: .unknown, hasCompletedTravelSetup: true, hasFinishedOnboarding: true) == .restoringSession)
    }
}
