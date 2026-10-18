import NuvroDesignSystem
import NuvroModels
import SwiftUI

struct DestinationPickCard: View {
    @Environment(\.designMetrics) private var metrics
    let destination: DestinationMatch
    let meta: String
    let reason: String?
    let isSaved: Bool
    let onOpen: () -> Void
    let onSave: () -> Void

    var body: some View {
        Button(action: onOpen) {
            VStack(spacing: 0) {
                DynamicDestinationImage(media: MediaDescriptor(url: destination.photoURL, placeholder: destination.photoReference, altText: destination.displayName))
                    .frame(height: metrics.points(150))
                    .overlay(alignment: .topTrailing) {
                        SaveButton(isSaved: isSaved, style: .control, action: onSave)
                            .scaleEffect(34.0 / 38.0)
                            .padding(metrics.points(10))
                    }
                HStack(alignment: .center, spacing: metrics.points(10)) {
                    VStack(alignment: .leading, spacing: metrics.points(3)) {
                        Text(destination.displayName).typeStyle(Typography.cardTitleLarge).foregroundStyle(Palette.label)
                        Text(meta).typeStyle(Typography.meta).foregroundStyle(Palette.labelAlpha(0.6))
                    }
                    Spacer(minLength: 0)
                    if let reason {
                        Text(reason)
                            .typeStyle(Typography.captionEmphasis)
                            .foregroundStyle(Palette.labelAlpha(0.55))
                            .multilineTextAlignment(.trailing)
                            .lineLimit(2)
                            .frame(maxWidth: metrics.points(110), alignment: .trailing)
                    }
                }
                .padding(EdgeInsets(top: metrics.points(12), leading: metrics.points(14), bottom: metrics.points(13), trailing: metrics.points(14)))
            }
            .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.card), style: .continuous))
            .surfaceCard(radius: Radius.card)
        }
        .buttonStyle(PressScaleButtonStyle())
        .accessibilityElement(children: .combine)
        .accessibilityLabel("\(destination.displayName), \(meta)")
    }
}

struct InspirationChips: View {
    @Environment(\.designMetrics) private var metrics
    let themes: [InspirationTheme]
    let isSelected: (InspirationTheme) -> Bool
    let onToggle: (InspirationTheme) -> Void

    var body: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: metrics.points(8)) {
                ForEach(themes) { theme in
                    let selected = isSelected(theme)
                    Button { onToggle(theme) } label: {
                        Text(theme.title)
                            .typeStyle(Typography.control)
                            .foregroundStyle(selected ? Palette.background : Palette.label)
                            .padding(.horizontal, metrics.points(14))
                            .frame(height: metrics.points(36))
                            .background(selected ? Palette.label : Palette.surface, in: Capsule())
                            .overlay(Capsule().strokeBorder(selected ? Palette.label : Palette.labelAlpha(0.1), lineWidth: 0.5))
                    }
                    .buttonStyle(PressScaleButtonStyle(scale: 0.96))
                    .accessibilityAddTraits(selected ? .isSelected : [])
                    .animation(Motion.fade(0.2), value: selected)
                }
            }
            .padding(.horizontal, metrics.points(20))
        }
        .scrollClipDisabled()
    }
}
