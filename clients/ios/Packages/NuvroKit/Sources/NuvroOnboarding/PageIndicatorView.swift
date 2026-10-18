import NuvroDesignSystem
import SwiftUI

struct PageIndicatorView: View {
    @Environment(\.designMetrics) private var metrics
    let stage: OnboardingStage
    let isShown: Bool

    var body: some View {
        HStack(spacing: metrics.points(7)) {
            ForEach(OnboardingStage.screens, id: \.rawValue) { screen in
                Circle()
                    .fill(screen == stage ? Palette.cloud : Palette.cloudAlpha(0.35))
                    .frame(width: metrics.points(6), height: metrics.points(6))
                    .animation(Motion.fade(Motion.indicatorDuration), value: stage)
            }
        }
        .padding(.horizontal, metrics.points(10))
        .padding(.vertical, metrics.points(7))
        .glassCapsule(.control)
        .position(x: metrics.horizontalCenter, y: metrics.aboveSafeBottom(150) - metrics.points(10))
        .opacity(isShown ? 1 : 0)
        .animation(Motion.fade(Motion.controlFadeDuration), value: isShown)
        .accessibilityElement(children: .ignore)
        .accessibilityLabel(indicatorLabel)
        .allowsHitTesting(false)
    }

    private var indicatorLabel: String {
        guard let page = stage.pageIndex else { return "" }
        return "Page \(page + 1) of \(OnboardingStage.screens.count)"
    }
}
