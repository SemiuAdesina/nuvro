import CoreGraphics

public struct ContentPose: Equatable, Sendable {
    public let x: CGFloat
    public let y: CGFloat
    public let scale: CGFloat
    public let opacity: Double

    public init(_ x: CGFloat, _ y: CGFloat, _ scale: CGFloat, _ opacity: Double) {
        self.x = x
        self.y = y
        self.scale = scale
        self.opacity = opacity
    }

    public var isVisible: Bool { opacity > 0 }
}

public enum StagePoses {
    public static func primaryCard(_ stage: OnboardingStage) -> ContentPose {
        switch stage {
        case .splash: ContentPose(101, 300, 0.60, 0)
        case .discover: ContentPose(101, 118, 1.00, 1)
        case .plan: ContentPose(71, 112, 1.30, 1)
        case .travel: ContentPose(71, 40, 1.30, 0)
        }
    }

    public static func secondaryCard(_ stage: OnboardingStage) -> ContentPose {
        switch stage {
        case .splash: ContentPose(60, 380, 0.50, 0)
        case .discover: ContentPose(14, 292, 0.88, 0.94)
        case .plan, .travel: ContentPose(-60, 330, 0.80, 0)
        }
    }

    public static func tertiaryCard(_ stage: OnboardingStage) -> ContentPose {
        switch stage {
        case .splash: ContentPose(250, 380, 0.50, 0)
        case .discover: ContentPose(248, 318, 0.88, 0.94)
        case .plan, .travel: ContentPose(320, 360, 0.80, 0)
        }
    }

    public static func worldOpacity(_ stage: OnboardingStage) -> Double {
        stage == .travel ? 0.18 : 1
    }

    public static func verticalDrift(for stage: OnboardingStage, shownAt target: OnboardingStage, distance: CGFloat) -> CGFloat {
        if stage == target { return 0 }
        return stage < target ? distance : -distance
    }

    public static func horizontalSlide(for stage: OnboardingStage, shownAt target: OnboardingStage) -> CGFloat {
        if stage == target { return 0 }
        return stage < target ? 40 : -40
    }

    public static let dragParallaxRatio: CGFloat = 0.35
    public static let dragLimit: CGFloat = 80
    public static let dragDeadZone: CGFloat = 4
    public static let swipeThreshold: CGFloat = 60
}
