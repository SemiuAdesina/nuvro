import SwiftUI
#if canImport(UIKit)
import UIKit
#endif

public enum NavigationTitleStyling {
    @MainActor
    public static func apply() {
        #if canImport(UIKit)
        let label = UIColor { traits in
            UIColor(traits.userInterfaceStyle == .dark ? Palette.labelDark : Palette.labelLight)
        }
        let appearance = UINavigationBar.appearance()
        appearance.largeTitleTextAttributes = [.foregroundColor: label]
        appearance.titleTextAttributes = [.foregroundColor: label]
        #endif
    }
}
