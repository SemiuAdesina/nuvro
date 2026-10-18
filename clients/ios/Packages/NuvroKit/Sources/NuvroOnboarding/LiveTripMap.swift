import NuvroDesignSystem
import SwiftUI

struct LiveTripRoute: Shape {
    static let designSize = CGSize(width: 342, height: 330)

    func path(in rect: CGRect) -> Path {
        let sx = rect.width / Self.designSize.width
        let sy = rect.height / Self.designSize.height
        func point(_ x: CGFloat, _ y: CGFloat) -> CGPoint { CGPoint(x: rect.minX + x * sx, y: rect.minY + y * sy) }
        var path = Path()
        path.move(to: point(62, 282))
        path.addCurve(to: point(118, 176), control1: point(90, 240), control2: point(70, 200))
        path.addCurve(to: point(212, 100), control1: point(166, 152), control2: point(200, 150))
        return path
    }
}

struct LiveTripMap: View {
    @Environment(\.designMetrics) private var metrics
    let hasReplanned: Bool
    let reduceMotion: Bool
    let youLabel: String
    let lunchLabel: String
    let outdoorLabel: String
    let indoorLabel: String

    private var mapSize: CGSize { metrics.size(342, 330) }

    var body: some View {
        ZStack(alignment: .topLeading) {
            MapGrid(spacing: metrics.points(28)).stroke(Palette.twilight.opacity(0.08), lineWidth: 1)
            Rectangle().fill(Palette.twilight.opacity(0.10))
                .frame(width: mapSize.width, height: metrics.points(22))
                .rotationEffect(.degrees(-8))
                .offset(y: metrics.points(140))
            Rectangle().fill(Palette.twilight.opacity(0.10))
                .frame(width: metrics.points(18), height: mapSize.height)
                .rotationEffect(.degrees(12))
                .offset(x: metrics.points(210))
            LinearGradient(colors: [Palette.aqua.opacity(0), Palette.aqua.opacity(0.35)], startPoint: .top, endPoint: .bottom)
                .frame(height: metrics.points(70))
                .offset(y: mapSize.height - metrics.points(70))
            LiveTripRoute().stroke(Palette.sky.opacity(0.25), style: StrokeStyle(lineWidth: metrics.points(6), lineCap: .round))
            LiveTripRoute()
                .trim(from: 0, to: hasReplanned ? 1 : 0)
                .stroke(Palette.sky, style: StrokeStyle(lineWidth: metrics.points(2.5), lineCap: .round))
                .animation(Motion.routeDraw, value: hasReplanned)
            youMarker.offset(x: metrics.points(52), y: metrics.points(272))
            Text(youLabel).typeStyle(Typography.smallEmphasis).foregroundStyle(Palette.label)
                .offset(x: metrics.points(78), y: metrics.points(274))
            ItineraryStop(dot: Palette.gold, name: "Cervejaria Ramiro", time: "13:15 · \(lunchLabel)", highlighted: false)
                .offset(x: metrics.points(126), y: metrics.points(150))
            ItineraryStop(dot: Palette.twilight, name: "Miradouro da Graça", time: "10:30 · \(outdoorLabel)", highlighted: false)
                .offset(x: metrics.points(200), y: metrics.points(64) + (hasReplanned && !reduceMotion ? metrics.points(-18) : 0))
                .opacity(hasReplanned ? 0 : 1)
                .animation(Motion.swap(reduceMotion: reduceMotion), value: hasReplanned)
            ItineraryStop(dot: Palette.sky, name: "MAAT Museum", time: "10:30 · \(indoorLabel)", highlighted: true)
                .offset(x: metrics.points(200), y: metrics.points(64) + (hasReplanned || reduceMotion ? 0 : metrics.points(18)))
                .opacity(hasReplanned ? 1 : 0)
                .animation(Motion.swap(reduceMotion: reduceMotion).delay(Motion.swapEntryDelay), value: hasReplanned)
        }
        .frame(width: mapSize.width, height: mapSize.height, alignment: .topLeading)
        .background(Palette.background)
        .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.map), style: .continuous))
        .shadow(color: Palette.navy.opacity(Shadow.mapOpacity), radius: metrics.points(Shadow.mapRadius), y: metrics.points(Shadow.mapY))
    }

    private var youMarker: some View {
        ZStack {
            PulseRing(active: !reduceMotion).frame(width: metrics.points(20), height: metrics.points(20))
            Circle().fill(Palette.sky)
                .overlay(Circle().strokeBorder(Color.white, lineWidth: metrics.points(2.5)))
                .frame(width: metrics.points(12), height: metrics.points(12))
                .shadow(color: Palette.navy.opacity(0.35), radius: metrics.points(3), y: metrics.points(2))
        }
        .frame(width: metrics.points(20), height: metrics.points(20))
    }
}

struct MapGrid: Shape {
    let spacing: CGFloat

    func path(in rect: CGRect) -> Path {
        var path = Path()
        var x = rect.minX
        while x <= rect.maxX {
            path.move(to: CGPoint(x: x, y: rect.minY)); path.addLine(to: CGPoint(x: x, y: rect.maxY)); x += spacing
        }
        var y = rect.minY
        while y <= rect.maxY {
            path.move(to: CGPoint(x: rect.minX, y: y)); path.addLine(to: CGPoint(x: rect.maxX, y: y)); y += spacing
        }
        return path
    }
}

struct PulseRing: View {
    let active: Bool
    @State private var expanded = false

    var body: some View {
        Circle()
            .fill(Palette.sky.opacity(0.35))
            .scaleEffect(expanded ? 2.2 : 1)
            .opacity(expanded ? 0 : 0.9)
            .onAppear { if active { withAnimation(Motion.pulse) { expanded = true } } }
    }
}

struct ItineraryStop: View {
    @Environment(\.designMetrics) private var metrics
    let dot: Color
    let name: String
    let time: String
    let highlighted: Bool

    var body: some View {
        HStack(spacing: metrics.points(8)) {
            Circle().fill(dot).frame(width: metrics.points(8), height: metrics.points(8))
            VStack(alignment: .leading, spacing: 0) {
                Text(name).typeStyle(Typography.smallEmphasis).foregroundStyle(Palette.label)
                Text(time).typeStyle(Typography.micro).foregroundStyle(highlighted ? Palette.sky : Palette.labelAlpha(0.6))
            }
        }
        .padding(.horizontal, metrics.points(10))
        .padding(.vertical, metrics.points(6))
        .background(Color.white, in: RoundedRectangle(cornerRadius: metrics.points(Radius.stop), style: .continuous))
        .overlay {
            if highlighted {
                RoundedRectangle(cornerRadius: metrics.points(Radius.stop), style: .continuous).strokeBorder(Palette.sky.opacity(0.5), lineWidth: 1)
            }
        }
        .shadow(color: Palette.navy.opacity(Shadow.stopOpacity), radius: metrics.points(Shadow.stopRadius), y: metrics.points(Shadow.stopY))
    }
}
