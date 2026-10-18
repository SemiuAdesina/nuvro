import NuvroAPI
import NuvroAuthentication
import NuvroConnectivity
import NuvroDesignSystem
import NuvroDiscover
import NuvroLocalization
import NuvroOnboarding
import NuvroModels
import NuvroPermissions
import NuvroPlan
import NuvroTravel
import NuvroTravelSetup
import SwiftUI

struct RootView: View {
    @State private var onboarding = RootView.makeOnboardingModel()
    @State private var authentication: AuthenticationCoordinator
    private let account: (any AccountGateway)?
    private let discovery: (any DiscoverGateway)?
    private let developmentCredential: String?
    @State private var connectivity = ConnectivityMonitor()
    @State private var permissions = PermissionCenter(authority: RootView.permissionAuthority())
    @State private var feedback = TransientFeedback()
    @AppStorage("hasFinishedOnboarding") private var hasFinishedOnboarding = false
    @State private var selectedTab: AppTab = .discover
    @State private var discoverPath: [AppRoute] = []
    @State private var tripPlan = TripPlanModel(plan: PrototypeContent.madeiraPlan, locale: PrototypeContent.locale)
    @State private var travelMode = RootView.makeTravelModeModel()
    @State private var tripsPath: [TripRoute] = []
    @AppStorage("hasStartedTrip") private var hasStartedTrip = false
    @AppStorage(ThemePreference.storageKey) private var theme: ThemePreference = .system
    @AppStorage("hasCompletedTravelSetup") private var hasCompletedTravelSetup = false
    @State private var travelSetup = TravelSetupModel(store: UserDefaultsTravelSetupStore(), airports: PrototypeContent.airports, locale: PrototypeContent.locale)
    @State private var travelProfile = UserDefaultsTravelSetupStore().loadProfile()
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    init() {
        let composition = BackendComposition.resolve()
        _authentication = State(initialValue: composition.authentication)
        account = composition.account
        discovery = composition.discovery
        developmentCredential = composition.developmentCredential
    }

    private var phase: LaunchPhase {
        LaunchPhase(authentication: authentication.state, hasCompletedTravelSetup: hasCompletedTravelSetup, hasFinishedOnboarding: hasFinishedOnboarding)
    }

    var body: some View {
        ZStack {
            switch phase {
            case .onboarding:
                OnboardingView(model: onboarding).transition(.opacity)
            case .restoringSession:
                DesignCanvas { _ in AuthenticationBootstrapView(language: onboarding.language) }.transition(.opacity)
            case .authentication:
                DesignCanvas { _ in AuthenticationView(coordinator: authentication, language: onboarding.language) }.transition(.opacity)
            case .travelSetup:
                DesignCanvas(fullBleed: false) { _ in
                    TravelSetupView(model: travelSetup) { profile in
                        travelProfile = profile
                        hasCompletedTravelSetup = true
                    }
                }
                .transition(.opacity)
            case .main:
                DesignCanvas(fullBleed: false) { _ in
                    MainTabs(
                        selectedTab: $selectedTab,
                        discoverPath: $discoverPath,
                        tripsPath: $tripsPath,
                        tripPlan: tripPlan,
                        travelMode: travelMode,
                        hasStartedTrip: $hasStartedTrip,
                        language: onboarding.language,
                        travelProfile: $travelProfile,
                        account: account,
                        discovery: discovery
                    )
                }
                .transition(.opacity)
            }
        }
        .overlay(alignment: .top) {
            if connectivity.isOffline, phase == .main {
                let copy = StateCopy.resolve(onboarding.language)
                NuvroOfflineBanner(title: connectivity.status == .reconnecting ? copy.reconnecting : copy.offlineTitle, detail: connectivity.status == .reconnecting ? nil : copy.offlineCachedBody, isReconnecting: connectivity.status == .reconnecting)
                    .padding(.horizontal, 20)
                    .transition(.move(edge: .top).combined(with: .opacity))
            }
        }
        .transientFeedback(feedback)
        .environment(authentication)
        .environment(connectivity)
        .environment(permissions)
        .environment(feedback)
        .animation(reduceMotion ? nil : Motion.settle(0.4), value: connectivity.status)
        .animation(reduceMotion ? nil : Motion.fade(0.5), value: phase)
        .preferredColorScheme(phase == .main || phase == .travelSetup ? theme.colorScheme : .dark)
        .task(id: onboarding.exit) { await handleOnboardingExit() }
        .onChange(of: phase) { _, next in
            if next == .onboarding, hasCompletedTravelSetup { onboarding = OnboardingFlowModel(language: onboarding.language, startingStage: .discover) }
        }
        .task { await restoreSession() }
        .task { connectivity.start() }
    }

    private func restoreSession() async {
        await authentication.restoreSession()
        guard let developmentCredential, authentication.state.requiresAuthentication else { return }
        await authentication.signIn(credential: developmentCredential)
    }

    private func handleOnboardingExit() async {
        guard let exit = onboarding.exit else { return }
        try? await Task.sleep(for: .seconds(0.35))
        guard !Task.isCancelled else { return }
        if exit == .existingTrip {
            hasStartedTrip = true
            selectedTab = .trips
        }
        hasFinishedOnboarding = true
    }

    private static func permissionAuthority() -> any PermissionAuthority {
        guard let index = CommandLine.arguments.firstIndex(of: "-uiTestPermissions"), CommandLine.arguments.indices.contains(index + 1) else {
            return SystemPermissionAuthority()
        }
        let state: PermissionState = CommandLine.arguments[index + 1] == "denied" ? .denied : .granted
        return ScriptedPermissionAuthority(states: Dictionary(uniqueKeysWithValues: PermissionKind.allCases.map { ($0, state) }))
    }

    private static func makeTravelModeModel() -> TravelModeModel {
        let day = PrototypeContent.madeiraPlan.day(3) ?? PrototypeItinerary.days[2]
        var components = DateComponents()
        components.year = 2026; components.month = 10; components.day = 14; components.hour = 9; components.minute = 26
        let now = Calendar.current.date(from: components) ?? .now
        return TravelModeModel(
            day: day, dayNumber: 3, destinationName: PrototypeContent.madeira.name, now: now,
            weather: WeatherSummary(condition: "Light clouds", temperatureCelsius: 23),
            pendingAdjustment: PrototypeItinerary.rainAdjustment
        )
    }

    private static func makeOnboardingModel() -> OnboardingFlowModel {
        OnboardingFlowModel(language: Language.detect(from: Locale.preferredLanguages))
    }
}
