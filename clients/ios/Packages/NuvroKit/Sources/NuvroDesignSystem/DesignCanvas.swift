import SwiftUI

public struct DesignMetrics: Equatable, Sendable {
    public static let referenceSize = CGSize(width: 402, height: 874)
    public static let referenceSafeTop: CGFloat = 62
    public static let referenceSafeBottom: CGFloat = 34

    public let size: CGSize
    public let safeTop: CGFloat
    public let safeBottom: CGFloat

    public init(size: CGSize, safeTop: CGFloat = referenceSafeTop, safeBottom: CGFloat = referenceSafeBottom) {
        self.size = size
        self.safeTop = safeTop
        self.safeBottom = safeBottom
    }

    public static let reference = DesignMetrics(size: referenceSize)

    public var scale: CGFloat {
        guard Self.referenceSize.width > 0, size.width > 0 else { return 1 }
        return size.width / Self.referenceSize.width
    }

    public var verticalSlack: CGFloat {
        size.height - Self.referenceSize.height * scale
    }

    public func points(_ designValue: CGFloat) -> CGFloat {
        designValue * scale
    }

    public func size(_ width: CGFloat, _ height: CGFloat) -> CGSize {
        CGSize(width: points(width), height: points(height))
    }

    public func x(_ designX: CGFloat) -> CGFloat {
        points(designX)
    }

    public func fromTop(_ designY: CGFloat) -> CGFloat {
        points(designY)
    }

    public func fromBottom(_ designDistance: CGFloat) -> CGFloat {
        size.height - points(designDistance)
    }

    public func belowSafeTop(_ designY: CGFloat) -> CGFloat {
        safeTop + points(designY - Self.referenceSafeTop)
    }

    public func aboveSafeBottom(_ designDistance: CGFloat) -> CGFloat {
        size.height - safeBottom - points(designDistance - Self.referenceSafeBottom)
    }

    public func topLeading(_ designX: CGFloat, _ designY: CGFloat, _ width: CGFloat, _ height: CGFloat) -> CGPoint {
        CGPoint(x: x(designX) + points(width) / 2, y: fromTop(designY) + points(height) / 2)
    }

    public func bottomLeading(_ designX: CGFloat, _ designBottom: CGFloat, _ width: CGFloat, _ height: CGFloat) -> CGPoint {
        CGPoint(x: x(designX) + points(width) / 2, y: fromBottom(designBottom) - points(height) / 2)
    }

    public var horizontalCenter: CGFloat {
        size.width / 2
    }
}

private struct DesignMetricsKey: EnvironmentKey {
    static let defaultValue = DesignMetrics.reference
}

public extension EnvironmentValues {
    var designMetrics: DesignMetrics {
        get { self[DesignMetricsKey.self] }
        set { self[DesignMetricsKey.self] = newValue }
    }
}

public struct DesignCanvas<Content: View>: View {
    private let fullBleed: Bool
    private let content: (DesignMetrics) -> Content

    public init(fullBleed: Bool = true, @ViewBuilder content: @escaping (DesignMetrics) -> Content) {
        self.fullBleed = fullBleed
        self.content = content
    }

    public var body: some View {
        GeometryReader { safeArea in
            if fullBleed {
                GeometryReader { fullBleedProxy in
                    let metrics = DesignMetrics(
                        size: fullBleedProxy.size,
                        safeTop: safeArea.safeAreaInsets.top,
                        safeBottom: safeArea.safeAreaInsets.bottom
                    )
                    content(metrics)
                        .environment(\.designMetrics, metrics)
                        .frame(width: fullBleedProxy.size.width, height: fullBleedProxy.size.height)
                }
                .ignoresSafeArea()
            } else {
                let insets = safeArea.safeAreaInsets
                let metrics = DesignMetrics(
                    size: CGSize(width: safeArea.size.width + insets.leading + insets.trailing, height: safeArea.size.height + insets.top + insets.bottom),
                    safeTop: insets.top,
                    safeBottom: insets.bottom
                )
                content(metrics)
                    .environment(\.designMetrics, metrics)
                    .frame(width: safeArea.size.width, height: safeArea.size.height)
            }
        }
    }
}
