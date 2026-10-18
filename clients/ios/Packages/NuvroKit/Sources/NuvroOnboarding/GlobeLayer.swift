import NuvroDesignSystem
import SwiftUI

struct GlobeLayer: View {
    @Environment(\.designMetrics) private var metrics
    let lightProgress: CGFloat

    var body: some View {
        ZStack(alignment: .topLeading) {
            globe
            landMass(x: 60, y: 392, width: 150, height: 60, color: Palette.gold.opacity(0.28), blur: 4)
            landMass(x: 230, y: 405, width: 120, height: 48, color: Palette.gold.opacity(0.22), blur: 4)
            landMass(x: 150, y: 470, width: 200, height: 70, color: Palette.aqua.opacity(0.18), blur: 6)
            route
        }
        .frame(width: metrics.size.width, height: metrics.size.height, alignment: .topLeading)
        .allowsHitTesting(false)
    }

    private var globe: some View {
        let diameter = metrics.points(682)
        return Circle()
            .fill(SkyGradient.globe)
            .overlay {
                Circle().strokeBorder(
                    LinearGradient(
                        colors: [Color(red: 1, green: 0.925, blue: 0.745).opacity(0.75), .clear],
                        startPoint: .top,
                        endPoint: UnitPoint(x: 0.5, y: 0.12)
                    ),
                    lineWidth: metrics.points(3)
                )
            }
            .overlay {
                Circle().fill(
                    RadialGradient(
                        colors: [Color(red: 1, green: 0.839, blue: 0.588).opacity(0.22), .clear],
                        center: UnitPoint(x: 0.5, y: 0.03),
                        startRadius: 0,
                        endRadius: metrics.points(120)
                    )
                )
            }
            .shadow(color: Color(red: 0.953, green: 0.827, blue: 0.604).opacity(0.35), radius: metrics.points(40), y: metrics.points(-20))
            .frame(width: diameter, height: diameter)
            .offset(x: metrics.x(-140), y: metrics.fromTop(342))
    }

    private func landMass(x: CGFloat, y: CGFloat, width: CGFloat, height: CGFloat, color: Color, blur: CGFloat) -> some View {
        Ellipse()
            .fill(color)
            .frame(width: metrics.points(width), height: metrics.points(height))
            .blur(radius: metrics.points(blur))
            .offset(x: metrics.x(x), y: metrics.fromTop(y))
    }

    private var route: some View {
        let size = CGSize(width: metrics.points(402), height: metrics.points(560))
        return ZStack(alignment: .topLeading) {
            GlobeRoute()
                .stroke(
                    Color.white.opacity(0.55),
                    style: StrokeStyle(lineWidth: metrics.points(1.2), dash: [metrics.points(3), metrics.points(5)])
                )
            TravellingLight(progress: lightProgress)
                .fill(Color(red: 1, green: 0.969, blue: 0.902))
                .shadow(color: Color(red: 1, green: 0.925, blue: 0.745).opacity(0.9), radius: metrics.points(6))
        }
        .frame(width: size.width, height: size.height, alignment: .topLeading)
    }
}
