import MapKit
import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import NuvroPermissions
import SwiftUI

public struct ExploreNearbyView: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(PermissionCenter.self) private var permissions: PermissionCenter?
    @Environment(TransientFeedback.self) private var feedback: TransientFeedback?
    @Bindable private var model: ExploreNearbyModel
    private let permissionCopy = PermissionCopy.resolve(.english)
    private let centre: GeoCoordinate
    private let onNext: () -> Void

    public init(model: ExploreNearbyModel, centre: GeoCoordinate, onNext: @escaping () -> Void) {
        self.model = model
        self.centre = centre
        self.onNext = onNext
    }

    public var body: some View {
        ZStack(alignment: .top) {
            NearbyMap(model: model, centre: centre).ignoresSafeArea(edges: .bottom)
            if model.isListPresentation {
                placeList.transition(.move(edge: .bottom).combined(with: .opacity))
            }
            VStack(alignment: .leading, spacing: metrics.points(10)) {
                searchEntry
                filters
                freeTimeCard
            }
            .padding(.horizontal, metrics.points(20))
            .padding(.top, metrics.points(4))
        }
        .background(Palette.background.ignoresSafeArea())
        .safeAreaInset(edge: .bottom) {
            if !model.hasResults, !model.isListPresentation {
                NuvroEmptyState(symbol: "mappin.slash", title: model.copy.noResultsTitle, message: model.copy.noResultsBody, action: (model.copy.clearFilters, { model.filter = nil }))
                    .frame(height: metrics.points(230))
                    .surfaceCard(radius: Radius.hero)
                    .padding(.horizontal, metrics.points(12)).padding(.bottom, metrics.points(8))
                    .transition(.opacity)
            } else if let selected = model.selected, !model.isListPresentation {
                NearbyPlaceCard(place: selected, model: model)
                    .id(selected.id)
                    .padding(.horizontal, metrics.points(12))
                    .padding(.bottom, metrics.points(8))
                    .transition(.move(edge: .bottom).combined(with: .opacity))
            }
        }
        .animation(Motion.settle(0.5), value: model.selectedID)
        .animation(Motion.settle(0.5), value: model.isListPresentation)
        .animation(Motion.fade(0.3), value: model.filter)
        .toolbar {
            ToolbarSwitcher(options: [false, true], selection: $model.isListPresentation, width: 118) { $0 ? model.copy.list : model.copy.map }
        }
        .task { if let permissions, !model.usesDestinationCentre { _ = await permissions.prepare(.location) } }
        .permissionPresentation(
            permissions, for: .location,
            content: PermissionExplainerContent(symbol: "location.fill", title: permissionCopy.locationTitle, message: permissionCopy.locationBody, deniedTitle: permissionCopy.locationDeniedTitle, deniedMessage: permissionCopy.locationDeniedBody, allow: permissionCopy.allow, notNow: permissionCopy.notNow, openSettings: permissionCopy.openSettings),
            fallback: (permissionCopy.enterLocation, { model.usesDestinationCentre = true })
        )
        .onChange(of: model.lastEvent) { _, event in if let event { feedback?.show(event); model.clearEvent() } }
    }

    private var searchEntry: some View {
        HStack(spacing: metrics.points(10)) {
            RouteArcGlyph(width: metrics.points(16))
            Text(model.copy.searchPlaceholder).typeStyle(Typography.body).foregroundStyle(Palette.labelAlpha(0.55)).lineLimit(1)
            Spacer(minLength: 0)
        }
        .padding(.horizontal, metrics.points(16))
        .frame(height: metrics.points(46))
        .glassCapsule(.light)
    }

    private var filters: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: metrics.points(7)) {
                ForEach(NearbyPlace.Category.allCases, id: \.self) { category in
                    let selected = model.filter == category
                    Button { model.toggleFilter(category) } label: {
                        Text(model.label(for: category))
                            .typeStyle(TypeStyle(size: 13.5, weight: .semibold))
                            .foregroundStyle(selected ? Palette.background : Palette.label)
                            .padding(.horizontal, metrics.points(13)).frame(height: metrics.points(34))
                            .background(selected ? AnyShapeStyle(Palette.navy) : AnyShapeStyle(Palette.surface.opacity(0.72)), in: Capsule())
                            .overlay(Capsule().strokeBorder(selected ? Palette.label : Palette.labelAlpha(0.1), lineWidth: 0.5))
                            .shadow(color: .clear, radius: 0)
                    }
                    .buttonStyle(PressScaleButtonStyle(scale: 0.96))
                    .accessibilityAddTraits(selected ? .isSelected : [])
                }
            }
            .padding(.horizontal, metrics.points(20))
            .padding(.vertical, metrics.points(2))
        }
        .scrollClipDisabled()
        .padding(.horizontal, -metrics.points(20))
    }

    private var freeTimeCard: some View {
        HStack(spacing: metrics.points(9)) {
            RouteArcGlyph(width: metrics.points(16))
            VStack(alignment: .leading, spacing: metrics.points(1)) {
                Text("\(model.copy.freeTime) \(model.freeTimeLabel) \(model.copy.free)").typeStyle(TypeStyle(size: 13, weight: .bold)).foregroundStyle(Palette.label)
                Text(model.copy.theseFit).typeStyle(TypeStyle(size: 11.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.62))
            }
        }
        .padding(.horizontal, metrics.points(13)).padding(.vertical, metrics.points(9))
        .glass(.light, in: RoundedRectangle(cornerRadius: metrics.points(16), style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: metrics.points(16), style: .continuous).strokeBorder(Palette.sky.opacity(0.25), lineWidth: 0.5))
        .accessibilityElement(children: .combine)
    }

    private var placeList: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(12)) {
                Text("\(model.visiblePlaces.count) \(model.copy.placesNearYou)").typeStyle(TypeStyle(size: 17, weight: .bold, tracking: -0.2)).foregroundStyle(Palette.label)
                ForEach(model.visiblePlaces) { place in
                    NearbyPlaceRow(place: place, model: model) { model.select(place); model.isListPresentation = false }
                }
                LinkButton("Tickets & documents", action: onNext).frame(maxWidth: .infinity)
            }
            .padding(EdgeInsets(top: metrics.points(16), leading: metrics.points(20), bottom: metrics.points(130), trailing: metrics.points(20)))
        }
        .scrollIndicators(.hidden)
        .background(Palette.background, in: UnevenRoundedRectangle(topLeadingRadius: metrics.points(26), topTrailingRadius: metrics.points(26)))
        .shadow(color: Palette.labelAlpha(0.18), radius: metrics.points(20), y: -metrics.points(12))
        .padding(.top, metrics.points(176))
        .ignoresSafeArea(edges: .bottom)
    }
}

struct NearbyMap: View {
    @Environment(\.designMetrics) private var metrics
    @Bindable var model: ExploreNearbyModel
    let centre: GeoCoordinate

    var body: some View {
        Map(initialPosition: .region(MKCoordinateRegion(center: CLLocationCoordinate2D(latitude: centre.latitude, longitude: centre.longitude), span: MKCoordinateSpan(latitudeDelta: 0.03, longitudeDelta: 0.03))), interactionModes: [.zoom], selection: $model.selectedID) {
            Annotation("You", coordinate: CLLocationCoordinate2D(latitude: centre.latitude, longitude: centre.longitude)) {
                Circle().fill(Palette.sky).frame(width: metrics.points(14), height: metrics.points(14))
                    .overlay(Circle().strokeBorder(.white, lineWidth: 2.5))
                    .shadow(color: Palette.navy.opacity(0.35), radius: 3, y: 2)
            }
            ForEach(model.visiblePlaces) { place in
                Annotation("", coordinate: CLLocationCoordinate2D(latitude: place.coordinate.latitude, longitude: place.coordinate.longitude)) {
                    NuvroMapPin(letter: place.markerLetter, label: place.categoryLabel, colour: markerColour(place.category), isSelected: model.selectedID == place.id)
                }
                .tag(place.id)
            }
        }
        .mapStyle(.standard(elevation: .flat, pointsOfInterest: .excludingAll))
        .mapControlVisibility(.hidden)
    }

    private func markerColour(_ category: NearbyPlace.Category) -> Color {
        switch category {
        case .food: Palette.gold
        case .coffee: Palette.aqua
        case .thingsToDo, .indoor, .shopping, .saved: Palette.twilight
        case .views: Palette.coral
        }
    }
}

struct NuvroMapPin: View {
    @Environment(\.designMetrics) private var metrics
    let letter: String
    let label: String
    let colour: Color
    let isSelected: Bool

    var body: some View {
        HStack(spacing: metrics.points(5)) {
            Text(letter)
                .font(.system(size: metrics.points(11), weight: .bold, design: .rounded))
                .foregroundStyle(.white)
                .rotationEffect(.degrees(45))
                .frame(width: metrics.points(isSelected ? 30 : 24), height: metrics.points(isSelected ? 30 : 24))
                .background(colour, in: UnevenRoundedRectangle(topLeadingRadius: 15, bottomLeadingRadius: 0, bottomTrailingRadius: 15, topTrailingRadius: 15))
                .overlay(UnevenRoundedRectangle(topLeadingRadius: 15, bottomLeadingRadius: 0, bottomTrailingRadius: 15, topTrailingRadius: 15).strokeBorder(.white, lineWidth: 2.5))
                .shadow(color: Palette.navy.opacity(0.32), radius: 4, y: 3)
                .rotationEffect(.degrees(-45))
            if isSelected {
                Text(label).typeStyle(TypeStyle(size: 11, weight: .semibold)).foregroundStyle(Palette.label)
                    .padding(.horizontal, metrics.points(6)).padding(.vertical, metrics.points(2))
                    .background(Color.white.opacity(0.9), in: RoundedRectangle(cornerRadius: metrics.points(6), style: .continuous))
            }
        }
        .scaleEffect(isSelected ? 1.06 : 1)
        .animation(Motion.settle(0.3), value: isSelected)
        .accessibilityLabel(label)
    }
}
