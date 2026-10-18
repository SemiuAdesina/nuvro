import Foundation
import NuvroAuthentication
import NuvroTravelSetup

enum LaunchResets {
    static func apply(arguments: [String] = CommandLine.arguments, defaults: UserDefaults = .standard) {
        if arguments.contains("-uiTestResetSession") { try? KeychainSessionStore().clear() }
        if arguments.contains("-uiTestResetOnboarding") { defaults.set(false, forKey: "hasFinishedOnboarding") }
        if arguments.contains("-uiTestResetTravelSetup") {
            defaults.set(false, forKey: "hasCompletedTravelSetup")
            UserDefaultsTravelSetupStore(defaults: defaults).clearDraft()
        }
    }
}
