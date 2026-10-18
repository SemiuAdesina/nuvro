import MapKit
import NuvroDesignSystem
import NuvroModels
import SwiftUI

public struct StayView: View {
    @Environment(\.designMetrics) private var metrics
    @Bindable private var model: StayModel
    private let dateRange: String
    private let travelers: String
    private let destinationName: String
    private let activityCentre: GeoCoordinate
    private let onViewStay: (StayOption) -> Void

    public init(model: StayModel, destinationName: String, dateRange: String, travelers: String, activityCentre: GeoCoordinate, onViewStay: @escaping (StayOption) -> Void) {
        self.model = model
        self.destinationName = destinationName
        self.dateRange = dateRange
        self.travelers = travelers
        self.activityCentre = activityCentre
        self.onViewStay = onViewStay
    }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(16)) {
                Text(travelers).typeStyle(Typography.metaEmphasis).foregroundStyle(Palette.labelAlpha(0.7))
                    .padding(.horizontal, metrics.points(12)).frame(height: metrics.points(32))
                    .background(Palette.labelAlpha(0.06), in: Capsule())
                filters
                if model.presentation == .map {
                    StayMap(options: model.options, activityCentre: activityCentre, activitiesLabel: model.copy.yourActivities)
                        .transition(.opacity)
                }
                if let hero = model.hero {
                    HeroStayCard(option: hero, model: model, onView: { onViewStay(hero) })
                }
                VStack(alignment: .leading, spacing: metrics.points(10)) {
                    Text(model.copy.otherFits).typeStyle(Typography.cardHeading).foregroundStyle(Palette.label).padding(.top, metrics.points(2))
                    ForEach(model.others) { option in
                        StayRow(option: option, model: model) { model.filter = option.recommendation }
                    }
                }
                Text(model.copy.disclosure).typeStyle(Typography.small).foregroundStyle(Palette.labelAlpha(0.45)).multilineTextAlignment(.center).frame(maxWidth: .infinity)
                InlineActions {
                    CallToActionButton(model.copy.compare) { model.presentation = .map }
                }
            }
            .padding(.horizontal, metrics.points(20))
            .padding(.top, metrics.points(8))
            .padding(.bottom, metrics.points(40))
            .animation(Motion.fade(0.3), value: model.presentation)
            .animation(Motion.fade(0.3), value: model.filter)
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .screenTitle(model.copy.title, subtitle: "\(destinationName) · \(dateRange)")
        .toolbar {
            ToolbarSwitcher(options: StayPresentation.allCases, selection: $model.presentation) { presentation in
                presentation == .list ? model.copy.list : model.copy.map
            }
        }
    }

    private var presentationSwitcher: some View {
        SegmentedChoice(options: StayPresentation.allCases, selection: $model.presentation, compact: true, isPill: true) { presentation in
            presentation == .list ? model.copy.list : model.copy.map
        }
        .frame(width: metrics.points(136))
    }

    private var filters: some View {
        ScrollView(.horizontal, showsIndicators: false) {
            HStack(spacing: metrics.points(8)) {
                ForEach(StayOption.Recommendation.allCases, id: \.self) { recommendation in
                    let selected = model.filter == recommendation
                    Button { model.filter = recommendation } label: {
                        Text(model.label(for: recommendation))
                            .typeStyle(Typography.control)
                            .foregroundStyle(selected ? Palette.background : Palette.label)
                            .padding(.horizontal, metrics.points(14)).frame(height: metrics.points(36))
                            .background(selected ? Palette.label : Palette.surface, in: Capsule())
                            .overlay(Capsule().strokeBorder(selected ? Palette.label : Palette.labelAlpha(0.1), lineWidth: 0.5))
                    }
                    .buttonStyle(PressScaleButtonStyle(scale: 0.96))
                    .accessibilityAddTraits(selected ? .isSelected : [])
                }
            }
            .padding(.horizontal, metrics.points(20))
        }
        .scrollClipDisabled()
        .padding(.horizontal, -metrics.points(20))
    }
}

struct StayMap: View {
    @Environment(\.designMetrics) private var metrics
    let options: [StayOption]
    let activityCentre: GeoCoordinate
    let activitiesLabel: String

    var body: some View {
        Map(initialPosition: .region(MKCoordinateRegion(center: CLLocationCoordinate2D(latitude: activityCentre.latitude, longitude: activityCentre.longitude - 0.02), span: MKCoordinateSpan(latitudeDelta: 0.06, longitudeDelta: 0.12))), interactionModes: []) {
            MapCircle(center: CLLocationCoordinate2D(latitude: activityCentre.latitude, longitude: activityCentre.longitude), radius: 900)
                .foregroundStyle(Palette.gold.opacity(0.18))
                .stroke(Palette.gold.opacity(0.7), style: StrokeStyle(lineWidth: 1, dash: [4, 4]))
            Annotation(activitiesLabel, coordinate: CLLocationCoordinate2D(latitude: activityCentre.latitude + 0.006, longitude: activityCentre.longitude)) {
                EmptyView()
            }
            ForEach(options) { option in
                Annotation(option.name, coordinate: CLLocationCoordinate2D(latitude: option.coordinate.latitude, longitude: option.coordinate.longitude)) {
                    Image(systemName: "bed.double.fill")
                        .font(.system(size: metrics.points(10), weight: .bold))
                        .foregroundStyle(.white)
                        .frame(width: metrics.points(24), height: metrics.points(24))
                        .background(option.recommendation == .bestOverall ? Palette.sky : Palette.twilight, in: Circle())
                        .overlay(Circle().strokeBorder(.white, lineWidth: 2.5))
                        .shadow(color: Palette.navy.opacity(0.3), radius: 4, y: 3)
                }
            }
        }
        .mapStyle(.standard(elevation: .flat, pointsOfInterest: .excludingAll))
        .frame(height: metrics.points(300))
        .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.mapPreview), style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: metrics.points(Radius.mapPreview), style: .continuous).strokeBorder(Palette.labelAlpha(0.08), lineWidth: 0.5))
    }
}
