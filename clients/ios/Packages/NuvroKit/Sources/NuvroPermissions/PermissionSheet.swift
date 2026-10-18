import NuvroDesignSystem
import SwiftUI

struct PermissionSheet: View {
    @Environment(\.designMetrics) private var metrics
    let center: PermissionCenter?
    let kind: PermissionKind
    let content: PermissionExplainerContent
    let fallback: (label: String, handler: () -> Void)?
    let openSettings: () -> Void

    private var isDenied: Bool {
        center?.presentation == .deniedGuidance(kind)
    }

    var body: some View {
        VStack(spacing: metrics.points(16)) {
            if isDenied {
                NuvroPermissionExplainer(
                    symbol: symbol, palette: .gold, title: content.deniedTitle, message: content.deniedMessage,
                    primary: (content.openSettings, openSettings),
                    secondary: fallback.map { fallback in (fallback.label, { fallback.handler(); center?.dismissPresentation() }) }
                )
            } else {
                NuvroPermissionExplainer(
                    symbol: symbol, title: content.title, message: content.message,
                    primary: (content.allow, { Task { _ = await center?.proceedFromExplainer() } }),
                    secondary: (content.notNow, { fallback?.handler(); center?.dismissPresentation() })
                )
            }
        }
        .padding(metrics.points(20))
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .top)
        .background(Palette.background)
        .presentationDetents([.medium])
        .presentationDragIndicator(.visible)
    }

    private var symbol: String {
        switch kind {
        case .location: "location.fill"
        case .notifications: "bell.badge.fill"
        case .camera: "camera.fill"
        case .microphone: "mic.fill"
        }
    }
}
