import SwiftUI

public struct TypeStyle: Equatable, Sendable {
    public let size: CGFloat
    public let weight: Font.Weight
    public let tracking: CGFloat
    public let lineSpacing: CGFloat

    public init(size: CGFloat, weight: Font.Weight, tracking: CGFloat = 0, lineSpacing: CGFloat = 0) {
        self.size = size
        self.weight = weight
        self.tracking = tracking
        self.lineSpacing = lineSpacing
    }

    public func font(scaledBy scale: CGFloat) -> Font {
        .system(size: size * scale, weight: weight)
    }
}

public enum Typography {
    public static let headline = TypeStyle(size: 31, weight: .bold, tracking: -0.6)
    public static let supporting = TypeStyle(size: 17, weight: .semibold, tracking: -0.1)
    public static let body = TypeStyle(size: 15, weight: .regular, lineSpacing: 3)
    public static let tagline = TypeStyle(size: 15, weight: .medium, tracking: 0.3)
    public static let callToAction = TypeStyle(size: 17, weight: .semibold, tracking: -0.2)
    public static let control = TypeStyle(size: 14, weight: .semibold)
    public static let secondaryAction = TypeStyle(size: 15, weight: .semibold)
    public static let sheetTitle = TypeStyle(size: 20, weight: .bold, tracking: -0.3)
    public static let listRow = TypeStyle(size: 16, weight: .regular)
    public static let cardTitle = TypeStyle(size: 15, weight: .semibold, tracking: -0.2)
    public static let cardTitleCompact = TypeStyle(size: 13, weight: .semibold)
    public static let caption = TypeStyle(size: 12, weight: .regular)
    public static let captionEmphasis = TypeStyle(size: 12, weight: .semibold)
    public static let small = TypeStyle(size: 11, weight: .regular)
    public static let smallEmphasis = TypeStyle(size: 11, weight: .semibold)
    public static let micro = TypeStyle(size: 10, weight: .regular)
    public static let microEmphasis = TypeStyle(size: 10, weight: .semibold)
    public static let estimate = TypeStyle(size: 15, weight: .bold)
    public static let photoLabel = TypeStyle(size: 9, weight: .medium, tracking: 0.54)

    public static let largeTitle = TypeStyle(size: 34, weight: .bold, tracking: -0.7)
    public static let title = TypeStyle(size: 20, weight: .bold, tracking: -0.3)
    public static let cardHeading = TypeStyle(size: 17, weight: .bold, tracking: -0.2)
    public static let heroTitle = TypeStyle(size: 23, weight: .bold, tracking: -0.4)
    public static let greeting = TypeStyle(size: 15, weight: .semibold)
    public static let fieldValue = TypeStyle(size: 17, weight: .semibold)
    public static let budgetFigure = TypeStyle(size: 40, weight: .bold, tracking: -1)
    public static let estimateFigure = TypeStyle(size: 30, weight: .bold, tracking: -0.7)
    public static let matchFigure = TypeStyle(size: 22, weight: .bold, tracking: -0.5)
    public static let matchLabel = TypeStyle(size: 9, weight: .bold, tracking: 1.08)
    public static let compactMatchFigure = TypeStyle(size: 16, weight: .bold, tracking: -0.3)
    public static let compactMatchLabel = TypeStyle(size: 8, weight: .bold, tracking: 0.96)
    public static let cardTitleLarge = TypeStyle(size: 16, weight: .semibold, tracking: -0.2)
    public static let meta = TypeStyle(size: 13, weight: .regular)
    public static let metaEmphasis = TypeStyle(size: 13, weight: .semibold)
    public static let fieldLabel = TypeStyle(size: 12, weight: .semibold)
    public static let factKey = TypeStyle(size: 10, weight: .semibold, tracking: 0.2)
    public static let bodyCompact = TypeStyle(size: 14, weight: .regular, lineSpacing: 3)
    public static let bodyCompactEmphasis = TypeStyle(size: 14, weight: .bold, tracking: -0.2)
    public static let link = TypeStyle(size: 14, weight: .semibold)
    public static let dayLabel = TypeStyle(size: 12, weight: .bold, tracking: 0.24)
    public static let statusPill = TypeStyle(size: 12, weight: .bold)
    public static let tileLabel = TypeStyle(size: 11.5, weight: .semibold)
}

public struct TypeStyleModifier: ViewModifier {
    @Environment(\.designMetrics) private var metrics
    @ScaledMetric(relativeTo: .body) private var dynamicType: CGFloat = 1
    let style: TypeStyle

    public func body(content: Content) -> some View {
        let scale = metrics.scale * min(dynamicType, 1.6)
        content
            .font(style.font(scaledBy: scale))
            .tracking(style.tracking * scale)
            .lineSpacing(style.lineSpacing * scale)
    }
}

public extension View {
    func typeStyle(_ style: TypeStyle) -> some View {
        modifier(TypeStyleModifier(style: style))
    }
}
