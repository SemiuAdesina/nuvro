import SwiftUI

public struct GradientProgressBar: View {
    @Environment(\.designMetrics) private var metrics
    let fraction: Double
    let height: CGFloat
    let isMuted: Bool

    public init(fraction: Double, height: CGFloat = 5, isMuted: Bool = false) {
        self.fraction = fraction
        self.height = height
        self.isMuted = isMuted
    }

    public var body: some View {
        GeometryReader { proxy in
            ZStack(alignment: .leading) {
                Capsule().fill(Palette.labelAlpha(0.07))
                Capsule()
                    .fill(isMuted ? AnyShapeStyle(Palette.labelAlpha(0.18)) : AnyShapeStyle(LinearGradient(colors: [Palette.aqua, Palette.sky], startPoint: .leading, endPoint: .trailing)))
                    .frame(width: proxy.size.width * min(max(fraction, 0), 1))
                    .animation(Motion.settle(0.6), value: fraction)
            }
        }
        .frame(height: metrics.points(height))
        .accessibilityElement(children: .ignore)
        .accessibilityValue("\(Int((fraction * 100).rounded())) percent")
    }
}

public struct EyebrowLabel: View {
    let text: String
    let colour: Color

    public init(_ text: String, colour: Color = Palette.labelAlpha(0.5)) {
        self.text = text
        self.colour = colour
    }

    public var body: some View {
        Text(text).typeStyle(TypeStyle(size: 11, weight: .bold, tracking: 1.3)).foregroundStyle(colour)
    }
}

public struct ChevronGlyph: View {
    @Environment(\.designMetrics) private var metrics

    public init() {}

    public var body: some View {
        Image(systemName: "chevron.forward").font(.system(size: metrics.points(12), weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.3))
    }
}

public struct GlassPillButton: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let systemImage: String?
    let action: () -> Void

    public init(_ title: String, systemImage: String? = nil, action: @escaping () -> Void) {
        self.title = title
        self.systemImage = systemImage
        self.action = action
    }

    public var body: some View {
        Button(action: action) {
            HStack(spacing: metrics.points(7)) {
                if let systemImage { Image(systemName: systemImage).font(.system(size: metrics.points(13), weight: .bold)).foregroundStyle(Palette.sky) }
                Text(title).typeStyle(Typography.control).foregroundStyle(Palette.label)
            }
            .padding(.horizontal, metrics.points(15))
            .frame(height: metrics.points(38))
            .glassCapsule(.light)
        }
        .buttonStyle(PressScaleButtonStyle(scale: 0.96))
    }
}
