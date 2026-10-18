import NuvroDesignSystem
import NuvroModels
import SwiftUI

public struct TripSummaryView: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @State private var hasEntered = false
    private let model: TripSummaryModel
    private let onPlanAnother: () -> Void
    private let onViewTrip: () -> Void

    public init(model: TripSummaryModel, onPlanAnother: @escaping () -> Void, onViewTrip: @escaping () -> Void) {
        self.model = model
        self.onPlanAnother = onPlanAnother
        self.onViewTrip = onViewTrip
    }

    private var summary: TripSummary { model.summary }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(20)) {
                hero
                VStack(alignment: .leading, spacing: metrics.points(20)) {
                    HStack(spacing: metrics.points(8)) {
                        ForEach(model.facts, id: \.label) { fact in
                            VStack(alignment: .leading, spacing: metrics.points(2)) {
                                Text(fact.value).typeStyle(TypeStyle(size: 19, weight: .bold, tracking: -0.4)).foregroundStyle(Palette.label)
                                Text(fact.label).typeStyle(TypeStyle(size: 11, weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.55))
                            }
                            .frame(maxWidth: .infinity, alignment: .leading)
                            .padding(.horizontal, metrics.points(10)).padding(.vertical, metrics.points(12))
                            .surfaceCard(radius: 16)
                        }
                    }
                    placesVisited
                    spending
                    learning
                    nextSuggestion
                    Text(model.copy.footer).typeStyle(TypeStyle(size: 11, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.45)).multilineTextAlignment(.center).frame(maxWidth: .infinity)
                    InlineActions {
                        CallToActionButton(model.copy.planAnother, action: onPlanAnother)
                        Button(model.copy.viewTrip, action: onViewTrip)
                            .typeStyle(Typography.secondaryAction).foregroundStyle(Palette.sky).frame(height: metrics.points(34))
                            .buttonStyle(.plain)
                    }
                }
                .padding(.horizontal, metrics.points(20))
                .padding(.bottom, metrics.points(40))
            }
            .opacity(hasEntered ? 1 : 0)
            .offset(y: hasEntered || reduceMotion ? 0 : metrics.points(16))
        }
        .scrollIndicators(.hidden)
        .ignoresSafeArea(edges: .top)
        .background(Palette.background.ignoresSafeArea())
        .onAppear { withAnimation(reduceMotion ? nil : Motion.settle(0.7)) { hasEntered = true } }
    }

    private var hero: some View {
        DynamicDestinationImage(media: summary.hero.descriptor, label: "photo · madeira trip hero, real destination photo", labelSize: 8)
            .frame(height: metrics.points(420))
            .photoScrim(LinearGradient(stops: [.init(color: Palette.navy.opacity(0.3), location: 0), .init(color: .clear, location: 0.26), .init(color: .clear, location: 0.52), .init(color: Palette.cloud, location: 1)], startPoint: .top, endPoint: .bottom))
            .overlay(alignment: .bottomLeading) {
                VStack(alignment: .leading, spacing: metrics.points(6)) {
                    Text("\(summary.destinationName) · \(summary.dateRangeLabel)".uppercased()).typeStyle(TypeStyle(size: 13, weight: .bold, tracking: 1.3)).foregroundStyle(Palette.labelAlpha(0.6))
                    Text(model.copy.headline).typeStyle(TypeStyle(size: 38, weight: .bold, tracking: -0.9)).foregroundStyle(Palette.label)
                }
                .padding(.horizontal, metrics.points(20)).padding(.bottom, metrics.points(26))
            }
    }

    private var placesVisited: some View {
        VStack(alignment: .leading, spacing: metrics.points(10)) {
            Text(model.copy.placesVisited).typeStyle(Typography.title).foregroundStyle(Palette.label)
            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: metrics.points(10)) {
                    ForEach(summary.places) { place in
                        VStack(spacing: metrics.points(7)) {
                            DynamicDestinationImage(media: place.media.descriptor, label: "photo · \(place.name.lowercased())", labelSize: 6)
                                .frame(width: metrics.points(118), height: metrics.points(118))
                                .clipShape(RoundedRectangle(cornerRadius: metrics.points(18), style: .continuous))
                            Text(place.name).typeStyle(TypeStyle(size: 13, weight: .semibold)).foregroundStyle(Palette.label).lineLimit(1)
                        }
                        .frame(width: metrics.points(118))
                    }
                }
                .padding(.horizontal, metrics.points(20))
            }
            .scrollClipDisabled()
            .padding(.horizontal, -metrics.points(20))
        }
    }

    private var spending: some View {
        HStack(alignment: .bottom, spacing: metrics.points(12)) {
            VStack(alignment: .leading, spacing: metrics.points(3)) {
                Text(model.copy.tripTotal).typeStyle(TypeStyle(size: 12, weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.55))
                Text(model.totalLabel).typeStyle(TypeStyle(size: 30, weight: .bold, tracking: -0.7)).foregroundStyle(Palette.label)
            }
            Spacer()
            if let budgetLabel = model.budgetLabel {
                Text(budgetLabel).typeStyle(TypeStyle(size: 13, weight: .semibold)).foregroundStyle(Palette.matchTeal)
                    .padding(.horizontal, metrics.points(11)).padding(.vertical, metrics.points(6))
                    .background(Palette.aqua.opacity(0.16), in: Capsule())
            }
        }
        .padding(metrics.points(16))
        .surfaceCard()
    }

    private var learning: some View {
        VStack(alignment: .leading, spacing: metrics.points(14)) {
            VStack(alignment: .leading, spacing: metrics.points(3)) {
                Text(model.copy.learningTitle).typeStyle(TypeStyle(size: 16, weight: .bold)).foregroundStyle(Palette.label)
                Text(model.copy.learningBody).typeStyle(TypeStyle(size: 13, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.65))
            }
            PreferenceFeedbackRow(title: model.copy.loved.uppercased(), tone: .loved, options: summary.lovedOptions, selected: model.loved, toggle: model.toggleLoved)
            PreferenceFeedbackRow(title: model.copy.notForMe.uppercased(), tone: .notForMe, options: summary.notForMeOptions, selected: model.notForMe, toggle: model.toggleNotForMe)
        }
        .padding(metrics.points(16))
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(LinearGradient(colors: [Palette.sky.opacity(0.08), Palette.aqua.opacity(0.10)], startPoint: .topLeading, endPoint: .bottomTrailing), in: RoundedRectangle(cornerRadius: metrics.points(Radius.card), style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: metrics.points(Radius.card), style: .continuous).strokeBorder(Palette.sky.opacity(0.18), lineWidth: 0.5))
    }

    private var nextSuggestion: some View {
        VStack(alignment: .leading, spacing: metrics.points(12)) {
            Text(model.copy.nextTitle).typeStyle(Typography.title).foregroundStyle(Palette.label)
            VStack(spacing: 0) {
                DynamicDestinationImage(media: summary.suggestion.media.descriptor, label: "photo · azores, real destination photo", labelSize: 7)
                    .frame(height: metrics.points(160))
                HStack(spacing: metrics.points(12)) {
                    VStack(alignment: .leading, spacing: metrics.points(3)) {
                        Text("\(summary.suggestion.name), \(summary.suggestion.country)").typeStyle(TypeStyle(size: 17, weight: .semibold, tracking: -0.2)).foregroundStyle(Palette.label)
                        Text(summary.suggestion.reason).typeStyle(TypeStyle(size: 12.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.62))
                    }
                    Spacer(minLength: 0)
                    Image(systemName: "chevron.forward").font(.system(size: metrics.points(12), weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.3))
                }
                .padding(EdgeInsets(top: metrics.points(12), leading: metrics.points(14), bottom: metrics.points(14), trailing: metrics.points(14)))
            }
            .clipShape(RoundedRectangle(cornerRadius: metrics.points(22), style: .continuous))
            .surfaceCard(radius: 22, elevated: true)
        }
    }
}

struct PreferenceFeedbackRow: View {
    enum Tone { case loved, notForMe }

    @Environment(\.designMetrics) private var metrics
    let title: String
    let tone: Tone
    let options: [String]
    let selected: Set<String>
    let toggle: (String) -> Void

    private var onColours: (fill: Color, border: Color, text: Color) {
        tone == .loved ? (Palette.aqua.opacity(0.22), Palette.aqua.opacity(0.5), Palette.matchTeal) : (Palette.gold.opacity(0.25), Palette.gold.opacity(0.6), Palette.amber)
    }

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(8)) {
            Text(title).typeStyle(TypeStyle(size: 11, weight: .bold, tracking: 1.1)).foregroundStyle(tone == .loved ? Palette.matchTeal : Palette.amber)
            FlowLayout(spacing: metrics.points(7), rowSpacing: metrics.points(7)) {
                ForEach(options, id: \.self) { option in
                    let isOn = selected.contains(option)
                    Button { toggle(option) } label: {
                        HStack(spacing: metrics.points(7)) {
                            Text(isOn ? "✓" : "+").typeStyle(TypeStyle(size: 11, weight: .bold))
                            Text(option).typeStyle(TypeStyle(size: 13.5, weight: .semibold)).lineLimit(1)
                        }
                        .fixedSize()
                        .foregroundStyle(isOn ? onColours.text : Palette.labelAlpha(0.6))
                        .padding(.horizontal, metrics.points(12)).frame(height: metrics.points(34))
                        .background(isOn ? onColours.fill : Palette.surface.opacity(0.8), in: Capsule())
                        .overlay(Capsule().strokeBorder(isOn ? onColours.border : Palette.labelAlpha(0.12), lineWidth: 0.5))
                    }
                    .buttonStyle(PressScaleButtonStyle(scale: 0.96))
                    .accessibilityAddTraits(isOn ? .isSelected : [])
                }
            }
            .frame(maxWidth: .infinity, alignment: .leading)
        }
    }
}
