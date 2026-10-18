import NuvroDesignSystem
import NuvroModels
import SwiftUI

public struct CompareView: View {
    @Environment(\.designMetrics) private var metrics
    private let model: SavedModel

    public init(model: SavedModel) {
        self.model = model
    }

    private var comparison: DestinationComparison { model.comparison }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(16)) {
                HStack(alignment: .top, spacing: metrics.points(10)) {
                    ForEach(comparison.columns) { column in
                        VStack(alignment: .leading, spacing: 0) {
                            DynamicDestinationImage(media: column.media.descriptor, labelSize: 6).frame(height: metrics.points(110))
                            VStack(alignment: .leading, spacing: metrics.points(7)) {
                                Text(column.name).typeStyle(TypeStyle(size: 16, weight: .bold, tracking: -0.2)).foregroundStyle(Palette.label)
                                HStack(spacing: metrics.points(6)) {
                                    Text("\(column.matchPercent)%").typeStyle(TypeStyle(size: 14, weight: .bold, tracking: -0.2)).foregroundStyle(Palette.matchTeal)
                                    Text(model.copy.match).typeStyle(TypeStyle(size: 8.5, weight: .bold, tracking: 0.85)).foregroundStyle(Palette.matchTeal)
                                }
                                .padding(.horizontal, metrics.points(9)).padding(.vertical, metrics.points(4)).background(Palette.aqua.opacity(0.16), in: Capsule())
                                Text(column.gist).typeStyle(TypeStyle(size: 12, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.62))
                            }
                            .padding(EdgeInsets(top: metrics.points(11), leading: metrics.points(12), bottom: metrics.points(13), trailing: metrics.points(12)))
                        }
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.card), style: .continuous))
                        .surfaceCard(radius: Radius.card, elevated: true)
                    }
                }
                VStack(spacing: 0) {
                    ForEach(comparison.metrics) { metric in
                        VStack(alignment: .leading, spacing: metrics.points(5)) {
                            Text(metric.key.uppercased()).typeStyle(TypeStyle(size: 11, weight: .bold, tracking: 0.66)).foregroundStyle(Palette.labelAlpha(0.5))
                            HStack(spacing: metrics.points(10)) {
                                metricCell(metric.left, wins: metric.winner == .left)
                                metricCell(metric.right, wins: metric.winner == .right)
                            }
                        }
                        .padding(.horizontal, metrics.points(13)).padding(.vertical, metrics.points(10))
                        if metric.id != comparison.metrics.last?.id { Divider().overlay(Palette.labelAlpha(0.07)) }
                    }
                }
                .surfaceCard()
                InsightCard(title: model.compareCopy.differences, showsMark: false) {
                    ForEach(comparison.differences) { difference in
                        VStack(alignment: .leading, spacing: metrics.points(2)) {
                            Text(difference.name).typeStyle(TypeStyle(size: 13.5, weight: .bold)).foregroundStyle(Palette.label)
                            Text(difference.text).typeStyle(TypeStyle(size: 13, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.7))
                        }
                        .padding(.horizontal, metrics.points(12)).padding(.vertical, metrics.points(11))
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .background(Palette.surface.opacity(0.85), in: RoundedRectangle(cornerRadius: metrics.points(14), style: .continuous))
                    }
                }
                HStack(spacing: metrics.points(10)) {
                    ForEach(comparison.columns) { column in
                        let chosen = model.chosenColumn == column.id
                        Button { model.choose(column) } label: {
                            Text(model.columnLabel(column)).typeStyle(TypeStyle(size: 15, weight: .semibold)).foregroundStyle(chosen ? .white : Palette.label)
                                .frame(maxWidth: .infinity).frame(height: metrics.points(50))
                                .background(chosen ? Palette.sky.opacity(0.88) : Palette.labelAlpha(0.06), in: Capsule())
                                .overlay(Capsule().strokeBorder(chosen ? Color.white.opacity(0.45) : Palette.labelAlpha(0.1), lineWidth: 0.5))
                                .shadow(color: chosen ? Palette.sky.opacity(0.14) : .clear, radius: 6, y: 3)
                        }
                        .buttonStyle(PressScaleButtonStyle(scale: 0.97))
                        .accessibilityAddTraits(chosen ? .isSelected : [])
                    }
                }
            }
            .padding(.horizontal, metrics.points(20)).padding(.top, metrics.points(8)).padding(.bottom, metrics.points(40))
            .animation(Motion.fade(0.25), value: model.chosenColumn)
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .screenTitle(model.compareCopy.title, subtitle: comparison.subtitle)
    }

    @ViewBuilder
    private func metricCell(_ value: String, wins: Bool) -> some View {
        if value.isEmpty {
            FreshnessLabel(tone: .unavailable, text: model.compareCopy.unavailable)
                .frame(maxWidth: .infinity, alignment: .leading)
        } else {
            HStack(spacing: metrics.points(6)) {
                Text(value).typeStyle(TypeStyle(size: 14.5, weight: wins ? .bold : .semibold)).foregroundStyle(wins ? Palette.label : Palette.labelAlpha(0.75))
                if wins { Text(model.compareCopy.better).typeStyle(TypeStyle(size: 10, weight: .bold)).foregroundStyle(Palette.matchTeal) }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }
}
