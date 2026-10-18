import NuvroDesignSystem
import NuvroModels
import SwiftUI

public struct SavedView: View {
    @Environment(\.designMetrics) private var metrics
    @Bindable private var model: SavedModel
    private let onCompare: () -> Void
    private let onExplore: () -> Void

    public init(model: SavedModel, onCompare: @escaping () -> Void, onExplore: @escaping () -> Void = {}) {
        self.model = model
        self.onCompare = onCompare
        self.onExplore = onExplore
    }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(16)) {
                SegmentedChoice(options: model.copy.tabs, selection: $model.selectedTab) { $0 }
                if model.isDestinationsTab, model.destinations.isEmpty {
                    NuvroEmptyState(symbol: "bookmark", title: model.copy.emptyTitle, message: model.copy.emptyBody, action: (model.copy.explore, onExplore)).padding(.top, metrics.points(30))
                } else if model.isDestinationsTab {
                    Text(model.selectionHint).typeStyle(TypeStyle(size: 12.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.55))
                    ForEach(model.destinations) { destination in SavedDestinationCard(destination: destination, model: model) }
                } else if model.places.isEmpty {
                    NuvroEmptyState(symbol: "mappin", title: model.copy.emptyPlacesTitle, message: model.copy.emptyPlacesBody).padding(.top, metrics.points(30))
                } else {
                    VStack(spacing: 0) {
                        ForEach(model.places) { place in
                            HStack(spacing: metrics.points(12)) {
                                DynamicDestinationImage(media: place.media.descriptor, labelSize: 6).frame(width: metrics.points(56), height: metrics.points(52)).clipShape(RoundedRectangle(cornerRadius: metrics.points(12), style: .continuous))
                                VStack(alignment: .leading, spacing: metrics.points(2)) {
                                    Text(place.name).typeStyle(TypeStyle(size: 15, weight: .semibold)).foregroundStyle(Palette.label)
                                    Text(place.subtitle).typeStyle(TypeStyle(size: 12, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
                                }
                                Spacer(minLength: 0)
                                ChevronGlyph()
                            }
                            .padding(.horizontal, metrics.points(13)).padding(.vertical, metrics.points(10))
                            if place.id != model.places.last?.id { Divider().overlay(Palette.labelAlpha(0.07)) }
                        }
                    }
                    .surfaceCard()
                }
            }
            .padding(.horizontal, metrics.points(20)).padding(.top, metrics.points(8)).padding(.bottom, metrics.points(40))
            .animation(Motion.fade(0.3), value: model.selectedTab)
            .animation(Motion.fade(0.25), value: model.selectedIDs)
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button(action: onCompare) {
                    Text(model.compareLabel).typeStyle(Typography.control).foregroundStyle(model.canCompare ? Palette.sky : Palette.labelAlpha(0.5))
                        .padding(.horizontal, metrics.points(6))
                }
                .disabled(!model.canCompare)
            }
        }
        .screenTitle(model.copy.title)
    }
}
struct SavedDestinationCard: View {
    @Environment(\.designMetrics) private var metrics
    let destination: SavedDestination
    let model: SavedModel

    var body: some View {
        let isSelected = model.isSelected(destination)
        Button { model.toggle(destination) } label: {
            VStack(spacing: 0) {
                DynamicDestinationImage(media: destination.media.descriptor, label: "photo · \(destination.name.lowercased())", labelSize: 7)
                    .frame(height: metrics.points(150))
                    .photoScrim(LinearGradient(colors: [.clear, Palette.navy.opacity(0.5)], startPoint: .center, endPoint: .bottom))
                    .overlay(alignment: .bottomLeading) {
                        VStack(alignment: .leading, spacing: metrics.points(2)) {
                            Text(destination.name).typeStyle(TypeStyle(size: 19, weight: .bold, tracking: -0.3)).foregroundStyle(Palette.cloud)
                            Text(destination.country).typeStyle(TypeStyle(size: 12.5, weight: .regular)).foregroundStyle(Palette.cloudAlpha(0.85))
                        }
                        .padding(EdgeInsets(top: 0, leading: metrics.points(14), bottom: metrics.points(12), trailing: metrics.points(14)))
                    }
                    .overlay(alignment: .topTrailing) {
                        HStack(spacing: metrics.points(6)) {
                            Text("\(destination.matchPercent)%").typeStyle(TypeStyle(size: 13, weight: .bold)).foregroundStyle(Palette.cloud)
                            Text(model.copy.match).typeStyle(TypeStyle(size: 8.5, weight: .bold, tracking: 0.85)).foregroundStyle(Palette.cloudAlpha(0.85))
                        }
                        .padding(.horizontal, metrics.points(11)).frame(height: metrics.points(30)).glassCapsule(.control).padding(metrics.points(14))
                    }
                    .overlay(alignment: .bottomTrailing) {
                        Text(isSelected ? "✓" : "+").typeStyle(TypeStyle(size: 12, weight: .bold)).foregroundStyle(.white)
                            .frame(width: metrics.points(26), height: metrics.points(26))
                            .background(isSelected ? Palette.sky : Palette.labelAlpha(0.35), in: Circle())
                            .overlay(Circle().strokeBorder(.white, lineWidth: 2))
                            .shadow(color: Palette.navy.opacity(0.3), radius: 4, y: 2)
                            .padding(EdgeInsets(top: 0, leading: 0, bottom: metrics.points(12), trailing: metrics.points(14)))
                    }
                VStack(alignment: .leading, spacing: metrics.points(2)) {
                    Text(destination.reason).typeStyle(TypeStyle(size: 13, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.7))
                    Text(destination.note).typeStyle(TypeStyle(size: 11.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.45))
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(EdgeInsets(top: metrics.points(11), leading: metrics.points(14), bottom: metrics.points(13), trailing: metrics.points(14)))
            }
            .background(Palette.surface, in: RoundedRectangle(cornerRadius: metrics.points(22), style: .continuous))
            .clipShape(RoundedRectangle(cornerRadius: metrics.points(22), style: .continuous))
            .overlay(RoundedRectangle(cornerRadius: metrics.points(22), style: .continuous).strokeBorder(isSelected ? Palette.sky.opacity(0.5) : Palette.labelAlpha(0.06), lineWidth: isSelected ? 2 : 0.5))
        }
        .buttonStyle(PressScaleButtonStyle(scale: 0.99))
        .accessibilityAddTraits(isSelected ? .isSelected : [])
    }
}

extension DestinationMedia {
    var descriptor: MediaDescriptor { MediaDescriptor(url: url, placeholder: placeholder, altText: altText, attribution: attribution) }
}
