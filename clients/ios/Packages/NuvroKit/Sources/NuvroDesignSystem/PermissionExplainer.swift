import SwiftUI

public struct NuvroPermissionExplainer: View {
    @Environment(\.designMetrics) private var metrics
    let symbol: String
    let palette: PhotoPalette
    let title: String
    let message: String
    let primary: (label: String, handler: () -> Void)
    let secondary: (label: String, handler: () -> Void)?

    public init(symbol: String, palette: PhotoPalette = .sky, title: String, message: String, primary: (label: String, handler: () -> Void), secondary: (label: String, handler: () -> Void)? = nil) {
        self.symbol = symbol
        self.palette = palette
        self.title = title
        self.message = message
        self.primary = primary
        self.secondary = secondary
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(12)) {
            IconTile(palette: palette, systemImage: symbol, size: 34)
            VStack(alignment: .leading, spacing: metrics.points(4)) {
                Text(title).typeStyle(Typography.cardHeading).foregroundStyle(Palette.label)
                Text(message).typeStyle(Typography.bodyCompact).foregroundStyle(Palette.labelAlpha(0.65))
            }
            HStack(spacing: metrics.points(10)) {
                Button(action: primary.handler) {
                    Text(primary.label).typeStyle(Typography.secondaryAction).foregroundStyle(Palette.cloud)
                        .padding(.horizontal, metrics.points(18)).frame(height: metrics.points(42))
                }
                .buttonStyle(GlassPillButtonStyle(style: .tinted(Palette.sky)))
                if let secondary {
                    Button(secondary.label, action: secondary.handler).typeStyle(Typography.link).foregroundStyle(Palette.labelAlpha(0.7)).buttonStyle(.plain)
                }
            }
            .padding(.top, metrics.points(2))
        }
        .padding(metrics.points(18))
        .frame(maxWidth: .infinity, alignment: .leading)
        .surfaceCard()
    }
}

public struct NuvroPickerRow: View {
    @Environment(\.designMetrics) private var metrics
    let label: String
    let value: String
    let action: () -> Void

    public init(_ label: String, value: String, action: @escaping () -> Void) {
        self.label = label
        self.value = value
        self.action = action
    }

    public var body: some View {
        Button(action: action) {
            HStack(spacing: metrics.points(12)) {
                Text(label).typeStyle(TypeStyle(size: 15.5, weight: .medium)).foregroundStyle(Palette.label).lineLimit(1).layoutPriority(1)
                Spacer(minLength: metrics.points(12))
                Text(value).typeStyle(TypeStyle(size: 14, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6)).lineLimit(1).truncationMode(.tail)
                ChevronGlyph()
            }
            .padding(.horizontal, metrics.points(15)).padding(.vertical, metrics.points(13))
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .accessibilityElement(children: .combine)
        .accessibilityValue(value)
        .accessibilityAddTraits(.isButton)
    }
}
