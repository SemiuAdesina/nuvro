import MapKit
import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import Observation
import SwiftUI

@Observable
@MainActor
public final class TransportModel {
    public let guide: TransportGuide
    public var selectedArrival: String
    public let copy: TransportCopy

    public init(guide: TransportGuide, language: Language = .english) {
        self.guide = guide
        self.selectedArrival = guide.arrivals.first { $0.tag == .bestForPlan }?.name ?? guide.arrivals.first?.name ?? ""
        self.copy = .resolve(language)
    }

    public func tagLabel(_ tag: ArrivalOption.Tag) -> String {
        switch tag {
        case .recommended: copy.recommended
        case .bestForPlan: copy.bestForPlan
        }
    }
}

public struct TransportView: View {
    @Environment(\.designMetrics) private var metrics
    @Bindable private var model: TransportModel
    private let subtitle: String

    public init(model: TransportModel, subtitle: String) {
        self.model = model
        self.subtitle = subtitle
    }

    private var copy: TransportCopy { model.copy }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(16)) {
                InsightCard(title: copy.bestForTrip) {
                    Text(model.guide.recommendationTitle).typeStyle(TypeStyle(size: 17, weight: .semibold, tracking: -0.2)).foregroundStyle(Palette.label)
                    Text(model.guide.recommendationBody).typeStyle(TypeStyle(size: 13.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.75))
                }
                VStack(alignment: .leading, spacing: metrics.points(11)) {
                    Text(copy.fromAirport).typeStyle(TypeStyle(size: 18, weight: .bold, tracking: -0.2)).foregroundStyle(Palette.label)
                    VStack(spacing: 0) {
                        ForEach(model.guide.arrivals) { option in
                            ArrivalRow(option: option, model: model)
                            if option.id != model.guide.arrivals.last?.id { Divider().overlay(Palette.labelAlpha(0.07)) }
                        }
                    }
                    .surfaceCard()
                }
                transferMap
                VStack(alignment: .leading, spacing: metrics.points(11)) {
                    Text(copy.aroundIsland).typeStyle(TypeStyle(size: 18, weight: .bold, tracking: -0.2)).foregroundStyle(Palette.label)
                    VStack(spacing: 0) {
                        ForEach(model.guide.local) { item in
                            HStack(spacing: metrics.points(12)) {
                                PhotoStandIn(palette: .named(item.placeholder), label: "", labelSize: 6).frame(width: metrics.points(30), height: metrics.points(30)).clipShape(RoundedRectangle(cornerRadius: metrics.points(9), style: .continuous))
                                VStack(alignment: .leading, spacing: metrics.points(2)) {
                                    Text(item.name).typeStyle(TypeStyle(size: 15, weight: .semibold)).foregroundStyle(Palette.label)
                                    Text(item.note).typeStyle(TypeStyle(size: 12, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
                                }
                                Spacer(minLength: 0)
                            }
                            .padding(.horizontal, metrics.points(14)).padding(.vertical, metrics.points(12))
                            if item.id != model.guide.local.last?.id { Divider().overlay(Palette.labelAlpha(0.07)) }
                        }
                    }
                    .surfaceCard()
                    Text(copy.disclaimer).typeStyle(TypeStyle(size: 11, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.45)).multilineTextAlignment(.center).frame(maxWidth: .infinity)
                }
            }
            .padding(.horizontal, metrics.points(20)).padding(.top, metrics.points(8)).padding(.bottom, metrics.points(40))
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .screenTitle(copy.title, subtitle: subtitle)
    }

    private var transferMap: some View {
        let airport = CLLocationCoordinate2D(latitude: model.guide.airport.latitude, longitude: model.guide.airport.longitude)
        let hotel = CLLocationCoordinate2D(latitude: model.guide.hotel.latitude, longitude: model.guide.hotel.longitude)
        let centre = CLLocationCoordinate2D(latitude: (airport.latitude + hotel.latitude) / 2, longitude: (airport.longitude + hotel.longitude) / 2)
        return Map(initialPosition: .region(MKCoordinateRegion(center: centre, span: MKCoordinateSpan(latitudeDelta: 0.12, longitudeDelta: 0.28))), interactionModes: []) {
            MapPolyline(coordinates: [airport, hotel]).stroke(Palette.sky, style: StrokeStyle(lineWidth: 2.5, dash: [5, 6]))
            Annotation(copy.airport, coordinate: airport) { NuvroMapMarker(letter: "A", colour: Palette.twilightText) }
            Annotation(copy.hotel, coordinate: hotel) { NuvroMapMarker(letter: "H", colour: Palette.sky) }
        }
        .mapStyle(.standard(elevation: .flat, pointsOfInterest: .excludingAll))
        .mapControlVisibility(.hidden)
        .frame(height: metrics.points(190))
        .overlay(alignment: .bottomLeading) {
            Text(model.guide.transferLabel).typeStyle(TypeStyle(size: 12, weight: .semibold)).foregroundStyle(Palette.label)
                .padding(.horizontal, metrics.points(12)).frame(height: metrics.points(30)).glassCapsule(.light).padding(metrics.points(12))
        }
        .clipShape(RoundedRectangle(cornerRadius: metrics.points(22), style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: metrics.points(22), style: .continuous).strokeBorder(Palette.labelAlpha(0.08), lineWidth: 0.5))
    }
}

struct ArrivalRow: View {
    @Environment(\.designMetrics) private var metrics
    let option: ArrivalOption
    let model: TransportModel

    var body: some View {
        Button { model.selectedArrival = option.name } label: {
            HStack(spacing: metrics.points(12)) {
                PhotoStandIn(palette: .named(option.placeholder), label: "", labelSize: 6).frame(width: metrics.points(32), height: metrics.points(32)).clipShape(RoundedRectangle(cornerRadius: metrics.points(10), style: .continuous))
                VStack(alignment: .leading, spacing: metrics.points(2)) {
                    HStack(spacing: metrics.points(7)) {
                        Text(option.name).typeStyle(TypeStyle(size: 15, weight: .semibold)).foregroundStyle(Palette.label)
                        if let tag = option.tag {
                            Text(model.tagLabel(tag)).typeStyle(TypeStyle(size: 10, weight: .bold, tracking: 0.5)).foregroundStyle(tag == .recommended ? Palette.matchTeal : Palette.sky)
                                .padding(.horizontal, metrics.points(7)).padding(.vertical, metrics.points(2.5))
                                .background(tag == .recommended ? Palette.aqua.opacity(0.2) : Palette.sky.opacity(0.14), in: Capsule())
                        }
                    }
                    Text(option.meta).typeStyle(TypeStyle(size: 12, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
                }
                Spacer(minLength: 0)
                Text(option.costLabel).typeStyle(TypeStyle(size: 14.5, weight: .bold)).foregroundStyle(Palette.label)
            }
            .padding(.horizontal, metrics.points(14)).padding(.vertical, metrics.points(13))
            .background(model.selectedArrival == option.name ? Palette.sky.opacity(0.06) : .clear)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .accessibilityAddTraits(model.selectedArrival == option.name ? .isSelected : [])
    }
}
