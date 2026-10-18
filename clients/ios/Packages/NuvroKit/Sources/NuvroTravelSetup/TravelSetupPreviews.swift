import NuvroDesignSystem
import NuvroModels
import SwiftUI

private enum TravelSetupPreviewFixtures {
    static let airports = [
        Airport(code: "LOS", city: "Lagos", name: "Murtala Muhammed International", country: "Nigeria"),
        Airport(code: "JFK", city: "New York", name: "John F. Kennedy International", country: "United States"),
        Airport(code: "CDG", city: "Paris", name: "Charles de Gaulle", country: "France"),
        Airport(code: "LHR", city: "London", name: "Heathrow", country: "United Kingdom")
    ]

    static func profile(residence: String, passports: [String], airport: String, currency: String, accessibility: Set<AccessibilityPreference> = []) -> TravelProfile {
        TravelProfile(countryOfResidence: residence, passportCountries: passports, preferredDepartureAirport: airport, preferredCurrency: currency, paperworkPreference: .keepSimple, interests: [.beach, .food], pace: .relaxed, weather: [.mild, .warm], crowds: .quiet, flights: [.fewerStops], stays: [.quietNeighbourhood], group: .partner, accessibility: accessibility)
    }

    @MainActor
    static func passportStep() -> some View {
        var draft = TravelSetupDraft()
        draft.countryOfResidence = "NG"
        draft.stepIndex = TravelSetupStep.passports.rawValue
        let model = TravelSetupModel(store: InMemoryTravelSetupStore(draft: draft), airports: airports, locale: Locale(identifier: "en_US"))
        return DesignCanvas(fullBleed: false) { _ in TravelSetupView(model: model, onComplete: { _ in }) }
    }

    @MainActor
    static func review(_ profile: TravelProfile) -> some View {
        var draft = TravelSetupDraft(profile: profile)
        draft.stepIndex = TravelSetupStep.review.rawValue
        let model = TravelSetupModel(store: InMemoryTravelSetupStore(draft: draft), airports: airports, locale: Locale(identifier: "en_US"))
        return DesignCanvas(fullBleed: false) { _ in TravelSetupView(model: model, onComplete: { _ in }) }
    }
}

#Preview("Welcome") {
    DesignCanvas(fullBleed: false) { _ in
        TravelSetupView(model: TravelSetupModel(store: InMemoryTravelSetupStore(), airports: TravelSetupPreviewFixtures.airports), onComplete: { _ in })
    }
}

#Preview("Nigeria · NGN · LOS") {
    TravelSetupPreviewFixtures.review(TravelSetupPreviewFixtures.profile(residence: "NG", passports: ["NG"], airport: "LOS", currency: "NGN"))
}

#Preview("USA · USD · JFK") {
    TravelSetupPreviewFixtures.review(TravelSetupPreviewFixtures.profile(residence: "US", passports: ["US"], airport: "JFK", currency: "USD"))
}

#Preview("France · EUR · CDG") {
    TravelSetupPreviewFixtures.review(TravelSetupPreviewFixtures.profile(residence: "FR", passports: ["FR"], airport: "CDG", currency: "EUR"))
}

#Preview("UK · GBP · LHR") {
    TravelSetupPreviewFixtures.review(TravelSetupPreviewFixtures.profile(residence: "GB", passports: ["GB"], airport: "LHR", currency: "GBP"))
}

#Preview("Dual passport, lives in Canada") {
    TravelSetupPreviewFixtures.review(TravelSetupPreviewFixtures.profile(residence: "CA", passports: ["NG", "GB"], airport: "JFK", currency: "USD"))
}

#Preview("Accessibility preferences, large text") {
    TravelSetupPreviewFixtures.review(TravelSetupPreviewFixtures.profile(residence: "GH", passports: ["GH"], airport: "LOS", currency: "GHS", accessibility: [.wheelchair, .reducedWalking]))
        .dynamicTypeSize(.accessibility2)
}

#Preview("Passport step") {
    TravelSetupPreviewFixtures.passportStep()
}
