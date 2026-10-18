import NuvroDesignSystem
import SwiftUI

struct CopyBlockView: View {
    @Environment(\.designMetrics) private var metrics
    let stage: OnboardingStage
    let dragOffset: CGFloat
    let reduceMotion: Bool
    let panels: [CopyPanel]

    var body: some View {
        ZStack(alignment: .bottomLeading) {
            ForEach(panels) { panel in
                panelView(panel)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .offset(x: metrics.points(slide(for: panel.stage)))
                    .opacity(stage == panel.stage ? 1 : 0)
                    .animation(reduceMotion ? nil : Motion.settle(Motion.textDuration), value: stage)
                    .animation(Motion.fade(Motion.textFadeDuration), value: stage)
            }
        }
        .frame(width: metrics.size.width - metrics.points(56), height: metrics.points(190), alignment: .bottomLeading)
        .position(x: metrics.horizontalCenter, y: metrics.aboveSafeBottom(184) - metrics.points(95))
        .allowsHitTesting(false)
    }

    private func slide(for target: OnboardingStage) -> CGFloat {
        if stage == target { return dragOffset * StagePoses.dragParallaxRatio }
        return reduceMotion ? 0 : StagePoses.horizontalSlide(for: stage, shownAt: target)
    }

    private func panelView(_ panel: CopyPanel) -> some View {
        VStack(alignment: .leading, spacing: metrics.points(10)) {
            Text(panel.headline).typeStyle(Typography.headline).foregroundStyle(Palette.cloud)
            if let supporting = panel.supporting {
                Text(supporting).typeStyle(Typography.supporting).foregroundStyle(Palette.gold)
            }
            Text(panel.body).typeStyle(Typography.body).foregroundStyle(Palette.cloudAlpha(0.78))
        }
        .multilineTextAlignment(.leading)
    }
}

struct CopyPanel: Identifiable, Equatable {
    let stage: OnboardingStage
    let headline: String
    let supporting: String?
    let body: String

    var id: Int { stage.rawValue }
}
