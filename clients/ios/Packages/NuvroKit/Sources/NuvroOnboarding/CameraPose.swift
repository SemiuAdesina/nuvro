import CoreGraphics

public struct CameraPose: Equatable, Sendable {
    public let zoom: CGFloat
    public let rise: CGFloat

    public init(zoom: CGFloat, rise: CGFloat) {
        self.zoom = zoom
        self.rise = rise
    }

    public static let anchor = CGPoint(x: 0.5, y: 0.42)

    public static let byStage: [OnboardingStage: CameraPose] = [
        .splash: CameraPose(zoom: 1.00, rise: 0),
        .discover: CameraPose(zoom: 1.22, rise: 70),
        .plan: CameraPose(zoom: 1.45, rise: 110),
        .travel: CameraPose(zoom: 2.10, rise: 210)
    ]

    public static func pose(for stage: OnboardingStage, reduceMotion: Bool) -> CameraPose {
        let effective = reduceMotion ? min(stage, .discover) : stage
        return byStage[effective] ?? byStage[.splash]!
    }

    public func layerScale(depth: CGFloat) -> CGFloat {
        1 + (zoom - 1) * depth
    }

    public func layerOffset(depth: CGFloat) -> CGFloat {
        rise * depth
    }
}

public enum LayerDepth {
    public static let far: CGFloat = 0.35
    public static let mid: CGFloat = 1.0
    public static let near: CGFloat = 1.25
}
