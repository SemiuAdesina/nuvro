import NuvroDesignSystem
import NuvroLocalization
import SwiftUI

struct LiveTripView: View {
    @Environment(\.designMetrics) private var metrics
    let stage: OnboardingStage
    let hasReplanned: Bool
    let reduceMotion: Bool
    let copy: OnboardingCopy

    private var isShown: Bool { stage == .travel }

    var body: some View {
        ZStack(alignment: .topLeading) {
            LiveTripMap(
                hasReplanned: hasReplanned,
                reduceMotion: reduceMotion,
                youLabel: copy.you,
                lunchLabel: copy.lunch,
                outdoorLabel: copy.outdoor,
                indoorLabel: copy.indoor
            )
            .offset(x: metrics.x(30), y: metrics.fromTop(96))

            weatherChange.offset(x: metrics.x(18), y: metrics.fromTop(118))

            notice
                .frame(width: metrics.size.width)
                .offset(y: metrics.fromTop(432) + (hasReplanned || reduceMotion ? 0 : metrics.points(12)))
                .opacity(hasReplanned ? 1 : 0)
                .animation(Motion.settle(0.6), value: hasReplanned)

            translationCard
                .offset(x: metrics.x(184), y: metrics.fromTop(372))
                .opacity(hasReplanned ? 1 : 0)
                .animation(Motion.fade(0.5).delay(0.3), value: hasReplanned)
        }
        .frame(width: metrics.size.width, height: metrics.size.height, alignment: .topLeading)
        .scaleEffect(isShown || reduceMotion ? 1 : 0.96)
        .offset(y: isShown || reduceMotion ? 0 : metrics.points(60))
        .opacity(isShown ? 1 : 0)
        .animation(reduceMotion ? nil : Motion.settle(Motion.cameraDuration), value: isShown)
        .animation(Motion.fade(reduceMotion ? Motion.reducedWorldFadeDuration : Motion.worldFadeDuration), value: isShown)
        .allowsHitTesting(false)
    }

    private var weatherChange: some View {
        HStack(spacing: metrics.points(6)) {
            Image(systemName: "sun.max.fill").foregroundStyle(Palette.gold)
            Image(systemName: "arrow.forward").opacity(0.6)
            Image(systemName: "cloud.rain.fill").foregroundStyle(Palette.rainBlue)
            Text("13:00")
        }
        .typeStyle(Typography.captionEmphasis)
        .foregroundStyle(Palette.cloud)
        .padding(.horizontal, metrics.points(11))
        .padding(.vertical, metrics.points(7))
        .glassCapsule(.dark)
    }

    private var notice: some View {
        HStack(spacing: metrics.points(10)) {
            Circle().fill(Palette.aqua)
                .frame(width: metrics.points(8), height: metrics.points(8))
                .shadow(color: Palette.aqua, radius: metrics.points(4))
            VStack(alignment: .leading, spacing: 0) {
                Text(copy.plansUpdated).typeStyle(TypeStyle(size: 13, weight: .semibold)).foregroundStyle(Palette.cloud)
                Text(copy.plansUpdatedDetail).typeStyle(Typography.small).foregroundStyle(Palette.cloudAlpha(0.75))
            }
        }
        .padding(.horizontal, metrics.points(14))
        .padding(.vertical, metrics.points(10))
        .glass(.notice, in: RoundedRectangle(cornerRadius: metrics.points(Radius.notice), style: .continuous))
        .shadow(color: Palette.navy.opacity(0.3), radius: metrics.points(15), y: metrics.points(10))
    }

    private var translationCard: some View {
        VStack(alignment: .leading, spacing: metrics.points(2)) {
            Text("Uma mesa para dois, por favor").typeStyle(Typography.captionEmphasis).foregroundStyle(Palette.label)
            Text(copy.translation).typeStyle(Typography.micro).foregroundStyle(Palette.labelAlpha(0.6))
        }
        .padding(.horizontal, metrics.points(12))
        .padding(.vertical, metrics.points(9))
        .background(Palette.surface, in: RoundedRectangle(cornerRadius: metrics.points(Radius.chip), style: .continuous))
        .shadow(color: Palette.navy.opacity(Shadow.chipOpacity), radius: metrics.points(Shadow.chipRadius), y: metrics.points(Shadow.chipY))
    }
}
