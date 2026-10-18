import NuvroDesignSystem
import SwiftUI

struct PlanningChipsView: View {
    @Environment(\.designMetrics) private var metrics
    let stage: OnboardingStage
    let reduceMotion: Bool

    private var isShown: Bool { stage == .plan }
    private var drift: CGFloat { reduceMotion ? 0 : StagePoses.verticalDrift(for: stage, shownAt: .plan, distance: 30) }

    var body: some View {
        ZStack(alignment: .topLeading) {
            PlanningConnectors()
                .stroke(Palette.cloudAlpha(0.55), style: StrokeStyle(lineWidth: metrics.points(1.2), dash: [metrics.points(3), metrics.points(5)]))
                .frame(width: metrics.points(402), height: metrics.points(560), alignment: .topLeading)
            Circle().fill(Palette.gold)
                .frame(width: metrics.points(8), height: metrics.points(8))
                .offset(x: metrics.x(196), y: metrics.fromTop(368))

            PlanningChip(spacing: 8) {
                Image(systemName: "airplane").font(.system(size: metrics.points(14))).foregroundStyle(Palette.sky)
            } title: { Text("$640") } detail: { Text("round trip") }
                .offset(x: metrics.x(22), y: metrics.fromTop(128))

            PlanningChip(spacing: 8) {
                Image(systemName: "sun.max.fill").font(.system(size: metrics.points(15))).foregroundStyle(Palette.sunlight)
            } title: { Text("24°") } detail: { Text("Oct · clear") }
                .offset(x: metrics.x(290), y: metrics.fromTop(124))

            PlanningChip(spacing: 9) {
                swatch(top: Palette.twilight, bottom: Palette.navy)
            } title: { Text("Bairro Alto Hotel") } detail: { Text("7 nights · $1,190") }
                .offset(x: metrics.x(26), y: metrics.fromTop(418))

            PlanningChip(spacing: 9) {
                swatch(top: Palette.gold, bottom: Palette.goldDeep)
            } title: { Text("Sintra day tour") } detail: { Text("Tue · 2 tickets") }
                .offset(x: metrics.x(246), y: metrics.fromTop(418))

            estimate
                .frame(width: metrics.size.width)
                .offset(y: metrics.fromTop(474))
        }
        .frame(width: metrics.size.width, height: metrics.size.height, alignment: .topLeading)
        .offset(y: metrics.points(drift))
        .opacity(isShown ? 1 : 0)
        .animation(reduceMotion ? nil : Motion.settle(Motion.cameraDuration), value: drift)
        .animation(Motion.fade(reduceMotion ? Motion.reducedWorldFadeDuration : Motion.worldFadeDuration), value: isShown)
        .allowsHitTesting(false)
    }

    private func swatch(top: Color, bottom: Color) -> some View {
        RoundedRectangle(cornerRadius: metrics.points(Radius.swatch), style: .continuous)
            .fill(LinearGradient(colors: [top, bottom], startPoint: .topLeading, endPoint: .bottomTrailing))
            .frame(width: metrics.points(22), height: metrics.points(22))
    }

    private var estimate: some View {
        HStack(alignment: .firstTextBaseline, spacing: metrics.points(6)) {
            Text("$2,480").typeStyle(Typography.estimate).foregroundStyle(Palette.cloud)
            Text("trip estimate").typeStyle(Typography.small).foregroundStyle(Palette.cloudAlpha(0.75))
        }
        .padding(.horizontal, metrics.points(16))
        .padding(.vertical, metrics.points(8))
        .glassCapsule(.dark)
    }
}

struct PlanningChip<Leading: View, Title: View, Detail: View>: View {
    @Environment(\.designMetrics) private var metrics
    let spacing: CGFloat
    @ViewBuilder let leading: Leading
    @ViewBuilder let title: Title
    @ViewBuilder let detail: Detail

    var body: some View {
        HStack(spacing: metrics.points(spacing)) {
            leading
            VStack(alignment: .leading, spacing: 0) {
                title.typeStyle(Typography.captionEmphasis).foregroundStyle(Palette.label)
                detail.typeStyle(Typography.micro).foregroundStyle(Palette.labelAlpha(0.6))
            }
        }
        .padding(.horizontal, metrics.points(12))
        .padding(.vertical, metrics.points(8))
        .background(Palette.surface, in: RoundedRectangle(cornerRadius: metrics.points(Radius.chip), style: .continuous))
        .shadow(color: Palette.navy.opacity(Shadow.chipOpacity), radius: metrics.points(Shadow.chipRadius), y: metrics.points(Shadow.chipY))
    }
}

struct PlanningConnectors: Shape {
    func path(in rect: CGRect) -> Path {
        let sx = rect.width / 402
        let sy = rect.height / 560
        func point(_ x: CGFloat, _ y: CGFloat) -> CGPoint { CGPoint(x: rect.minX + x * sx, y: rect.minY + y * sy) }
        var path = Path()
        path.move(to: point(200, 322)); path.addLine(to: point(200, 372))
        path.move(to: point(200, 372)); path.addLine(to: point(96, 412))
        path.move(to: point(200, 372)); path.addLine(to: point(304, 412))
        path.move(to: point(82, 166)); path.addLine(to: point(110, 190))
        path.move(to: point(322, 162)); path.addLine(to: point(296, 188))
        return path
    }
}
