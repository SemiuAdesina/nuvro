import NuvroDesignSystem
import SwiftUI

struct SignatureHeroCard: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let subtitle: String
    let action: String
    let onAction: () -> Void

    var body: some View {
        ZStack(alignment: .topLeading) {
            SkyGradient.heroDawn
            miniWorld
            VStack(alignment: .leading, spacing: metrics.points(8)) {
                Text(title).typeStyle(Typography.heroTitle).foregroundStyle(Palette.cloud)
                Text(subtitle).typeStyle(Typography.bodyCompact).foregroundStyle(Palette.cloudAlpha(0.82))
                Button(action: onAction) {
                    Text(action)
                        .typeStyle(Typography.control)
                        .foregroundStyle(Palette.cloud)
                        .padding(.horizontal, metrics.points(18))
                        .frame(height: metrics.points(40))
                }
                .buttonStyle(GlassPillButtonStyle(style: .dark))
                .padding(.top, metrics.points(10))
            }
            .frame(maxWidth: metrics.points(250), alignment: .leading)
            .padding(EdgeInsets(top: metrics.points(22), leading: metrics.points(22), bottom: metrics.points(20), trailing: metrics.points(22)))
        }
        .frame(maxWidth: .infinity, minHeight: metrics.points(214), alignment: .topLeading)
        .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.hero), style: .continuous))
        .accessibilityElement(children: .contain)
    }

    private var miniWorld: some View {
        GeometryReader { proxy in
            ZStack(alignment: .topLeading) {
                Circle()
                    .fill(SkyGradient.globe)
                    .overlay(Circle().strokeBorder(Color(red: 1, green: 0.925, blue: 0.745).opacity(0.75), lineWidth: metrics.points(3)).mask(LinearGradient(colors: [.white, .clear], startPoint: .top, endPoint: .center)))
                    .shadow(color: Color(red: 0.953, green: 0.827, blue: 0.604).opacity(0.35), radius: metrics.points(30), y: metrics.points(-20))
                    .frame(width: metrics.points(300), height: metrics.points(300))
                    .offset(x: proxy.size.width - metrics.points(170), y: proxy.size.height - metrics.points(120))
                Ellipse()
                    .fill(Color.white.opacity(0.55))
                    .frame(width: metrics.points(120), height: metrics.points(26))
                    .blur(radius: metrics.points(6))
                    .offset(x: proxy.size.width - metrics.points(160), y: metrics.points(40))
                HeroRoute()
                    .stroke(Color.white.opacity(0.6), style: StrokeStyle(lineWidth: metrics.points(1.2), dash: [metrics.points(3), metrics.points(5)]))
                    .frame(width: metrics.points(240), height: metrics.points(214))
                    .offset(x: proxy.size.width - metrics.points(240), y: metrics.points(10))
                Circle()
                    .fill(Color(red: 1, green: 0.969, blue: 0.902))
                    .frame(width: metrics.points(6), height: metrics.points(6))
                    .shadow(color: Color(red: 1, green: 0.925, blue: 0.745).opacity(0.9), radius: metrics.points(6))
                    .position(x: proxy.size.width - metrics.points(30), y: metrics.points(130))
                MapPinGlyph(color: Palette.gold)
                    .frame(width: metrics.points(10), height: metrics.points(10))
                    .position(x: proxy.size.width - metrics.points(70), y: proxy.size.height - metrics.points(46))
            }
        }
        .allowsHitTesting(false)
        .accessibilityHidden(true)
    }
}

struct HeroRoute: Shape {
    func path(in rect: CGRect) -> Path {
        let sx = rect.width / 240
        let sy = rect.height / 214
        var path = Path()
        path.move(to: CGPoint(x: rect.minX + 30 * sx, y: rect.minY + 150 * sy))
        path.addQuadCurve(to: CGPoint(x: rect.minX + 210 * sx, y: rect.minY + 120 * sy), control: CGPoint(x: rect.minX + 120 * sx, y: rect.minY + 40 * sy))
        return path
    }
}

struct MapPinGlyph: View {
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
