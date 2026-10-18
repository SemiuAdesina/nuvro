import NuvroDesignSystem
import NuvroModels
import SwiftUI

struct FactsStrip: View {
    @Environment(\.designMetrics) private var metrics
    let facts: [DestinationFact]

    var body: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: metrics.points(8)) {
                ForEach(facts) { fact in
                    VStack(alignment: .leading, spacing: metrics.points(2)) {
                        Text(fact.key).typeStyle(Typography.factKey).foregroundStyle(Palette.labelAlpha(0.55))
                        Text(fact.value).typeStyle(Typography.bodyCompactEmphasis).foregroundStyle(Palette.label)
                    }
                    .padding(.horizontal, metrics.points(13))
                    .padding(.vertical, metrics.points(9))
                    .surfaceCard(radius: Radius.chip)
                }
            }
            .padding(.horizontal, metrics.points(20))
            .padding(.vertical, metrics.points(6))
        }
        .scrollClipDisabled()
    }
}

struct WhyPickedCard: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let summary: String
    let reasons: [DestinationFact]

    private var columns: [GridItem] { [GridItem(.flexible(), spacing: metrics.points(8)), GridItem(.flexible(), spacing: metrics.points(8))] }

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(10)) {
            HStack(spacing: metrics.points(8)) {
                RouteArcGlyph(width: metrics.points(16))
                Text(title).typeStyle(TypeStyle(size: 15, weight: .bold)).foregroundStyle(Palette.label)
            }
            Text(summary).typeStyle(Typography.bodyCompact).foregroundStyle(Palette.labelAlpha(0.75))
            LazyVGrid(columns: columns, spacing: metrics.points(8)) {
                ForEach(reasons) { reason in
                    VStack(alignment: .leading, spacing: metrics.points(1)) {
                        Text(reason.key).typeStyle(Typography.small).foregroundStyle(Palette.labelAlpha(0.55))
                        Text(reason.value).typeStyle(Typography.metaEmphasis).foregroundStyle(Palette.label)
                    }
                    .padding(.horizontal, metrics.points(11))
                    .padding(.vertical, metrics.points(9))
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(Palette.surface.opacity(0.8), in: RoundedRectangle(cornerRadius: metrics.points(Radius.stop), style: .continuous))
                }
            }
            .padding(.top, metrics.points(2))
        }
        .padding(metrics.points(16))
        .background(
            LinearGradient(colors: [Palette.sky.opacity(0.08), Palette.aqua.opacity(0.10)], startPoint: .topLeading, endPoint: .bottomTrailing),
            in: RoundedRectangle(cornerRadius: metrics.points(Radius.card), style: .continuous)
        )
        .overlay(RoundedRectangle(cornerRadius: metrics.points(Radius.card), style: .continuous).strokeBorder(Palette.sky.opacity(0.18), lineWidth: 0.5))
    }
}

struct CategoryGrid: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let onSelect: (GuideCategory) -> Void

    private var rows: [GridItem] { [GridItem(.fixed(metrics.points(46)), spacing: metrics.points(8)), GridItem(.fixed(metrics.points(46)), spacing: metrics.points(8))] }

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(12)) {
            Text(title).typeStyle(Typography.title).foregroundStyle(Palette.label).padding(.horizontal, metrics.points(20))
            ScrollView(.horizontal, showsIndicators: false) {
                LazyHGrid(rows: rows, spacing: metrics.points(8)) {
                    ForEach(Array(GuideCategory.allCases.enumerated()), id: \.element) { index, category in
                        Button { onSelect(category) } label: {
                            HStack(spacing: metrics.points(9)) {
                                IconTile(palette: PhotoPalette.iconTiles[index % 4], systemImage: category.systemImage)
                                Text(category.title).typeStyle(Typography.metaEmphasis).foregroundStyle(Palette.label).lineLimit(1)
                            }
                            .padding(.leading, metrics.points(8))
                            .padding(.trailing, metrics.points(12))
                            .frame(height: metrics.points(46))
                            .surfaceCard(radius: Radius.chip)
                        }
                        .buttonStyle(PressScaleButtonStyle(scale: 0.97))
                    }
                }
                .padding(.horizontal, metrics.points(20))
                .padding(.vertical, metrics.points(4))
            }
            .scrollClipDisabled()
        }
    }
}

struct AreasRow: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let seeAll: String
    let areas: [StayArea]

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(12)) {
            SectionHeader(title, action: seeAll, onAction: {}).padding(.horizontal, metrics.points(20))
            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: metrics.points(10)) {
                    ForEach(areas) { area in
                        VStack(alignment: .leading, spacing: 0) {
                            PhotoStandIn(palette: .named(area.photoReference), label: "", labelSize: 8)
                                .frame(height: metrics.points(96))
                                .overlay(alignment: .topLeading) {
                                    Text(area.highlight)
                                        .typeStyle(Typography.microEmphasis)
                                        .foregroundStyle(Palette.cloud)
                                        .padding(.horizontal, metrics.points(8))
                                        .padding(.vertical, metrics.points(4))
                                        .glassCapsule(.dark)
                                        .padding(metrics.points(10))
                                }
                            VStack(alignment: .leading, spacing: metrics.points(2)) {
                                Text(area.name).typeStyle(Typography.cardTitleCompact).foregroundStyle(Palette.label)
                                Text(area.summary).typeStyle(Typography.small).foregroundStyle(Palette.labelAlpha(0.6)).lineLimit(1)
                            }
                            .padding(EdgeInsets(top: metrics.points(10), leading: metrics.points(12), bottom: metrics.points(12), trailing: metrics.points(12)))
                        }
                        .frame(width: metrics.points(156))
                        .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.cardCompact), style: .continuous))
                        .surfaceCard(radius: Radius.cardCompact)
                    }
                }
                .padding(.horizontal, metrics.points(20))
                .padding(.vertical, metrics.points(4))
            }
            .scrollClipDisabled()
        }
    }
}

struct ActivitiesList: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let seeAll: String
    let activities: [Activity]

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(12)) {
            SectionHeader(title, action: seeAll, onAction: {})
            VStack(spacing: 0) {
                ForEach(activities) { activity in
                    HStack(spacing: metrics.points(12)) {
                        PhotoStandIn(palette: .named(activity.photoReference), label: "", labelSize: 8)
                            .frame(width: metrics.points(56), height: metrics.points(48))
                            .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.stop), style: .continuous))
                        VStack(alignment: .leading, spacing: metrics.points(2)) {
                            Text(activity.name).typeStyle(Typography.cardTitle).foregroundStyle(Palette.label)
                            Text(activity.summary).typeStyle(Typography.caption).foregroundStyle(Palette.labelAlpha(0.6))
                        }
                        Spacer(minLength: 0)
                        Image(systemName: "chevron.forward").font(.system(size: metrics.points(13), weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.35))
                    }
                    .padding(.horizontal, metrics.points(12))
                    .padding(.vertical, metrics.points(11))
                    if activity != activities.last {
                        Divider().overlay(Palette.labelAlpha(0.07))
                    }
                }
            }
            .surfaceCard()
        }
    }
}
