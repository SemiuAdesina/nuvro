import SwiftUI

public struct NuvroLoadingState: View {
    @Environment(\.designMetrics) private var metrics
    let label: String
    let isCompact: Bool

    public init(_ label: String, isCompact: Bool = false) {
        self.label = label
        self.isCompact = isCompact
    }

    public var body: some View {
        VStack(spacing: metrics.points(10)) {
            ProgressView().tint(Palette.sky)
            Text(label).typeStyle(Typography.caption).foregroundStyle(Palette.labelAlpha(0.55))
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, metrics.points(isCompact ? 18 : 60))
        .accessibilityElement(children: .combine)
    }
}

public struct NuvroEmptyState: View {
    @Environment(\.designMetrics) private var metrics
    let symbol: String
    let title: String
    let message: String
    let action: (label: String, handler: () -> Void)?

    public init(symbol: String, title: String, message: String, action: (label: String, handler: () -> Void)? = nil) {
        self.symbol = symbol
        self.title = title
        self.message = message
        self.action = action
    }

    public var body: some View {
        ContentUnavailableView {
            VStack(spacing: metrics.points(12)) {
                IconTile(palette: .sky, systemImage: symbol, size: 30)
                Text(title).typeStyle(Typography.title).foregroundStyle(Palette.label)
            }
        } description: {
            Text(message).typeStyle(Typography.body).foregroundStyle(Palette.labelAlpha(0.62))
        } actions: {
            if let action {
                Button(action: action.handler) {
                    Text(action.label).typeStyle(Typography.secondaryAction).foregroundStyle(Palette.cloud)
                        .padding(.horizontal, metrics.points(20)).frame(height: metrics.points(44))
                }
                .buttonStyle(GlassPillButtonStyle(style: .tinted(Palette.sky)))
                .padding(.top, metrics.points(6))
            }
        }
    }
}

public struct NuvroErrorState: View {
    @Environment(\.designMetrics) private var metrics
    let symbol: String
    let title: String
    let message: String
    let retry: (label: String, handler: () -> Void)?

    public init(symbol: String = "exclamationmark.triangle", title: String, message: String, retry: (label: String, handler: () -> Void)? = nil) {
        self.symbol = symbol
        self.title = title
        self.message = message
        self.retry = retry
    }

    public var body: some View {
        VStack(spacing: metrics.points(10)) {
            IconTile(palette: .gold, systemImage: symbol, size: 30)
            Text(title).typeStyle(Typography.cardHeading).foregroundStyle(Palette.label)
            Text(message).typeStyle(Typography.bodyCompact).foregroundStyle(Palette.labelAlpha(0.62)).multilineTextAlignment(.center)
            if let retry {
                SurfaceLinkButton(retry.label, action: retry.handler).padding(.top, metrics.points(4))
            }
        }
        .padding(metrics.points(22))
        .frame(maxWidth: .infinity)
        .surfaceCard()
        .accessibilityElement(children: .contain)
    }
}

public struct NuvroInlineError: View {
    @Environment(\.designMetrics) private var metrics
    let text: String

    public init(_ text: String) {
        self.text = text
    }

    public var body: some View {
        HStack(spacing: metrics.points(6)) {
            Image(systemName: "exclamationmark.circle.fill").font(.system(size: metrics.points(12), weight: .semibold)).accessibilityHidden(true)
            Text(text).typeStyle(Typography.caption)
        }
        .foregroundStyle(Palette.amber)
        .accessibilityElement(children: .combine)
    }
}

public struct FreshnessLabel: View {
    public enum Tone: Sendable { case fresh, cached, stale, unavailable }

    @Environment(\.designMetrics) private var metrics
    let tone: Tone
    let text: String

    public init(tone: Tone, text: String) {
        self.tone = tone
        self.text = text
    }

    private var colour: Color {
        switch tone {
        case .fresh: Palette.matchTeal
        case .cached, .stale: Palette.amber
        case .unavailable: Palette.labelAlpha(0.5)
        }
    }

    private var symbol: String {
        switch tone {
        case .fresh: "dot.radiowaves.left.and.right"
        case .cached: "arrow.down.circle"
        case .stale: "clock"
        case .unavailable: "questionmark.circle"
        }
    }

    public var body: some View {
        HStack(spacing: metrics.points(5)) {
            Image(systemName: symbol).font(.system(size: metrics.points(10), weight: .semibold)).accessibilityHidden(true)
            Text(text).typeStyle(Typography.smallEmphasis)
        }
        .foregroundStyle(colour)
        .accessibilityElement(children: .combine)
    }
}
