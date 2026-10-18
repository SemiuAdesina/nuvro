import NuvroDesignSystem
import SwiftUI

@main
struct NuvroApp: App {
    init() {
        LaunchResets.apply()
        NavigationTitleStyling.apply()
    }

    var body: some Scene {
        WindowGroup {
            RootView()
        }
    }
}
