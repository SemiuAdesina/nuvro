import NuvroDesignSystem
import SwiftUI

struct SkyLayer: View {
    @Environment(\.designMetrics) private var metrics

    var body: some View {
        ZStack(alignment: .topLeading) {
            Circle()
                .fill(SkyGradient.sunGlow)
                .frame(width: metrics.points(320), height: metrics.points(320))
                .offset(x: metrics.x(150), y: metrics.fromTop(110))

            cloud(x: 26, y: 168, width: 150, height: 34, opacity: 0.62, blur: 7)
            cloud(x: 236, y: 118, width: 120, height: 28, opacity: 0.50, blur: 6)
            cloud(x: 120, y: 240, width: 96, height: 22, opacity: 0.55, blur: 5)
            cloud(x: 284, y: 290, width: 150, height: 36, opacity: 0.55, blur: 7, warm: true)
        }
        .frame(width: metrics.size.width, height: metrics.size.height, alignment: .topLeading)
        .allowsHitTesting(false)
    }

    private func cloud(x: CGFloat, y: CGFloat, width: CGFloat, height: CGFloat, opacity: Double, blur: CGFloat, warm: Bool = false) -> some View {
        Ellipse()
            .fill((warm ? Color(red: 1, green: 0.965, blue: 0.902) : Color.white).opacity(opacity))
            .frame(width: metrics.points(width), height: metrics.points(height))
            .blur(radius: metrics.points(blur))
            .offset(x: metrics.x(x), y: metrics.fromTop(y))
    }
}
