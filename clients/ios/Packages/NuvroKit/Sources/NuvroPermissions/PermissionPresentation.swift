import SwiftUI

public struct PermissionExplainerContent: Sendable {
    public let symbol: String
    public let title: String
    public let message: String
    public let deniedTitle: String
    public let deniedMessage: String
    public let allow: String
    public let notNow: String
    public let openSettings: String

    public init(symbol: String, title: String, message: String, deniedTitle: String, deniedMessage: String, allow: String, notNow: String, openSettings: String) {
        self.symbol = symbol
        self.title = title
        self.message = message
        self.deniedTitle = deniedTitle
        self.deniedMessage = deniedMessage
        self.allow = allow
        self.notNow = notNow
        self.openSettings = openSettings
    }
}

public struct PermissionPresentationModifier: ViewModifier {
    @Environment(\.openURL) private var openURL
    let center: PermissionCenter?
    let kind: PermissionKind
    let content: PermissionExplainerContent
    let fallback: (label: String, handler: () -> Void)?

    public func body(content view: Content) -> some View {
        view.sheet(isPresented: Binding(get: { isShowing }, set: { if !$0 { center?.dismissPresentation() } })) {
            PermissionSheet(center: center, kind: kind, content: content, fallback: fallback, openSettings: openSystemSettings)
        }
    }

    private var isShowing: Bool {
        guard let center else { return false }
        return center.presentation == .explainer(kind) || center.presentation == .deniedGuidance(kind)
    }

    private func openSystemSettings() {
        if let url = SystemSettings.url { openURL(url) }
        center?.dismissPresentation()
    }
}

public extension View {
    func permissionPresentation(_ center: PermissionCenter?, for kind: PermissionKind, content: PermissionExplainerContent, fallback: (label: String, handler: () -> Void)? = nil) -> some View {
        modifier(PermissionPresentationModifier(center: center, kind: kind, content: content, fallback: fallback))
    }
}
