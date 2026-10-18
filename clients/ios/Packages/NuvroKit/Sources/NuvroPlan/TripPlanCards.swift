import NuvroDesignSystem
import NuvroModels
import SwiftUI

struct ItineraryPreviewCard: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let dayLabel: String
    let viewItinerary: String
    let days: [ItineraryDay]
    let onViewItinerary: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(12)) {
            HStack(alignment: .firstTextBaseline) {
                Text(title).typeStyle(Typography.cardHeading).foregroundStyle(Palette.label)
                Spacer()
                LinkButton(viewItinerary, action: onViewItinerary)
            }
            VStack(spacing: 0) {
                ForEach(Array(days.enumerated()), id: \.element.id) { index, day in
                    HStack(alignment: .top, spacing: metrics.points(14)) {
                        VStack(spacing: metrics.points(4)) {
                            Circle()
                                .fill(index == 0 ? Palette.sky : Palette.aqua)
                                .overlay(Circle().strokeBorder(.white, lineWidth: 2))
                                .overlay(Circle().strokeBorder(index == 0 ? Palette.sky : Palette.aqua, lineWidth: 1.5).padding(-1.5))
                                .frame(width: metrics.points(10), height: metrics.points(10))
                                .padding(.top, metrics.points(5))
                            if index < days.count - 1 {
                                Rectangle().fill(Palette.sky.opacity(0.2)).frame(width: 2).frame(maxHeight: .infinity)
                            }
                        }
                        .frame(width: metrics.points(14))
                        VStack(alignment: .leading, spacing: metrics.points(2)) {
                            Text("\(dayLabel) \(day.dayNumber)").typeStyle(Typography.dayLabel).foregroundStyle(Palette.sky)
                            Text(day.summary).typeStyle(Typography.secondaryAction).foregroundStyle(Palette.label)
                        }
                        .padding(.bottom, index < days.count - 1 ? metrics.points(16) : 0)
                        Spacer(minLength: 0)
                    }
                    .fixedSize(horizontal: false, vertical: true)
                }
            }
        }
        .padding(metrics.points(16))
        .surfaceCard()
    }
}

struct PlanningModulesGrid: View {
    @Environment(\.designMetrics) private var metrics
    let onSelect: (PlanningModule) -> Void

    private var columns: [GridItem] { Array(repeating: GridItem(.flexible(), spacing: metrics.points(10)), count: 4) }

    var body: some View {
        LazyVGrid(columns: columns, spacing: metrics.points(12)) {
            ForEach(Array(PlanningModule.allCases.enumerated()), id: \.element) { index, module in
                Button { onSelect(module) } label: {
                    VStack(spacing: metrics.points(7)) {
                        IconTile(palette: PhotoPalette.iconTiles[index % 4], systemImage: module.systemImage, size: 26)
                            .frame(maxWidth: .infinity)
                            .aspectRatio(1, contentMode: .fit)
                            .surfaceCard(radius: Radius.tile)
                        Text(module.title)
                            .typeStyle(Typography.tileLabel)
                            .foregroundStyle(Palette.labelAlpha(0.8))
                            .lineLimit(1)
                            .minimumScaleFactor(0.8)
                    }
                }
                .buttonStyle(PressScaleButtonStyle(scale: 0.96))
            }
        }
    }
}

struct AskNuvroCard: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let placeholder: String
    let suggestions: [String]
    @Binding var text: String
    let onSuggest: (String) -> Void
    let onSend: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(12)) {
            HStack(spacing: metrics.points(8)) {
                RouteArcGlyph(width: metrics.points(16))
                Text(title).typeStyle(TypeStyle(size: 15, weight: .bold)).foregroundStyle(Palette.label)
            }
            FlowLayout(spacing: metrics.points(8), rowSpacing: metrics.points(8)) {
                ForEach(suggestions, id: \.self) { suggestion in
                    Button { onSuggest(suggestion) } label: {
                        Text("“\(suggestion)”")
                            .typeStyle(Typography.metaEmphasis)
                            .foregroundStyle(Palette.label)
                            .padding(.horizontal, metrics.points(12))
                            .padding(.vertical, metrics.points(8))
                            .background(Palette.surface.opacity(0.85), in: Capsule())
                            .overlay(Capsule().strokeBorder(Palette.labelAlpha(0.08), lineWidth: 0.5))
                    }
                    .buttonStyle(PressScaleButtonStyle(scale: 0.97))
                }
            }
            HStack(spacing: metrics.points(10)) {
                Button(action: onSend) {
                    Text(text.isEmpty ? placeholder : text)
                        .typeStyle(Typography.bodyCompact)
                        .foregroundStyle(text.isEmpty ? Palette.labelAlpha(0.4) : Palette.label)
                        .lineLimit(1)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .contentShape(Rectangle())
                }
                .buttonStyle(.plain)
                .accessibilityLabel(placeholder)
                Button(action: onSend) {
                    Image(systemName: "arrow.up")
                        .font(.system(size: metrics.points(13), weight: .bold))
                        .foregroundStyle(.white)
                        .frame(width: metrics.points(34), height: metrics.points(34))
                        .background(Palette.sky.opacity(0.88), in: Circle())
                }
                .buttonStyle(PressScaleButtonStyle(scale: 0.92))
                .disabled(text.isEmpty)
                .opacity(text.isEmpty ? 0.5 : 1)
                .accessibilityLabel("Ask")
            }
            .padding(.leading, metrics.points(16))
            .padding(.trailing, metrics.points(6))
            .frame(height: metrics.points(46))
            .glassCapsule(.light)
        }
        .padding(metrics.points(16))
        .background(
            LinearGradient(colors: [Palette.sky.opacity(0.08), Palette.aqua.opacity(0.10)], startPoint: .topLeading, endPoint: .bottomTrailing),
            in: RoundedRectangle(cornerRadius: metrics.points(Radius.card), style: .continuous)
        )
        .overlay(RoundedRectangle(cornerRadius: metrics.points(Radius.card), style: .continuous).strokeBorder(Palette.sky.opacity(0.18), lineWidth: 0.5))
    }
}
