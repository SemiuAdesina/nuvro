import NuvroDesignSystem
import NuvroModels
import SwiftUI

struct DestinationHero: View {
    @Environment(\.designMetrics) private var metrics
    let match: DestinationMatch

    var body: some View {
        PhotoStandIn(palette: .named(match.photoReference), label: "photo · \(match.name.lowercased()), golden hour", labelAlignment: .leading)
            .frame(height: metrics.points(400))
            .overlay(
                LinearGradient(
                    stops: [
                        .init(color: Palette.navy.opacity(0.35), location: 0),
                        .init(color: Palette.navy.opacity(0), location: 0.30),
                        .init(color: Palette.navy.opacity(0), location: 0.55),
                        .init(color: Palette.background, location: 1)
                    ],
                    startPoint: .top, endPoint: .bottom
                )
            )
            .overlay(alignment: .bottom) {
                HStack(alignment: .bottom, spacing: metrics.points(12)) {
                    VStack(alignment: .leading, spacing: metrics.points(4)) {
                        Text("\(match.name),\n\(match.country)")
                            .typeStyle(Typography.largeTitle)
                            .foregroundStyle(Palette.label)
                            .lineSpacing(0)
                        Text("\(match.region) · \(match.name == "Madeira" ? "Funchal" : match.region)")
                            .typeStyle(Typography.bodyCompact)
                            .foregroundStyle(Palette.labelAlpha(0.65))
                    }
                    Spacer(minLength: 0)
                    MatchBadge(percent: match.score.percent)
                }
                .padding(.horizontal, metrics.points(20))
                .padding(.bottom, metrics.points(22))
            }
            .accessibilityElement(children: .contain)
    }
}
