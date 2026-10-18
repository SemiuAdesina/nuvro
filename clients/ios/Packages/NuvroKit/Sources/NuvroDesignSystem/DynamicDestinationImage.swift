import SwiftUI

public struct MediaDescriptor: Sendable, Hashable {
    public let url: URL?
    public let placeholder: String
    public let altText: String
    public let attribution: String?

    public init(url: URL?, placeholder: String, altText: String, attribution: String? = nil) {
        self.url = url
        self.placeholder = placeholder
        self.altText = altText
        self.attribution = attribution
    }
}

public struct DynamicDestinationImage: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @State private var isRevealed = false
    let media: MediaDescriptor
    let label: String
    let labelSize: CGFloat

    public init(media: MediaDescriptor, label: String = "", labelSize: CGFloat = 9) {
        self.media = media
        self.label = label
        self.labelSize = labelSize
    }

    public var body: some View {
        ZStack {
            PhotoStandIn(palette: .named(media.placeholder), label: label, labelSize: labelSize)
            if let url = media.url {
                AsyncImage(url: url, transaction: Transaction(animation: reduceMotion ? nil : .easeOut(duration: 0.45))) { phase in
                    if let image = phase.image {
                        image.resizable().scaledToFill().transition(.opacity)
                    }
                }
            }
        }
        .clipped()
        .accessibilityLabel(media.altText)
    }
}

public struct AttributionLabel: View {
    @Environment(\.designMetrics) private var metrics
    let text: String

    public init(_ text: String) {
        self.text = text
    }

    public var body: some View {
        Text(text)
            .typeStyle(Typography.micro)
            .foregroundStyle(Palette.labelAlpha(0.45))
            .lineLimit(1)
    }
}

public struct NuvroMapMarker: View {
    @Environment(\.designMetrics) private var metrics
    let letter: String
    let colour: Color
    let isSelected: Bool

    public init(letter: String, colour: Color, isSelected: Bool = false) {
        self.letter = letter
        self.colour = colour
        self.isSelected = isSelected
    }

    public var body: some View {
        Text(letter)
            .font(.system(size: metrics.points(11), weight: .bold, design: .rounded))
            .foregroundStyle(.white)
            .frame(width: metrics.points(isSelected ? 30 : 26), height: metrics.points(isSelected ? 30 : 26))
            .background(colour, in: Circle())
            .overlay(Circle().strokeBorder(.white, lineWidth: 2.5))
            .shadow(color: Palette.navy.opacity(isSelected ? 0.4 : 0.3), radius: isSelected ? 6 : 4, y: 3)
            .animation(Motion.fade(0.2), value: isSelected)
    }
}

public struct OfflineStatus: View {
    @Environment(\.designMetrics) private var metrics
    let isAvailable: Bool
    let availableText: String
    let missingText: String

    public init(isAvailable: Bool, availableText: String, missingText: String) {
        self.isAvailable = isAvailable
        self.availableText = availableText
        self.missingText = missingText
    }

    public var body: some View {
        HStack(spacing: metrics.points(5)) {
            Image(systemName: isAvailable ? "arrow.down.circle.fill" : "circle.dashed")
                .font(.system(size: metrics.points(11), weight: .semibold))
            Text(isAvailable ? availableText : missingText).typeStyle(Typography.smallEmphasis)
        }
        .foregroundStyle(isAvailable ? Palette.matchTeal : Palette.labelAlpha(0.45))
        .accessibilityElement(children: .combine)
    }
}

public struct LanguagePairSelector: View {
    @Environment(\.designMetrics) private var metrics
    let source: String
    let target: String
    let swap: () -> Void

    public init(source: String, target: String, swap: @escaping () -> Void) {
        self.source = source
        self.target = target
        self.swap = swap
    }

    public var body: some View {
        HStack(spacing: metrics.points(10)) {
            Text(source).typeStyle(Typography.control).foregroundStyle(Palette.label)
            Button(action: swap) {
                Image(systemName: "arrow.left.arrow.right")
                    .font(.system(size: metrics.points(12), weight: .bold))
                    .foregroundStyle(Palette.sky)
                    .frame(width: metrics.points(28), height: metrics.points(28))
                    .background(Palette.sky.opacity(0.12), in: Circle())
            }
            .buttonStyle(PressScaleButtonStyle(scale: 0.9))
            .accessibilityLabel("Swap languages")
            Text(target).typeStyle(Typography.control).foregroundStyle(Palette.label)
        }
        .padding(.horizontal, metrics.points(14))
        .frame(height: metrics.points(42))
        .glassCapsule(.light)
    }
}

public extension View {
    func photoScrim(_ gradient: LinearGradient) -> some View {
        overlay(Rectangle().fill(gradient))
    }
}
