import MapKit
import NuvroDesignSystem
import NuvroModels
import SwiftUI

struct MapPreview: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let openMap: String
    let guide: DestinationGuide

    private var region: MKCoordinateRegion {
        MKCoordinateRegion(
            center: CLLocationCoordinate2D(latitude: guide.centre.latitude, longitude: guide.centre.longitude),
            span: MKCoordinateSpan(latitudeDelta: 0.12, longitudeDelta: 0.2)
        )
    }

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(12)) {
            Text(title).typeStyle(Typography.title).foregroundStyle(Palette.label)
            Map(initialPosition: .region(region), interactionModes: []) {
                ForEach(guide.places) { place in
                    Annotation(place.name, coordinate: CLLocationCoordinate2D(latitude: place.coordinate.latitude, longitude: place.coordinate.longitude)) {
                        Image(systemName: symbol(for: place.kind))
                            .font(.system(size: metrics.points(11), weight: .bold))
                            .foregroundStyle(.white)
                            .frame(width: metrics.points(24), height: metrics.points(24))
                            .background(color(for: place.kind), in: Circle())
                            .overlay(Circle().strokeBorder(.white, lineWidth: 2.5))
                            .shadow(color: Palette.navy.opacity(0.3), radius: 4, y: 3)
                    }
                }
            }
            .mapStyle(.standard(elevation: .flat, pointsOfInterest: .excludingAll))
            .frame(height: metrics.points(210))
            .overlay(alignment: .bottomTrailing) {
                Text(openMap)
                    .typeStyle(Typography.captionEmphasis)
                    .foregroundStyle(Palette.label)
                    .padding(.horizontal, metrics.points(12))
                    .frame(height: metrics.points(32))
                    .glassCapsule(.light)
                    .padding(metrics.points(12))
            }
            .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.mapPreview), style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: metrics.points(Radius.mapPreview), style: .continuous).strokeBorder(Palette.labelAlpha(0.08), lineWidth: 0.5))
            .accessibilityLabel("Map of \(guide.places.map(\.name).joined(separator: ", "))")
        }
    }

    private func symbol(for kind: MapPlace.Kind) -> String {
        switch kind {
        case .stay: "bed.double.fill"
        case .attraction: "star.fill"
        case .restaurant: "fork.knife"
        case .saved: "bookmark.fill"
        }
    }

    private func color(for kind: MapPlace.Kind) -> Color {
        switch kind {
        case .stay: Palette.twilight
        case .attraction: Palette.gold
        case .restaurant: Palette.sky
        case .saved: Palette.aqua
        }
    }
}
