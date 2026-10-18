import NuvroDesignSystem
import SwiftUI

struct ConversationalEntry: View {
    @Environment(\.designMetrics) private var metrics
    let placeholder: String
    let onTap: () -> Void
    let onVoice: () -> Void

    var body: some View {
        HStack(spacing: metrics.points(12)) {
            Button(action: onTap) {
                HStack(spacing: metrics.points(12)) {
                    RouteArcGlyph(width: metrics.points(18))
                    Text(placeholder)
                        .typeStyle(Typography.listRow)
                        .foregroundStyle(Palette.labelAlpha(0.5))
                        .lineLimit(1)
                        .frame(maxWidth: .infinity, alignment: .leading)
                }
                .padding(.leading, metrics.points(18))
                .frame(height: metrics.points(56))
                .contentShape(Rectangle())
            }
            .buttonStyle(PressScaleButtonStyle(scale: 0.985))
            .accessibilityLabel("Describe your ideal vacation")
            .accessibilityHint(placeholder)

            Button(action: onVoice) {
                Image(systemName: "mic.fill")
                    .font(.system(size: metrics.points(15), weight: .semibold))
                    .foregroundStyle(Palette.sky)
                    .frame(width: metrics.points(40), height: metrics.points(40))
                    .background(Palette.sky.opacity(0.12), in: Circle())
            }
            .buttonStyle(PressScaleButtonStyle(scale: 0.92))
            .padding(.trailing, metrics.points(8))
            .accessibilityLabel("Describe your vacation by voice")
        }
        .frame(height: metrics.points(56))
        .glassCapsule(.light)
    }
}
