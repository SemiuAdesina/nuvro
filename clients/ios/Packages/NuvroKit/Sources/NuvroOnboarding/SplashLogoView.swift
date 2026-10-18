import NuvroDesignSystem
import SwiftUI

struct SplashLogoView: View {
    @Environment(\.designMetrics) private var metrics
    let stage: OnboardingStage
    let hasBooted: Bool
    let reduceMotion: Bool
    let tagline: String

    var body: some View {
        VStack(spacing: metrics.points(14)) {
            Wordmark(size: metrics.points(46))
            Text(tagline)
                .typeStyle(Typography.tagline)
                .foregroundStyle(Palette.cloudAlpha(0.85))
        }
        .frame(width: metrics.size.width)
        .scaleEffect(pose.scale)
        .offset(y: metrics.fromTop(168) + metrics.points(pose.rise))
        .opacity(pose.opacity)
        .animation(reduceMotion ? nil : Motion.settle(Motion.cameraDuration), value: pose.scale)
        .animation(reduceMotion ? nil : Motion.settle(Motion.cameraDuration), value: pose.rise)
        .animation(Motion.fade(Motion.worldFadeDuration), value: pose.opacity)
        .allowsHitTesting(false)
    }

    private struct Pose: Equatable {
        let scale: CGFloat
        let rise: CGFloat
        let opacity: Double
    }

    private var pose: Pose {
        switch (stage, hasBooted) {
        case (.splash, false): Pose(scale: 0.96, rise: 8, opacity: 0)
        case (.splash, true): Pose(scale: 1, rise: 0, opacity: 1)
        default: reduceMotion ? Pose(scale: 1, rise: 0, opacity: 0) : Pose(scale: 1.25, rise: -30, opacity: 0)
        }
    }
}
