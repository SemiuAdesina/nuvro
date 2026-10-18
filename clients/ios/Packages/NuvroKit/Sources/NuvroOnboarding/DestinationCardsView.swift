import NuvroDesignSystem
import SwiftUI

struct DestinationCardsView: View {
    @Environment(\.designMetrics) private var metrics
    let stage: OnboardingStage
    let reduceMotion: Bool
    let matchLabel: String

    var body: some View {
        ZStack(alignment: .topLeading) {
            posed(StagePoses.primaryCard(stage), delay: 0) {
                PrimaryDestinationCard(matchLabel: matchLabel).frame(width: metrics.points(200))
            }
            posed(StagePoses.secondaryCard(stage), delay: 0.08) {
                CompactDestinationCard(palette: .kyoto, label: "photo · kyoto", name: "Japan", score: "89%").frame(width: metrics.points(150))
            }
            posed(StagePoses.tertiaryCard(stage), delay: 0.16) {
                CompactDestinationCard(palette: .rainforest, label: "photo · rainforest", name: "Costa Rica", score: "87%").frame(width: metrics.points(150))
            }
        }
        .frame(width: metrics.size.width, height: metrics.size.height, alignment: .topLeading)
        .allowsHitTesting(false)
    }

    private func posed<Content: View>(_ pose: ContentPose, delay: TimeInterval, @ViewBuilder content: () -> Content) -> some View {
        content()
            .scaleEffect(pose.scale)
            .offset(x: metrics.x(pose.x), y: metrics.fromTop(pose.y))
            .opacity(pose.opacity)
            .animation(reduceMotion ? nil : Motion.settle(Motion.cameraDuration).delay(delay), value: pose.x)
            .animation(reduceMotion ? nil : Motion.settle(Motion.cameraDuration).delay(delay), value: pose.y)
            .animation(reduceMotion ? nil : Motion.settle(Motion.cameraDuration).delay(delay), value: pose.scale)
            .animation(Motion.fade(reduceMotion ? Motion.reducedWorldFadeDuration : Motion.worldFadeDuration).delay(delay), value: pose.opacity)
    }
}

struct PrimaryDestinationCard: View {
    @Environment(\.designMetrics) private var metrics
    let matchLabel: String

    var body: some View {
        VStack(spacing: 0) {
            PhotoStandIn(palette: .lisbon, label: "photo · lisbon coast", labelSize: 9)
                .frame(height: metrics.points(104))
            HStack {
                VStack(alignment: .leading, spacing: metrics.points(2)) {
                    Text("Portugal").typeStyle(Typography.cardTitle).foregroundStyle(Palette.label)
                    Text("Lisbon · 7 nights").typeStyle(Typography.caption).foregroundStyle(Palette.labelAlpha(0.6))
                }
                Spacer(minLength: 0)
                Text("94% \(matchLabel)")
                    .typeStyle(Typography.captionEmphasis)
                    .foregroundStyle(Palette.matchTeal)
                    .padding(.horizontal, metrics.points(9))
                    .padding(.vertical, metrics.points(5))
                    .background(Palette.aqua.opacity(0.18), in: Capsule())
            }
            .padding(EdgeInsets(top: metrics.points(10), leading: metrics.points(12), bottom: metrics.points(11), trailing: metrics.points(12)))
        }
        .background(Palette.surface)
        .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.card), style: .continuous))
        .shadow(color: Palette.navy.opacity(Shadow.cardOpacity), radius: metrics.points(Shadow.cardRadius), y: metrics.points(Shadow.cardY))
        .shadow(color: Palette.labelAlpha(0.15), radius: metrics.points(3), y: metrics.points(2))
    }
}

struct CompactDestinationCard: View {
    @Environment(\.designMetrics) private var metrics
    let palette: PhotoPalette
    let label: String
    let name: String
    let score: String

    var body: some View {
        VStack(spacing: 0) {
            PhotoStandIn(palette: palette, label: label, labelSize: 8)
                .frame(height: metrics.points(70))
            HStack {
                Text(name).typeStyle(Typography.cardTitleCompact).foregroundStyle(Palette.label)
                Spacer(minLength: 0)
                Text(score).typeStyle(Typography.smallEmphasis).foregroundStyle(Palette.matchTeal)
            }
            .padding(EdgeInsets(top: metrics.points(8), leading: metrics.points(10), bottom: metrics.points(9), trailing: metrics.points(10)))
        }
        .background(Palette.surface)
        .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.cardCompact), style: .continuous))
        .shadow(color: Palette.navy.opacity(0.32), radius: metrics.points(20), y: metrics.points(18))
    }
}
