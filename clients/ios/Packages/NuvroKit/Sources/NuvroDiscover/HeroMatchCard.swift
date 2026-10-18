import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import SwiftUI

struct HeroMatchCard: View {
    @Environment(\.designMetrics) private var metrics
    let match: DestinationMatch
    let copy: MatchResultsCopy
    let locale: Locale
    let isRevealed: Bool
    let isSaved: Bool
    let isComparing: Bool
    let onExplore: () -> Void
    let onSave: () -> Void
    let onCompare: () -> Void

    var body: some View {
        VStack(spacing: 0) {
            photo
            VStack(alignment: .leading, spacing: metrics.points(14)) {
                HStack(spacing: metrics.points(8)) {
                    FactTile(key: copy.estimatedTrip, value: estimateLabel)
                    FactTile(key: copy.weather, value: weatherLabel)
                    FactTile(key: copy.flight, value: flightLabel)
                }
                tags
                VStack(alignment: .leading, spacing: metrics.points(8)) {
                    Divider().overlay(Palette.labelAlpha(0.08))
                    Text(copy.whyItFits).typeStyle(TypeStyle(size: 15, weight: .bold)).foregroundStyle(Palette.label).padding(.top, metrics.points(8))
                    ForEach(match.reasons) { reason in
                        MatchReasonRow(text: reason.text, isPositive: reason.isPositive)
                    }
                }
                HStack(spacing: metrics.points(10)) {
                    CallToActionButton("\(copy.explore) \(match.name)", action: onExplore)
                    Button(action: onCompare) {
                        Text(isComparing ? copy.comparing : copy.compare)
                            .typeStyle(Typography.secondaryAction)
                            .foregroundStyle(isComparing ? Palette.background : Palette.label)
                            .padding(.horizontal, metrics.points(18))
                            .frame(height: metrics.points(54))
                            .background(isComparing ? Palette.label : Palette.labelAlpha(0.06), in: Capsule())
                            .overlay(Capsule().strokeBorder(Palette.labelAlpha(0.1), lineWidth: 0.5))
                    }
                    .buttonStyle(PressScaleButtonStyle(scale: 0.97))
                    .animation(Motion.fade(0.2), value: isComparing)
                }
                .padding(.top, metrics.points(4))
            }
            .padding(EdgeInsets(top: metrics.points(14), leading: metrics.points(16), bottom: metrics.points(16), trailing: metrics.points(16)))
        }
        .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.hero), style: .continuous))
        .surfaceCard(radius: Radius.hero, elevated: true)
    }

    private var photo: some View {
        Button(action: onExplore) {
            DynamicDestinationImage(media: MediaDescriptor(url: match.photoURL, placeholder: match.photoReference, altText: match.displayName))
                .frame(height: metrics.points(230))
                .overlay(SkyGradient.photoFooterScrim)
                .overlay(alignment: .topTrailing) {
                    SaveButton(isSaved: isSaved, style: .control, action: onSave).padding(metrics.points(14))
                }
                .overlay(alignment: .bottomLeading) {
                    VStack(alignment: .leading, spacing: metrics.points(2)) {
                        Text(match.displayName).typeStyle(TypeStyle(size: 24, weight: .bold, tracking: -0.4)).foregroundStyle(Palette.cloud)
                        Text("\(match.region) · \(match.durationDays) days").typeStyle(Typography.meta).foregroundStyle(Palette.cloudAlpha(0.85))
                    }
                    .padding(.leading, metrics.points(18))
                    .padding(.bottom, metrics.points(16))
                }
                .overlay(alignment: .bottomTrailing) {
                    MatchRing(percent: match.score.percent, isRevealed: isRevealed)
                        .padding(.trailing, metrics.points(16))
                        .padding(.bottom, metrics.points(14))
                }
        }
        .buttonStyle(.plain)
        .accessibilityLabel("\(match.displayName), \(match.score.percent) percent match")
    }

    private var tags: some View {
        HStack(spacing: metrics.points(6)) {
            ForEach(Array(match.tags.prefix(3).enumerated()), id: \.offset) { index, tag in
                TagChip(tag, tone: [.aqua, .gold, .twilight][index % 3])
            }
        }
    }

    private var estimateLabel: String {
        guard let estimate = match.estimate else { return unknown }
        let high = estimate.high.majorUnits.formatted(.number.precision(.fractionLength(0)).locale(locale))
        return "\(estimate.low.formattedWhole(locale: locale))–\(high)"
    }

    private var weatherLabel: String {
        guard let weather = match.weather else { return unknown }
        return "\(weather.condition) · \(weather.temperatureCelsius)°"
    }

    private var flightLabel: String {
        guard let flight = match.flight else { return unknown }
        return "\(flight.stops) \(flight.stops == 1 ? "stop" : "stops") · ~\(flight.approximateHours)h"
    }

    private var unknown: String { "Not yet known" }
}
