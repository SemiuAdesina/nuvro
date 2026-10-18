import SwiftUI

public struct ScreenTitle: ViewModifier {
    let title: String
    let subtitle: String?

    public func body(content: Content) -> some View {
        #if os(iOS)
        if #available(iOS 26.0, *) {
            content
                .navigationTitle(title)
                .navigationSubtitle(subtitle ?? "")
                .navigationBarTitleDisplayMode(.large)
        } else {
            content
                .navigationTitle(title)
                .navigationBarTitleDisplayMode(.large)
        }
        #else
        content.navigationTitle(title)
        #endif
    }
}

public extension View {
    func screenTitle(_ title: String, subtitle: String? = nil) -> some View {
        modifier(ScreenTitle(title: title, subtitle: subtitle))
    }
}
