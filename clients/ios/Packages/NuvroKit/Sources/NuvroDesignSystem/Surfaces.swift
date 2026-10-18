import SwiftUI

public struct SurfaceCard: ViewModifier {
    @Environment(\.designMetrics) private var metrics
    let radius: CGFloat
    let elevated: Bool

    public func body(content: Content) -> some View {
        let shape = RoundedRectangle(cornerRadius: metrics.points(radius), style: .continuous)
        return content
            .background(Palette.surface, in: shape)
            .overlay(shape.strokeBorder(Palette.labelAlpha(0.07), lineWidth: 0.5))
            .shadow(color: .clear, radius: 0)
    }
}

public extension View {
    func surfaceCard(radius: CGFloat = Radius.card, elevated: Bool = false) -> some View {
        modifier(SurfaceCard(radius: radius, elevated: elevated))
    }
}

public struct SectionHeader: View {
    let title: String
    let action: String?
    let onAction: (() -> Void)?

    public init(_ title: String, action: String? = nil, onAction: (() -> Void)? = nil) {
        self.title = title
        self.action = action
        self.onAction = onAction
    }

    public var body: some View {
        HStack(alignment: .firstTextBaseline) {
            Text(title).typeStyle(Typography.title).foregroundStyle(Palette.label)
            Spacer()
            if let action, let onAction {
                LinkButton(action, action: onAction)
            }
        }
    }
}

public struct TagChip: View {
    public enum Tone: Sendable {
        case aqua, gold, twilight, sky

        var fill: Color {
            switch self {
            case .aqua: Palette.aqua.opacity(0.18)
            case .gold: Palette.gold.opacity(0.25)
            case .twilight: Palette.twilight.opacity(0.12)
            case .sky: Palette.sky.opacity(0.14)
            }
        }

        var text: Color {
            switch self {
            case .aqua: Palette.matchTeal
            case .gold: Palette.amber
            case .twilight: Palette.twilight
            case .sky: Palette.sky
            }
        }
    }

    @Environment(\.designMetrics) private var metrics
    let text: String
    let tone: Tone

    public init(_ text: String, tone: Tone) {
        self.text = text
        self.tone = tone
    }

    public var body: some View {
        Text(text)
            .typeStyle(Typography.captionEmphasis)
            .foregroundStyle(tone.text)
            .padding(.horizontal, metrics.points(10))
            .padding(.vertical, metrics.points(5))
            .background(tone.fill, in: Capsule())
    }
}

public struct FactTile: View {
    @Environment(\.designMetrics) private var metrics
    let key: String
    let value: String
    let onSurface: Bool

    public init(key: String, value: String, onSurface: Bool = false) {
        self.key = key
        self.value = value
        self.onSurface = onSurface
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(3)) {
            Text(key).typeStyle(Typography.smallEmphasis).foregroundStyle(Palette.labelAlpha(0.55))
            Text(value)
                .typeStyle(Typography.bodyCompactEmphasis)
                .foregroundStyle(Palette.label)
                .lineLimit(1)
                .minimumScaleFactor(0.8)
        }
        .padding(.horizontal, metrics.points(12))
        .padding(.vertical, metrics.points(10))
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(
            onSurface ? Palette.surface.opacity(0.8) : Palette.labelAlpha(0.045),
            in: RoundedRectangle(cornerRadius: metrics.points(Radius.chip), style: .continuous)
        )
    }
}

public struct FieldLabel: View {
    let text: String

    public init(_ text: String) {
        self.text = text
    }

    public var body: some View {
        Text(text).typeStyle(Typography.fieldLabel).foregroundStyle(Palette.labelAlpha(0.55))
    }
}

public struct RouteArcGlyph: View {
    let width: CGFloat

    public init(width: CGFloat = 18) {
        self.width = width
    }

    private var height: CGFloat { width * 30 / 46 }

    public var body: some View {
        ZStack {
            RouteArc()
                .stroke(Palette.gold, style: StrokeStyle(lineWidth: width / 9, lineCap: .round))
            Circle()
                .fill(Palette.gold)
                .frame(width: width * 0.24, height: width * 0.24)
                .position(RouteArc.pinCenter(in: CGRect(x: 0, y: 0, width: width, height: height)))
        }
        .frame(width: width, height: height)
        .accessibilityHidden(true)
    }
}
