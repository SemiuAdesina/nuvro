import SwiftUI

public enum Motion {
    public static let cameraDuration: TimeInterval = 1.0
    public static let textDuration: TimeInterval = 0.6
    public static let worldFadeDuration: TimeInterval = 0.6
    public static let textFadeDuration: TimeInterval = 0.45
    public static let reducedWorldFadeDuration: TimeInterval = 0.55
    public static let reducedTextFadeDuration: TimeInterval = 0.45
    public static let swapDuration: TimeInterval = 0.7
    public static let swapFadeDuration: TimeInterval = 0.5
    public static let swapEntryDelay: TimeInterval = 0.25
    public static let routeDrawDuration: TimeInterval = 1.2
    public static let pulseDuration: TimeInterval = 1.8
    public static let planeDuration: TimeInterval = 1.6
    public static let planeDelay: TimeInterval = 0.3
    public static let indicatorDuration: TimeInterval = 0.3
    public static let controlFadeDuration: TimeInterval = 0.4

    public static let splashHold: TimeInterval = 1.6
    public static let replanDelay: TimeInterval = 1.1

    public static func settle(_ duration: TimeInterval) -> Animation {
        .timingCurve(0.2, 0.8, 0.2, 1, duration: duration)
    }

    public static func fade(_ duration: TimeInterval) -> Animation {
        .easeInOut(duration: duration)
    }

    public static func camera(reduceMotion: Bool) -> Animation {
        reduceMotion ? fade(reducedWorldFadeDuration) : settle(cameraDuration)
    }

    public static func text(reduceMotion: Bool) -> Animation {
        reduceMotion ? fade(reducedTextFadeDuration) : settle(textDuration)
    }

    public static func swap(reduceMotion: Bool) -> Animation {
        reduceMotion ? fade(swapFadeDuration) : settle(swapDuration)
    }

    public static let pulse: Animation = .easeOut(duration: pulseDuration).repeatForever(autoreverses: false)
    public static let routeDraw: Animation = .easeInOut(duration: routeDrawDuration)
    public static let plane: Animation = .easeInOut(duration: planeDuration).delay(planeDelay)
}
