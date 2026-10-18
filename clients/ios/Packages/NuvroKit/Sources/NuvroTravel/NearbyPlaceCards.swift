import NuvroDesignSystem
import NuvroModels
import SwiftUI

struct NearbyPlaceCard: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(\.openURL) private var openURL
    let place: NearbyPlace
    let model: ExploreNearbyModel

    var body: some View {
        VStack(spacing: 0) {
            Capsule().fill(Palette.labelAlpha(0.18)).frame(width: metrics.points(36), height: metrics.points(5)).padding(.top, metrics.points(8))
            VStack(alignment: .leading, spacing: metrics.points(12)) {
                HStack(alignment: .top, spacing: metrics.points(12)) {
                    DynamicDestinationImage(media: place.media.descriptor, label: "", labelSize: 7)
                        .frame(width: metrics.points(92), height: metrics.points(82))
                        .clipShape(RoundedRectangle(cornerRadius: metrics.points(16), style: .continuous))
                    VStack(alignment: .leading, spacing: metrics.points(4)) {
                        Text(place.name).typeStyle(TypeStyle(size: 19, weight: .bold, tracking: -0.3)).foregroundStyle(Palette.label)
                        HStack(spacing: metrics.points(8)) {
                            Text(place.travelLabel).typeStyle(TypeStyle(size: 12.5, weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.7))
                            Text("✓ \(place.hoursLabel)").typeStyle(TypeStyle(size: 12.5, weight: .semibold)).foregroundStyle(Palette.matchTeal)
                        }
                        Text(place.reason).typeStyle(TypeStyle(size: 12.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.65))
                        if let attribution = place.media.attribution {
                            Text(attribution).typeStyle(TypeStyle(size: 9.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.4))
                        }
                    }
                    Spacer(minLength: 0)
                }
                HStack(spacing: metrics.points(8)) {
                    Button { if let url = model.directionsURL(for: place) { openURL(url) } } label: {
                        Text(model.copy.go).typeStyle(TypeStyle(size: 15.5, weight: .semibold)).foregroundStyle(.white).frame(maxWidth: .infinity).frame(height: metrics.points(46))
                    }
                    .buttonStyle(GlassCallToActionStyle())
                    Button { model.addToToday(place) } label: {
                        Text(model.isAdded(place) ? model.copy.addedToToday : model.copy.addToToday)
                            .typeStyle(TypeStyle(size: 15, weight: .semibold)).foregroundStyle(model.isAdded(place) ? Palette.matchTeal : Palette.label)
                            .frame(maxWidth: .infinity).frame(height: metrics.points(46))
                            .background(model.isAdded(place) ? Palette.aqua.opacity(0.22) : Palette.labelAlpha(0.06), in: Capsule())
                            .overlay(Capsule().strokeBorder(Palette.labelAlpha(0.1), lineWidth: 0.5))
                    }
                    .buttonStyle(PressScaleButtonStyle(scale: 0.97))
                    Button { model.toggleSaved(place) } label: {
                        Image(systemName: model.isSaved(place) ? "bookmark.fill" : "bookmark")
                            .font(.system(size: metrics.points(14), weight: .semibold)).foregroundStyle(Palette.label)
                            .frame(width: metrics.points(46), height: metrics.points(46))
                            .background(Palette.labelAlpha(0.06), in: Circle())
                            .overlay(Circle().strokeBorder(Palette.labelAlpha(0.1), lineWidth: 0.5))
                    }
                    .buttonStyle(PressScaleButtonStyle(scale: 0.95))
                    .accessibilityLabel(model.isSaved(place) ? "Saved" : "Save")
                }
            }
            .padding(EdgeInsets(top: metrics.points(12), leading: metrics.points(16), bottom: metrics.points(16), trailing: metrics.points(16)))
        }
        .background(Palette.surface, in: RoundedRectangle(cornerRadius: metrics.points(26), style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: metrics.points(26), style: .continuous).strokeBorder(Palette.labelAlpha(0.06), lineWidth: 0.5))
        .shadow(color: Palette.labelAlpha(0.16), radius: metrics.points(18), y: metrics.points(8))
    }
}

struct NearbyPlaceRow: View {
    @Environment(\.designMetrics) private var metrics
    let place: NearbyPlace
    let model: ExploreNearbyModel
    let onSelect: () -> Void

    var body: some View {
        Button(action: onSelect) {
            HStack(spacing: metrics.points(12)) {
                DynamicDestinationImage(media: place.media.descriptor, labelSize: 7)
                    .frame(width: metrics.points(78), height: metrics.points(70))
                    .clipShape(RoundedRectangle(cornerRadius: metrics.points(14), style: .continuous))
                VStack(alignment: .leading, spacing: metrics.points(3)) {
                    Text(place.name).typeStyle(TypeStyle(size: 15, weight: .semibold)).foregroundStyle(Palette.label)
                    Text("\(place.travelLabel) · \(place.categoryLabel)").typeStyle(TypeStyle(size: 12, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
                    Text(place.reason).typeStyle(TypeStyle(size: 11.5, weight: .semibold)).foregroundStyle(Palette.matchTeal)
                }
                Spacer(minLength: 0)
                Image(systemName: "chevron.forward").font(.system(size: metrics.points(12), weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.3))
            }
            .padding(metrics.points(10))
            .surfaceCard(radius: Radius.cardCompact)
        }
        .buttonStyle(PressScaleButtonStyle(scale: 0.99))
        .accessibilityAddTraits(model.selected?.id == place.id ? .isSelected : [])
    }
}

extension DestinationMedia {
    var descriptor: MediaDescriptor {
        MediaDescriptor(url: url, placeholder: placeholder, altText: altText, attribution: attribution)
    }
}
