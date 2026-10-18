import SwiftUI

public struct MatchBadge: View {
    @Environment(\.designMetrics) private var metrics
    let percent: Int
    let onPhoto: Bool

    public init(percent: Int, onPhoto: Bool = false) {
        self.percent = percent
        self.onPhoto = onPhoto
    }

    public var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: metrics.points(6)) {
            Text("\(percent)%").typeStyle(TypeStyle(size: 15, weight: .bold, tracking: -0.2))
            Text("MATCH").typeStyle(Typography.matchLabel)
        }
        .foregroundStyle(onPhoto ? Palette.aqua : Palette.matchTeal)
        .padding(.horizontal, metrics.points(12))
        .padding(.vertical, metrics.points(8))
        .background(Palette.aqua.opacity(onPhoto ? 0.22 : 0.2), in: Capsule())
        .overlay(Capsule().strokeBorder(Palette.aqua.opacity(0.5), lineWidth: 0.5))
        .accessibilityElement(children: .ignore)
        .accessibilityLabel("\(percent) percent match")
    }
}

public struct CompactMatchBadge: View {
    @Environment(\.designMetrics) private var metrics
    let percent: Int

    public init(percent: Int) {
        self.percent = percent
    }

    public var body: some View {
        VStack(spacing: metrics.points(1)) {
            Text("\(percent)%").typeStyle(Typography.compactMatchFigure)
            Text("MATCH").typeStyle(Typography.compactMatchLabel)
        }
        .foregroundStyle(Palette.matchTeal)
        .padding(.horizontal, metrics.points(10))
        .padding(.vertical, metrics.points(6))
        .background(Palette.aqua.opacity(0.16), in: RoundedRectangle(cornerRadius: metrics.points(Radius.stop), style: .continuous))
        .accessibilityElement(children: .ignore)
        .accessibilityLabel("\(percent) percent match")
    }
}

public struct MatchRing: View {
    @Environment(\.designMetrics) private var metrics
    let percent: Int
    let isRevealed: Bool

    public init(percent: Int, isRevealed: Bool) {
        self.percent = percent
        self.isRevealed = isRevealed
    }

    private var diameter: CGFloat { metrics.points(84) }

    public var body: some View {
        ZStack {
            Circle().fill(Palette.navy.opacity(0.45))
            Circle().strokeBorder(Color.white.opacity(0.35), lineWidth: 0.5)
            Circle()
                .stroke(Color.white.opacity(0.18), lineWidth: metrics.points(4))
                .padding(metrics.points(6))
            Circle()
                .trim(from: 0, to: isRevealed ? Double(percent) / 100 : 0)
                .stroke(Palette.aqua, style: StrokeStyle(lineWidth: metrics.points(4), lineCap: .round))
                .rotationEffect(.degrees(-90))
                .padding(metrics.points(6))
                .animation(Motion.settle(1.0), value: isRevealed)
            VStack(spacing: metrics.points(3)) {
                Text("\(percent)%").typeStyle(Typography.matchFigure)
                Text("MATCH").typeStyle(Typography.matchLabel)
            }
            .foregroundStyle(Palette.cloud)
        }
        .frame(width: diameter, height: diameter)
        .accessibilityElement(children: .ignore)
        .accessibilityLabel("\(percent) percent match")
    }
}

public struct MatchReasonRow: View {
    @Environment(\.designMetrics) private var metrics
    let text: String
    let isPositive: Bool

    public init(text: String, isPositive: Bool) {
        self.text = text
        self.isPositive = isPositive
    }

    public var body: some View {
        HStack(spacing: metrics.points(10)) {
            Image(systemName: isPositive ? "checkmark" : "exclamationmark.triangle.fill")
                .font(.system(size: metrics.points(10), weight: .bold))
                .foregroundStyle(isPositive ? Palette.matchTeal : Palette.amber)
                .frame(width: metrics.points(20), height: metrics.points(20))
                .background(
                    (isPositive ? Palette.aqua.opacity(0.2) : Palette.gold.opacity(0.3)),
                    in: Circle()
                )
            Text(text)
                .typeStyle(Typography.bodyCompact)
                .foregroundStyle(isPositive ? Palette.label : Palette.labelAlpha(0.75))
        }
        .accessibilityElement(children: .ignore)
        .accessibilityLabel("\(isPositive ? "Fits" : "Trade-off"): \(text)")
    }
}
