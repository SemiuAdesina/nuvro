import NuvroDesignSystem
import NuvroModels
import SwiftUI

struct HeroStayCard: View {
    @Environment(\.designMetrics) private var metrics
    let option: StayOption
    let model: StayModel
    let onView: () -> Void

    var body: some View {
        VStack(spacing: 0) {
            PhotoStandIn(palette: .named(option.photoReference), label: "photo · hotel terrace, funchal bay", labelAlignment: .topLeading)
                .frame(height: metrics.points(200))
                .overlay(alignment: .bottomLeading) {
                    Text(model.label(for: option.recommendation))
                        .typeStyle(Typography.statusPill).foregroundStyle(Palette.cloud)
                        .padding(.horizontal, metrics.points(12)).frame(height: metrics.points(30))
                        .background(Palette.navy.opacity(0.62), in: Capsule())
                        .overlay(Capsule().strokeBorder(Palette.cloudAlpha(0.35), lineWidth: 0.5))
                        .padding(metrics.points(14))
                }
                .overlay(alignment: .topTrailing) {
                    SaveButton(isSaved: model.isSaved(option), style: .control) { model.toggleSaved(option) }.padding(metrics.points(14))
                }
            VStack(alignment: .leading, spacing: metrics.points(13)) {
                HStack(alignment: .top, spacing: metrics.points(12)) {
                    VStack(alignment: .leading, spacing: metrics.points(3)) {
                        Text(option.name).typeStyle(TypeStyle(size: 19, weight: .bold, tracking: -0.3)).foregroundStyle(Palette.label)
                        Text("\(option.area) · \(option.distanceNote)").typeStyle(Typography.meta).foregroundStyle(Palette.labelAlpha(0.6))
                    }
                    Spacer(minLength: 0)
                    VStack(alignment: .trailing, spacing: metrics.points(1)) {
                        Text(model.nightlyLabel(option)).typeStyle(TypeStyle(size: 17, weight: .bold, tracking: -0.3)).foregroundStyle(Palette.label)
                        Text(model.copy.perNight).typeStyle(Typography.small).foregroundStyle(Palette.labelAlpha(0.55))
                    }
                }
                HStack(spacing: metrics.points(8)) {
                    FactTile(key: "\(option.nights) \(model.copy.nights)", value: model.totalLabel(option))
                    FactTile(key: model.copy.rating, value: "\(option.rating.formatted(.number.precision(.fractionLength(1)))) ★")
                    FactTile(key: model.copy.toPlans, value: "~\(option.minutesToPlans) min")
                }
                VStack(alignment: .leading, spacing: metrics.points(7)) {
                    Divider().overlay(Palette.labelAlpha(0.08))
                    Text(model.copy.whyPicked).typeStyle(TypeStyle(size: 14, weight: .bold)).foregroundStyle(Palette.label).padding(.top, metrics.points(8))
                    ForEach(option.reasons, id: \.self) { CheckRow($0, size: 13.5) }
                }
                HStack(spacing: metrics.points(10)) {
                    CallToActionButton(model.copy.viewStay, action: onView)
                    Button { model.toggleSaved(option) } label: {
                        Text(model.isSaved(option) ? model.copy.saved : model.copy.save)
                            .typeStyle(Typography.secondaryAction).foregroundStyle(Palette.label)
                            .padding(.horizontal, metrics.points(20)).frame(height: metrics.points(54))
                            .background(Palette.labelAlpha(0.06), in: Capsule())
                            .overlay(Capsule().strokeBorder(Palette.labelAlpha(0.1), lineWidth: 0.5))
                    }
                    .buttonStyle(PressScaleButtonStyle(scale: 0.97))
                }
                .padding(.top, metrics.points(2))
            }
            .padding(EdgeInsets(top: metrics.points(14), leading: metrics.points(16), bottom: metrics.points(16), trailing: metrics.points(16)))
        }
        .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.hero), style: .continuous))
        .surfaceCard(radius: Radius.hero, elevated: true)
    }
}

struct StayRow: View {
    @Environment(\.designMetrics) private var metrics
    let option: StayOption
    let model: StayModel
    let onSelect: () -> Void

    private var labelColour: Color {
        switch option.recommendation {
        case .bestValue: Palette.matchTeal
        case .bestLocation: Palette.sky
        case .bestPremium: Palette.amber
        case .bestOverall: Palette.label
        }
    }

    var body: some View {
        Button(action: onSelect) {
            HStack(spacing: metrics.points(12)) {
                PhotoStandIn(palette: .named(option.photoReference), label: "", labelSize: 8)
                    .frame(width: metrics.points(78), height: metrics.points(70))
                    .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.thumbnail), style: .continuous))
                VStack(alignment: .leading, spacing: metrics.points(3)) {
                    Text(model.label(for: option.recommendation).uppercased()).typeStyle(TypeStyle(size: 11, weight: .bold, tracking: 0.22)).foregroundStyle(labelColour)
                    Text(option.name).typeStyle(Typography.cardTitle).foregroundStyle(Palette.label)
                    Text("\(option.area) · \(option.distanceNote)").typeStyle(Typography.caption).foregroundStyle(Palette.labelAlpha(0.6)).lineLimit(1)
                }
                Spacer(minLength: 0)
                VStack(alignment: .trailing, spacing: metrics.points(1)) {
                    Text(model.nightlyLabel(option)).typeStyle(TypeStyle(size: 15, weight: .bold)).foregroundStyle(Palette.label)
                    Text("/ night").typeStyle(Typography.small).foregroundStyle(Palette.labelAlpha(0.55))
                }
            }
            .padding(metrics.points(10))
            .surfaceCard(radius: Radius.cardCompact)
        }
        .buttonStyle(PressScaleButtonStyle())
        .accessibilityElement(children: .combine)
    }
}
