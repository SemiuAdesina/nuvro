import SwiftUI

public struct InsightCard<Content: View>: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let showsMark: Bool
    let content: Content

    public init(title: String, showsMark: Bool = true, @ViewBuilder content: () -> Content) {
        self.title = title
        self.showsMark = showsMark
        self.content = content()
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(10)) {
            HStack(spacing: metrics.points(8)) {
                if showsMark { RouteArcGlyph(width: metrics.points(16)) }
                Text(title).typeStyle(TypeStyle(size: 15, weight: .bold)).foregroundStyle(Palette.label)
            }
            content
        }
        .padding(metrics.points(16))
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(
            LinearGradient(colors: [Palette.sky.opacity(0.08), Palette.aqua.opacity(0.10)], startPoint: .topLeading, endPoint: .bottomTrailing),
            in: RoundedRectangle(cornerRadius: metrics.points(Radius.card), style: .continuous)
        )
        .background(Palette.surface, in: RoundedRectangle(cornerRadius: metrics.points(Radius.card), style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: metrics.points(Radius.card), style: .continuous).strokeBorder(Palette.sky.opacity(0.18), lineWidth: 0.5))
    }
}

public struct SurfaceLinkButton: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let action: () -> Void

    public init(_ title: String, action: @escaping () -> Void) {
        self.title = title
        self.action = action
    }

    public var body: some View {
        Button(action: action) {
            Text(title)
                .typeStyle(Typography.link)
                .foregroundStyle(Palette.sky)
                .padding(.horizontal, metrics.points(16))
                .frame(height: metrics.points(38))
                .background(Palette.surface.opacity(0.9), in: Capsule())
                .overlay(Capsule().strokeBorder(Palette.labelAlpha(0.1), lineWidth: 0.5))
                .shadow(color: .clear, radius: 0)
        }
        .buttonStyle(PressScaleButtonStyle(scale: 0.97))
    }
}

public struct CheckRow: View {
    @Environment(\.designMetrics) private var metrics
    let text: String
    let size: CGFloat

    public init(_ text: String, size: CGFloat = 14) {
        self.text = text
        self.size = size
    }

    public var body: some View {
        HStack(spacing: metrics.points(10)) {
            Image(systemName: "checkmark")
                .font(.system(size: metrics.points(10), weight: .bold))
                .foregroundStyle(Palette.matchTeal)
                .frame(width: metrics.points(20), height: metrics.points(20))
                .background(Palette.aqua.opacity(0.2), in: Circle())
            Text(text).typeStyle(TypeStyle(size: size, weight: .regular)).foregroundStyle(Palette.label)
        }
    }
}

public struct BudgetRing: View {
    @Environment(\.designMetrics) private var metrics
    let segments: [(color: Color, fraction: Double)]
    let centreTitle: String
    let centreLabel: String

    public init(segments: [(color: Color, fraction: Double)], centreTitle: String, centreLabel: String) {
        self.segments = segments
        self.centreTitle = centreTitle
        self.centreLabel = centreLabel
    }

    public var body: some View {
        ZStack {
            Circle().stroke(Palette.labelAlpha(0.07), lineWidth: metrics.points(9))
            ForEach(Array(segments.enumerated()), id: \.offset) { index, segment in
                let start = segments.prefix(index).reduce(0) { $0 + $1.fraction }
                Circle()
                    .trim(from: start, to: start + segment.fraction)
                    .stroke(segment.color, lineWidth: metrics.points(9))
                    .rotationEffect(.degrees(-90))
            }
            VStack(spacing: metrics.points(2)) {
                Text(centreTitle).typeStyle(TypeStyle(size: 17, weight: .bold, tracking: -0.3)).foregroundStyle(Palette.label)
                Text(centreLabel).typeStyle(TypeStyle(size: 9, weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.55))
            }
        }
        .padding(metrics.points(5))
        .frame(width: metrics.points(92), height: metrics.points(92))
        .accessibilityElement(children: .ignore)
        .accessibilityLabel("\(centreTitle) \(centreLabel)")
    }
}

public struct GlassActionTile: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let systemImage: String
    let palette: PhotoPalette
    let action: () -> Void

    public init(title: String, systemImage: String, palette: PhotoPalette, action: @escaping () -> Void) {
        self.title = title
        self.systemImage = systemImage
        self.palette = palette
        self.action = action
    }

    public var body: some View {
        Button(action: action) {
            VStack(spacing: metrics.points(6)) {
                IconTile(palette: palette, systemImage: systemImage, size: 24)
                Text(title).typeStyle(Typography.tileLabel).foregroundStyle(Palette.label).lineLimit(1).minimumScaleFactor(0.8)
            }
            .frame(maxWidth: .infinity)
            .padding(.vertical, metrics.points(12))
            .padding(.horizontal, metrics.points(6))
        }
        .buttonStyle(GlassTileButtonStyle())
    }
}

public struct GlassTileButtonStyle: ButtonStyle {
    @Environment(\.designMetrics) private var metrics

    public init() {}

    public func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .glass(.light, in: RoundedRectangle(cornerRadius: metrics.points(Radius.tile), style: .continuous))
            .shadow(color: Palette.labelAlpha(0.04), radius: 2, y: 1)
            .scaleEffect(configuration.isPressed ? 0.96 : 1)
            .animation(.easeOut(duration: 0.15), value: configuration.isPressed)
    }
}
