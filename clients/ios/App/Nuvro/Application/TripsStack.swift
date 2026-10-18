import NuvroConnectivity
import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import NuvroPlan
import NuvroTravel
import SwiftUI

struct TripsStack: View {
    @Environment(ConnectivityMonitor.self) private var connectivity
    @Binding var path: [TripRoute]
    let tripPlan: TripPlanModel
    let travelMode: TravelModeModel
    let language: Language
    let onPlanTrip: () -> Void
    @State private var library = TripsLibraryModel(entries: PrototypeLibrary.entries)
    @State private var itinerary: ItineraryModel
    @State private var stays = StayModel(options: PrototypeBudget.stays, locale: PrototypeContent.locale)
    @State private var companion = LanguageCompanionModel(pair: PrototypeCompanion.languagePair, sample: PrototypeCompanion.sampleTranslation, phrases: PrototypeCompanion.phrases)
    @State private var nearby = ExploreNearbyModel(places: PrototypeCompanion.nearbyPlaces, freeTime: PrototypeCompanion.freeTime, locationName: "Funchal", destinationName: PrototypeContent.madeira.name)
    @State private var wallet = TripWalletModel(documents: PrototypeCompanion.documents)
    @State private var alerts = TripAlertsModel(alerts: PrototypeCompanion.alerts, preferences: PrototypeLibrary.notifications)
    @State private var flights = FlightsModel(options: PrototypeLibrary.flights, locale: PrototypeContent.locale)
    @State private var experiences = ExperiencesModel(hero: PrototypeLibrary.heroExperience, experiences: PrototypeLibrary.experiences)
    @State private var transport = TransportModel(guide: PrototypeLibrary.transport)
    @State private var checklist = ChecklistModel(groups: PrototypeLibrary.checklist, insight: PrototypeLibrary.checklistInsight)
    @State private var group = GroupModel(members: PrototypeLibrary.members, decisions: PrototypeLibrary.decisions, changes: PrototypeLibrary.changes, budgetLabel: "$4,820", budgetNote: "estimated · 3 travellers")
    @State private var practical = PracticalModel(entryCheckedLabel: "Checked 2 days ago", entryRows: PrototypeLibrary.entryRows, sections: PrototypeLibrary.infoSections, emergency: PrototypeLibrary.emergency, openSection: "Money")
    @State private var summary = TripSummaryModel(summary: PrototypeCompanion.summary, loved: ["Coastal places", "Local food", "Relaxed mornings"], notForMe: ["Long transfers"], locale: PrototypeContent.locale)

    init(path: Binding<[TripRoute]>, tripPlan: TripPlanModel, travelMode: TravelModeModel, language: Language, onPlanTrip: @escaping () -> Void) {
        self._path = path
        self.tripPlan = tripPlan
        self.travelMode = travelMode
        self.language = language
        self.onPlanTrip = onPlanTrip
        self._itinerary = State(initialValue: ItineraryModel(plan: { tripPlan.plan }, language: language))
    }

    var body: some View {
        NavigationStack(path: $path) {
            TripsLibraryView(model: library, onContinueTrip: { path.append(.travelMode) }, onOpenPlan: { path.append(.hub) }, onOpenPast: { path.append(.summary) }, onPlanTrip: onPlanTrip)
                .navigationDestination(for: TripRoute.self) { route in
                    destination(for: route)
                }
        }
        .onChange(of: connectivity.status, initial: true) { _, status in
            travelMode.weatherFreshness = status == .online ? .fresh : .cached
        }
    }

    private var hub: some View {
        TripPlanHubView(
            model: tripPlan,
            onViewItinerary: { path.append(.itinerary) },
            onViewBudget: { path.append(.budget) },
            onSelectModule: { module in
                switch module {
                case .stay: path.append(.stay)
                case .flights: path.append(.flights)
                case .thingsToDo: path.append(.experiences)
                case .gettingAround: path.append(.transport)
                case .packing: path.append(.checklist)
                case .documents, .esim, .weather: path.append(.practical)
                }
            },
            onTravelers: { path.append(.group) }
        )
    }

    @ViewBuilder
    private func destination(for route: TripRoute) -> some View {
        switch route {
        case .hub:
            hub
        case .flights:
            FlightsView(model: flights, routeLabel: "New York → Madeira · \(tripPlan.plan.dateRangeLabel)", travelers: tripPlan.plan.travelersLabel)
        case .experiences:
            ExperiencesView(model: experiences, subtitle: "\(tripPlan.plan.destination.name) · \(tripPlan.plan.dateRangeLabel)")
        case .transport:
            TransportView(model: transport, subtitle: "\(tripPlan.plan.destination.name) · \(tripPlan.plan.dateRangeLabel)")
        case .checklist:
            ChecklistView(model: checklist, subtitle: "\(tripPlan.plan.destination.name) · 6 days to go")
        case .group:
            GroupPlanningView(model: group, subtitle: "\(tripPlan.plan.destination.name) · \(tripPlan.plan.dateRangeLabel)")
        case .practical:
            PracticalInfoView(model: practical, subtitle: "\(tripPlan.plan.destination.name) · \(tripPlan.plan.destination.country)")
        case .itinerary:
            FullItineraryView(model: itinerary, statusText: tripPlan.copy.planning, onOptimize: { path.append(.budget) }, onViewBudget: { path.append(.budget) })
                .toolbar(.hidden, for: .tabBar)
        case .budget:
            BudgetView(model: tripPlan, language: language, onViewStays: { path.append(.stay) })
                .toolbar(.hidden, for: .tabBar)
        case .stay:
            StayView(model: stays, destinationName: tripPlan.plan.destination.name, dateRange: tripPlan.plan.dateRangeLabel, travelers: tripPlan.plan.travelersLabel, activityCentre: PrototypeContent.madeiraGuide.centre) { _ in
                path = [.travelMode]
            }
            .toolbar(.hidden, for: .tabBar)
        case .travelMode:
            TravelModeHomeView(
                model: travelMode,
                actions: TravelModeActions(
                    fullItinerary: { path.append(.itinerary) }, reviewPlan: { path.append(.replan) },
                    navigate: { path.append(.nearby) }, tickets: { path.append(.wallet) }, translate: { path.append(.language) }, updates: { path.append(.alerts) }
                ),
                updateCount: alerts.activeCount
            )
            .navigationBarBackButtonHidden(true)
        case .language:
            LanguageCompanionView(model: companion, dayLabel: "Day \(travelMode.dayNumber) · \(travelMode.destinationName)") { path.append(.nearby) }
                .toolbar(.hidden, for: .tabBar)
        case .nearby:
            ExploreNearbyView(model: nearby, centre: GeoCoordinate(latitude: 32.6485, longitude: -16.9080)) { path.append(.wallet) }
        case .wallet:
            TripWalletView(model: wallet, destinationName: tripPlan.plan.destination.name, dateRange: tripPlan.plan.dateRangeLabel) { path.append(.alerts) }
                .toolbar(.hidden, for: .tabBar)
        case .alerts:
            TripAlertsView(model: alerts) { destination in
                switch destination {
                case .replanning: path.append(.replan)
                case .travelMode: path = [.travelMode]
                case .document: path.append(.wallet)
                case .summary: path.append(.summary)
                }
            }
        case .summary:
            TripSummaryView(model: summary, onPlanAnother: { path = [] }, onViewTrip: { path = [.travelMode] })
                .toolbar(.hidden, for: .tabBar)
        case .replan:
            LiveReplanningView(model: travelMode) {
                if let applied = travelMode.appliedAdjustment { tripPlan.apply(applied) }
                path.removeLast()
            }
            .toolbar(.hidden, for: .tabBar)
        }
    }
}
