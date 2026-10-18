public enum LaunchPhase: Equatable, Sendable {
    case restoringSession
    case authentication
    case travelSetup
    case onboarding
    case main

    public init(authentication: AuthenticationState, hasCompletedTravelSetup: Bool, hasFinishedOnboarding: Bool) {
        switch authentication {
        case .unknown, .refreshing:
            self = .restoringSession
        case .signedOut, .expired, .authenticating:
            self = .authentication
        case .signedIn:
            if !hasCompletedTravelSetup {
                self = .travelSetup
            } else if !hasFinishedOnboarding {
                self = .onboarding
            } else {
                self = .main
            }
        }
    }
}
