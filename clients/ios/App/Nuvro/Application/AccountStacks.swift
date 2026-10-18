import NuvroAPI
import NuvroAuthentication
import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import NuvroProfile
import NuvroSaved
import NuvroTravelSetup
import SwiftUI

struct SavedStack: View {
    let onExplore: () -> Void
    @State private var path: [SavedRoute] = []
    @State private var model = SavedModel(destinations: PrototypeLibrary.savedDestinations, places: PrototypeLibrary.savedPlaces, comparison: PrototypeLibrary.comparison, preselected: ["madeira", "lisbon"])

    var body: some View {
        NavigationStack(path: $path) {
            SavedView(model: model, onCompare: { path.append(.compare) }, onExplore: onExplore)
                .navigationDestination(for: SavedRoute.self) { route in
                    switch route {
                    case .compare: CompareView(model: model)
                    }
                }
        }
    }
}

struct ProfileStack: View {
    @Environment(AuthenticationCoordinator.self) private var authentication
    @Binding var travelProfile: TravelProfile?
    let account: (any AccountGateway)?
    @State private var path: [ProfileRoute] = []
    @State private var editingStep: TravelSetupStep?
    @State private var profile = ProfileModel(profile: PrototypeLibrary.profile, choices: PrototypeLibrary.preferenceChoices)
    @State private var region: RegionModel
    @State private var settings = SettingsModel(groups: PrototypeLibrary.settingsGroups, notifications: PrototypeLibrary.notifications, permissions: PrototypeLibrary.permissions)

    init(travelProfile: Binding<TravelProfile?>, account: (any AccountGateway)? = nil) {
        self._travelProfile = travelProfile
        self.account = account
        let base = PrototypeLibrary.region
        let travel = travelProfile.wrappedValue
        let airport = PrototypeContent.airports.first { $0.code == travel?.preferredDepartureAirport }
        let settings = RegionSettings(
            appLanguage: base.appLanguage, spokenLanguages: base.spokenLanguages,
            region: travel.map { TravelCatalog.regionName($0.countryOfResidence, locale: PrototypeContent.locale) } ?? base.region, regionNote: base.regionNote,
            currency: travel?.preferredCurrency ?? base.currency, homeAirport: airport?.shortLabel ?? base.homeAirport
        )
        self._region = State(initialValue: RegionModel(settings: settings, units: UnitPreferences(distance: .miles, temperature: .fahrenheit, clock: .twelveHour, date: .monthFirst), sampleDate: .now, choices: PrototypeLibrary.regionChoices))
    }

    var body: some View {
        NavigationStack(path: $path) {
            ProfileView(model: profile, travelProfile: travelProfile, onRegion: { path.append(.region) }, onSettings: { path.append(.settings) }, onEditTravelAccess: { editingStep = $0 })
                .navigationDestination(for: ProfileRoute.self) { route in
                    switch route {
                    case .region: RegionView(model: region)
                    case .settings: SettingsView(model: settings)
                    }
                }
        }
        .sheet(item: $editingStep) { step in
            TravelSetupEditSheet(step: step, airports: PrototypeContent.airports) { travelProfile = $0 }
        }
        .task(id: travelProfile) {
            if let account {
                await profile.load(from: account, name: authentication.state.session?.displayName)
            } else if let travelProfile {
                profile.apply(TravelerProfile.make(from: travelProfile, name: authentication.state.session?.displayName, airports: PrototypeContent.airports, locale: PrototypeContent.locale))
            }
        }
        .onChange(of: profile.usesTripHistory) { _, enabled in
            guard let account else { return }
            Task { await profile.syncHistoryLearning(enabled, with: account) }
        }
        .task {
            settings.actions = AccountActions(
                signOut: { await authentication.signOut() },
                deleteAccount: { try await authentication.deleteAccount() }
            )
        }
    }
}

struct TravelSetupEditSheet: View {
    @Environment(\.dismiss) private var dismiss
    let step: TravelSetupStep
    let airports: [Airport]
    let onSaved: (TravelProfile) -> Void
    @State private var model: TravelSetupModel

    init(step: TravelSetupStep, airports: [Airport], onSaved: @escaping (TravelProfile) -> Void) {
        self.step = step
        self.airports = airports
        self.onSaved = onSaved
        self._model = State(initialValue: TravelSetupModel(store: UserDefaultsTravelSetupStore(), airports: airports, editing: step, locale: PrototypeContent.locale))
    }

    var body: some View {
        DesignCanvas(fullBleed: false) { _ in
            TravelSetupView(model: model, onComplete: { onSaved($0); dismiss() }, onCancel: { dismiss() })
        }
        .presentationDetents([.large])
    }
}
