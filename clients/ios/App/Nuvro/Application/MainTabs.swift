import NuvroAPI
import NuvroDesignSystem
import NuvroDestination
import NuvroDiscover
import NuvroLocalization
import NuvroModels
import NuvroPlan
import NuvroTravel
import SwiftUI

struct MainTabs: View {
    @Binding var selectedTab: AppTab
    @Binding var discoverPath: [AppRoute]
    @Binding var tripsPath: [TripRoute]
    let tripPlan: TripPlanModel
    let travelMode: TravelModeModel
    @Binding var hasStartedTrip: Bool
    let language: Language
    @Binding var travelProfile: TravelProfile?
    let account: (any AccountGateway)?
    let discovery: (any DiscoverGateway)?
    @State private var discover: DiscoverHomeModel
    @State private var setup: MatchSetupModel

    init(selectedTab: Binding<AppTab>, discoverPath: Binding<[AppRoute]>, tripsPath: Binding<[TripRoute]>, tripPlan: TripPlanModel, travelMode: TravelModeModel, hasStartedTrip: Binding<Bool>, language: Language, travelProfile: Binding<TravelProfile?>, account: (any AccountGateway)? = nil, discovery: (any DiscoverGateway)? = nil) {
        self._selectedTab = selectedTab
        self._discoverPath = discoverPath
        self._tripsPath = tripsPath
        self.tripPlan = tripPlan
        self.travelMode = travelMode
        self._hasStartedTrip = hasStartedTrip
        self.language = language
        self._travelProfile = travelProfile
        self.account = account
        self.discovery = discovery
        let profile = travelProfile.wrappedValue
        if let discovery {
            let brief = DiscoverBriefs.home(profile: profile)
            self._discover = State(initialValue: DiscoverHomeModel(state: .loading, source: { try await discovery.matches(brief) }, language: language, locale: PrototypeContent.locale))
        } else {
            self._discover = State(initialValue: DiscoverHomeModel(suggestions: PrototypeContent.suggestions, language: language, locale: PrototypeContent.locale))
        }
        let currency = profile?.preferredCurrency ?? "USD"
        self._setup = State(initialValue: MatchSetupModel(
            airports: PrototypeContent.airports, departureCode: profile?.preferredDepartureAirport ?? "JFK", language: language, currency: currency,
            locale: PrototypeContent.locale, budgetScale: PrototypeContent.budgetScale(for: currency), interests: profile?.interests
        ))
    }
    @Namespace private var heroNamespace

    var body: some View {
        TabView(selection: $selectedTab) {
            Tab("Discover", systemImage: "safari.fill", value: .discover) { discoverStack }
            Tab("Trips", systemImage: "suitcase.fill", value: .trips) { tripsTab }
            Tab("Saved", systemImage: "bookmark.fill", value: .saved) { SavedStack(onExplore: { selectedTab = .discover; discoverPath = [] }) }
            Tab("Profile", systemImage: "person.fill", value: .profile) { ProfileStack(travelProfile: $travelProfile, account: account) }
        }
        .tint(Palette.sky)
    }

    private var discoverStack: some View {
        NavigationStack(path: $discoverPath) {
            DiscoverHomeView(
                model: discover,
                onStartSearch: { discoverPath.append(.matchSetup) },
                onExploreByBudget: { discoverPath.append(.matchSetup) },
                onSeeAll: { discoverPath.append(.matchResults) },
                onOpen: { discoverPath.append(.destination($0)) }
            )
            .navigationDestination(for: AppRoute.self) { route in
                destination(for: route)
                    .toolbar(.hidden, for: .tabBar)
            }
        }
        .task { await discover.refresh() }
    }

    private func findMatches() {
        guard let discovery else { discoverPath.append(.matchResults); return }
        let brief = DiscoverBriefs.search(setup, profile: travelProfile)
        Task { await discover.load(from: { try await discovery.matches(brief) }) }
        discoverPath.append(.matchResults)
    }

    private var results: [DestinationMatch] {
        discovery == nil ? PrototypeContent.matches : discover.suggestions
    }

    @ViewBuilder
    private func destination(for route: AppRoute) -> some View {
        switch route {
        case .matchSetup:
            MatchSetupView(model: setup, onFindMatches: findMatches)
        case .matchResults:
            MatchResultsView(matches: results, language: language, locale: PrototypeContent.locale, heroNamespace: heroNamespace, onEditPreferences: { discoverPath.removeLast() }, onExplore: { discoverPath.append(.destination($0)) })
        case .destination(let match):
            DestinationDetailView(match: match, guide: PrototypeContent.madeiraGuide, budgetCeiling: setup.budget, estimate: PrototypeContent.madeiraPlan.estimatedTotal, language: language, locale: PrototypeContent.locale) {
                hasStartedTrip = true
                discoverPath.append(.tripPlan)
            }
            .navigationTransition(.zoom(sourceID: match.id, in: heroNamespace))
        case .tripPlan:
            TripPlanHubView(model: tripPlan, onViewItinerary: { selectedTab = .trips; discoverPath = []; tripsPath = [.itinerary] }, onViewBudget: { selectedTab = .trips; discoverPath = []; tripsPath = [.budget] }, onSelectModule: { module in if module == .stay { selectedTab = .trips; discoverPath = []; tripsPath = [.stay] } })
        }
    }

    @ViewBuilder
    private var tripsTab: some View {
        if hasStartedTrip {
            TripsStack(path: $tripsPath, tripPlan: tripPlan, travelMode: travelMode, language: language, onPlanTrip: { selectedTab = .discover; discoverPath = [.matchSetup] })
        } else {
            NavigationStack {
                NuvroEmptyState(symbol: "suitcase", title: LibraryCopy.resolve(language).emptyTitle, message: LibraryCopy.resolve(language).emptyBody, action: (LibraryCopy.resolve(language).findVacation, { selectedTab = .discover; discoverPath = [.matchSetup] }))
                    .background(Palette.background.ignoresSafeArea())
                    .screenTitle(LibraryCopy.resolve(language).title)
            }
        }
    }
}
