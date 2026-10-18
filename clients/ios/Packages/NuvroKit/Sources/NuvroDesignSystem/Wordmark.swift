import CoreText
import SwiftUI

public enum BrandTypeface {
    public static let outfitSemibold = "Outfit-SemiBold"

    public static var isOutfitRegistered: Bool {
        let descriptor = CTFontDescriptorCreateWithNameAndSize(outfitSemibold as CFString, 12)
        let font = CTFontCreateWithFontDescriptor(descriptor, 12, nil)
        let resolved = CTFontCopyPostScriptName(font) as String
        return resolved == outfitSemibold
    }

    public static func wordmark(size: CGFloat) -> Font {
        isOutfitRegistered ? .custom(outfitSemibold, size: size) : .system(size: size, weight: .semibold)
    }

    public static let wordmarkTrackingRatio: CGFloat = 0.16
}

public struct Wordmark: View {
    private let size: CGFloat
    private let color: Color

    public init(size: CGFloat = 46, color: Color = Palette.cloud) {
        self.size = size
        self.color = color
    }

    private var arcWidth: CGFloat { size }
    private var arcOffset: CGSize { CGSize(width: -6 * size / 46, height: -8 * size / 46) }

    public var body: some View {
        HStack(spacing: 0) {
            letters("NU")
            letters("V")
                .overlay(alignment: .topLeading) {
                    RouteArcMark(width: arcWidth).offset(arcOffset)
                }
            letters("RO")
        }
        .shadow(color: Palette.navy.opacity(0.35), radius: 12, y: 2)
        .accessibilityElement(children: .ignore)
        .accessibilityLabel("NUVRO")
    }

    private func letters(_ text: String) -> some View {
        Text(text)
            .font(BrandTypeface.wordmark(size: size))
            .tracking(size * BrandTypeface.wordmarkTrackingRatio)
            .foregroundStyle(color)
    }
}
