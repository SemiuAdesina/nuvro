import NuvroDesignSystem
import NuvroModels
import SwiftUI

struct TripIdentityHeader: View {
    @Environment(\.designMetrics) private var metrics
    let destination: DestinationMatch
    let subtitle: String
    var compact = false

    var body: some View {
        HStack(spacing: metrics.points(14)) {
            PhotoStandIn(palette: .named(destination.photoReference), label: "", labelSize: 8)
                .frame(width: metrics.points(compact ? 56 : 64), height: metrics.points(compact ? 56 : 64))
                .clipShape(RoundedRectangle(cornerRadius: metrics.points(compact ? 16 : Radius.tripIdentity), style: .continuous))
                .shadow(color: Palette.labelAlpha(0.06), radius: metrics.points(3), y: metrics.points(1))
            VStack(alignment: .leading, spacing: metrics.points(compact ? 2 : 3)) {
                Text(destination.name).typeStyle(compact ? TypeStyle(size: 28, weight: .bold, tracking: -0.6) : Typography.headline).foregroundStyle(Palette.label)
                Text(subtitle).typeStyle(compact ? Typography.bodyCompact : Typography.body).foregroundStyle(Palette.labelAlpha(0.62))
            }
        }
    }
}

struct ReadinessCard: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let label: String
    let fraction: Double
    let items: [ReadinessItem]
    let onToggle: (ReadinessItem) -> Void

    private var columns: [GridItem] { [GridItem(.flexible(), spacing: metrics.points(14)), GridItem(.flexible(), spacing: metrics.points(14))] }

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(12)) {
            HStack(alignment: .firstTextBaseline) {
                Text(title).typeStyle(Typography.cardHeading).foregroundStyle(Palette.label)
                Spacer()
                Text(label).typeStyle(Typography.link).foregroundStyle(Palette.labelAlpha(0.6)).contentTransition(.numericText())
            }
            GeometryReader { proxy in
                ZStack(alignment: .leading) {
                    Capsule().fill(Palette.labelAlpha(0.07))
                    Capsule()
                        .fill(LinearGradient(colors: [Palette.aqua, Palette.sky], startPoint: .leading, endPoint: .trailing))
                        .frame(width: proxy.size.width * fraction)
                        .animation(Motion.settle(0.8), value: fraction)
                }
            }
            .frame(height: metrics.points(6))
            .accessibilityElement(children: .ignore)
            .accessibilityLabel(label)
            LazyVGrid(columns: columns, alignment: .leading, spacing: metrics.points(8)) {
                ForEach(items) { item in
                    Button { onToggle(item) } label: {
                        HStack(spacing: metrics.points(9)) {
                            Circle()
                                .fill(item.isReady ? Palette.aqua : Color.clear)
                                .overlay(Circle().strokeBorder(item.isReady ? Palette.aqua : Palette.labelAlpha(0.25), lineWidth: 1.5))
                                .overlay {
                                    if item.isReady {
                                        Image(systemName: "checkmark").font(.system(size: metrics.points(10), weight: .bold)).foregroundStyle(Palette.label)
                                    }
                                }
                                .frame(width: metrics.points(20), height: metrics.points(20))
                            Text(item.name).typeStyle(Typography.bodyCompact).foregroundStyle(item.isReady ? Palette.label : Palette.labelAlpha(0.6))
                            Spacer(minLength: 0)
                        }
                        .contentShape(Rectangle())
                    }
                    .buttonStyle(.plain)
                    .accessibilityAddTraits(item.isReady ? .isSelected : [])
                    .accessibilityValue(item.isReady ? "Ready" : "Not ready")
                    .animation(Motion.fade(0.2), value: item.isReady)
                }
            }
            .padding(.top, metrics.points(2))
        }
        .padding(metrics.points(16))
        .surfaceCard()
    }
}

struct BudgetSummaryCard: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let estimate: String
    let status: String?
    let isWithinBudget: Bool
    let lines: [BudgetLine]
    let share: (BudgetLine) -> Double
    let viewBudget: String
    let locale: Locale
    let onViewBudget: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(12)) {
            HStack(alignment: .bottom) {
                VStack(alignment: .leading, spacing: metrics.points(2)) {
                    FieldLabel(title)
                    Text(estimate).typeStyle(Typography.estimateFigure).foregroundStyle(Palette.label)
                }
                Spacer()
                if let status {
                    Text(status)
                        .typeStyle(Typography.metaEmphasis)
                        .foregroundStyle(isWithinBudget ? Palette.matchTeal : Palette.amber)
                        .padding(.horizontal, metrics.points(10))
                        .padding(.vertical, metrics.points(5))
                        .background((isWithinBudget ? Palette.aqua : Palette.gold).opacity(0.16), in: Capsule())
                }
            }
            HStack(spacing: metrics.points(2)) {
                ForEach(lines) { line in
                    Rectangle().fill(BudgetCategoryTone.color(for: line.category)).frame(width: max(2, barWidth(for: line)))
                }
            }
            .frame(height: metrics.points(10))
            .clipShape(Capsule())
            .accessibilityHidden(true)
            FlowLayout(spacing: metrics.points(8), rowSpacing: metrics.points(8)) {
                ForEach(lines) { line in
                    HStack(spacing: metrics.points(6)) {
                        Circle().fill(BudgetCategoryTone.color(for: line.category)).frame(width: metrics.points(8), height: metrics.points(8))
                        Text(line.category).typeStyle(Typography.caption).foregroundStyle(Palette.labelAlpha(0.65))
                        Text(line.amount.formattedWhole(locale: locale)).typeStyle(Typography.captionEmphasis).foregroundStyle(Palette.label)
                    }
                    .padding(.trailing, metrics.points(6))
                }
            }
            LinkButton(viewBudget, action: onViewBudget).padding(.top, metrics.points(2))
        }
        .padding(metrics.points(16))
        .surfaceCard()
    }

    private func barWidth(for line: BudgetLine) -> CGFloat {
        (metrics.size.width - metrics.points(72)) * share(line)
    }
}
