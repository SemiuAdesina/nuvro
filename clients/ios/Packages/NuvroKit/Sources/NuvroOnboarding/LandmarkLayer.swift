import NuvroDesignSystem
import SwiftUI

struct LandmarkLayer: View {
    @Environment(\.designMetrics) private var metrics

    var body: some View {
        ZStack(alignment: .topLeading) {
            skyline.offset(x: metrics.x(70), y: metrics.fromTop(296))
            Mountains()
                .fill(LinearGradient(
                    stops: [
                        .init(color: Color(red: 0.965, green: 0.886, blue: 0.722), location: 0),
                        .init(color: Color(red: 0.541, green: 0.604, blue: 0.776), location: 0.26),
                        .init(color: Palette.twilight, location: 1)
                    ],
                    startPoint: .top,
                    endPoint: .bottom
                ))
                .frame(width: metrics.points(74), height: metrics.points(60))
                .offset(x: metrics.x(158), y: metrics.fromTop(290))
            monument.offset(x: metrics.x(262), y: metrics.fromTop(318))
            lagoon
            MapPin(color: Palette.cloud).frame(width: metrics.points(10), height: metrics.points(10)).offset(x: metrics.x(96), y: metrics.fromTop(274))
            MapPin(color: Palette.gold).frame(width: metrics.points(10), height: metrics.points(10)).offset(x: metrics.x(190), y: metrics.fromTop(270))
            MapPin(color: Palette.aqua).frame(width: metrics.points(10), height: metrics.points(10)).offset(x: metrics.x(334), y: metrics.fromTop(318))
        }
        .frame(width: metrics.size.width, height: metrics.size.height, alignment: .topLeading)
        .allowsHitTesting(false)
    }

    private var skyline: some View {
        HStack(alignment: .bottom, spacing: metrics.points(3)) {
            building(width: 12, height: 34, radius: 2, light: false)
            building(width: 15, height: 58, radius: 2, light: true)
            building(width: 11, height: 46, radius: 5, light: true, pale: true)
            building(width: 14, height: 28, radius: 2, light: false)
        }
    }

    private func building(width: CGFloat, height: CGFloat, radius: CGFloat, light: Bool, pale: Bool = false) -> some View {
        let top = pale ? Color(red: 0.541, green: 0.604, blue: 0.776) : light ? Color(red: 0.482, green: 0.549, blue: 0.733) : Color(red: 0.353, green: 0.431, blue: 0.639)
        let bottom = light ? Palette.twilight : Color(red: 0.180, green: 0.247, blue: 0.431)
        return UnevenRoundedRectangle(topLeadingRadius: metrics.points(radius), topTrailingRadius: metrics.points(radius))
            .fill(LinearGradient(colors: [top, bottom], startPoint: .top, endPoint: .bottom))
            .overlay(alignment: .trailing) {
                Rectangle().fill(Color.black.opacity(0.18)).frame(width: metrics.points(3))
            }
            .clipShape(UnevenRoundedRectangle(topLeadingRadius: metrics.points(radius), topTrailingRadius: metrics.points(radius)))
            .frame(width: metrics.points(width), height: metrics.points(height))
    }

    private var monument: some View {
        VStack(spacing: 0) {
            Circle().fill(Palette.gold).frame(width: metrics.points(6), height: metrics.points(6))
            UnevenRoundedRectangle(topLeadingRadius: metrics.points(16), topTrailingRadius: metrics.points(16))
                .fill(LinearGradient(colors: [Color(red: 0.953, green: 0.827, blue: 0.604), Color(red: 0.788, green: 0.604, blue: 0.357)], startPoint: .top, endPoint: .bottom))
                .frame(width: metrics.points(32), height: metrics.points(16))
            Colonnade()
                .frame(width: metrics.points(44), height: metrics.points(14))
        }
        .frame(width: metrics.points(44))
    }

    private var lagoon: some View {
        ZStack(alignment: .topLeading) {
            Ellipse()
                .fill(RadialGradient(
                    colors: [Color(red: 0.608, green: 0.918, blue: 0.894), Palette.aqua, Color(red: 0.247, green: 0.624, blue: 0.706)],
                    center: UnitPoint(x: 0.4, y: 0.4),
                    startRadius: 0,
                    endRadius: metrics.points(40)
                ))
                .frame(width: metrics.points(80), height: metrics.points(28))
                .offset(x: metrics.x(300), y: metrics.fromTop(344))
            Ellipse()
                .fill(LinearGradient(colors: [Color(red: 0.965, green: 0.886, blue: 0.722), Palette.gold], startPoint: .top, endPoint: .bottom))
                .frame(width: metrics.points(54), height: metrics.points(16))
                .offset(x: metrics.x(318), y: metrics.fromTop(340))
        }
    }
}

struct Mountains: Shape {
    func path(in rect: CGRect) -> Path {
        var path = Path()
        path.move(to: CGPoint(x: rect.minX, y: rect.maxY))
        path.addLine(to: CGPoint(x: rect.minX + rect.width * 0.38, y: rect.minY))
        path.addLine(to: CGPoint(x: rect.minX + rect.width * 0.62, y: rect.minY + rect.height * 0.42))
        path.addLine(to: CGPoint(x: rect.minX + rect.width * 0.78, y: rect.minY + rect.height * 0.20))
        path.addLine(to: CGPoint(x: rect.maxX, y: rect.maxY))
        path.closeSubpath()
        return path
    }
}

struct Colonnade: View {
    @Environment(\.designMetrics) private var metrics

    var body: some View {
        Canvas { context, size in
            let unit = size.width / 44
            var offset: CGFloat = 0
            while offset < size.width {
                context.fill(Path(CGRect(x: offset, y: 0, width: 4 * unit, height: size.height)), with: .color(Palette.cloud))
                context.fill(Path(CGRect(x: offset + 4 * unit, y: 0, width: 2 * unit, height: size.height)), with: .color(Color(red: 0.725, green: 0.682, blue: 0.604)))
                context.fill(Path(CGRect(x: offset + 6 * unit, y: 0, width: 2 * unit, height: size.height)), with: .color(Palette.cloud))
                offset += 8 * unit
            }
            context.fill(Path(CGRect(x: 0, y: 0, width: size.width, height: 2 * unit)), with: .color(Palette.gold))
        }
        .clipShape(UnevenRoundedRectangle(bottomLeadingRadius: metrics.points(2), bottomTrailingRadius: metrics.points(2)))
    }
}

struct MapPin: View {
    let color: Color

    var body: some View {
        GeometryReader { proxy in
            let radius = proxy.size.width / 2
            UnevenRoundedRectangle(topLeadingRadius: radius, bottomLeadingRadius: 0, bottomTrailingRadius: radius, topTrailingRadius: radius)
                .fill(color)
                .rotationEffect(.degrees(-45))
                .shadow(color: Palette.navy.opacity(0.4), radius: 3, y: 2)
        }
    }
}
