import Foundation
import NuvroModels
import Testing

@testable import NuvroTravelSetup

@Suite("TravelSetupModel")
@MainActor
struct TravelSetupModelTests {
    private let airports = [Airport(code: "LOS", city: "Lagos", name: "Murtala Muhammed", country: "Nigeria"), Airport(code: "JFK", city: "New York", name: "JFK", country: "United States")]
    private let locale = Locale(identifier: "en_US")

    private func model(store: InMemoryTravelSetupStore = InMemoryTravelSetupStore()) -> TravelSetupModel {
        TravelSetupModel(store: store, airports: airports, locale: locale)
    }

    private func fill(_ model: TravelSetupModel) {
        model.update { $0.countryOfResidence = "NG" }
        model.addPassport("NG")
        model.update { $0.permitAnswer = .no }
        model.update { $0.departureAirport = "LOS" }
        model.update { $0.currency = "NGN" }
        model.update { $0.paperwork = .keepSimple }
        model.toggle(.beach, in: \.interests)
        model.update { $0.pace = .relaxed }
        model.toggle(.warm, in: \.weather)
        model.update { $0.crowds = .quiet }
        model.update { $0.group = .partner }
    }

    @Test("a new traveller starts at welcome and cannot continue past a question without an answer")
    func gating() {
        let setup = model()
        #expect(setup.step == .welcome)
        #expect(setup.canContinue)
        setup.advance()
        #expect(setup.step == .residence)
        #expect(!setup.canContinue)
        setup.update { $0.countryOfResidence = "NG" }
        #expect(setup.canContinue)
        setup.advance()
        #expect(setup.step == .passports)
        #expect(setup.progressIndex == 2)
    }

    @Test("residence and passport are separate and multiple passports are kept in order")
    func passports() {
        let setup = model()
        setup.update { $0.countryOfResidence = "NG" }
        #expect(setup.draft.passportCountries.isEmpty)
        setup.addPassport("GB")
        setup.addPassport("NG")
        setup.addPassport("GB")
        #expect(setup.draft.passportCountries == ["GB", "NG"])
        setup.removePassport("GB")
        #expect(setup.draft.passportCountries == ["NG"])
        setup.declinePassport()
        #expect(setup.draft.passportCountries.isEmpty)
        #expect(setup.draft.hasDeclinedPassport)
        setup.addPassport("NG")
        #expect(!setup.draft.hasDeclinedPassport)
    }

    @Test("optional steps can be skipped, required ones cannot, and back retraces")
    func skipAndBack() {
        let setup = model()
        setup.update { $0.countryOfResidence = "NG" }
        setup.addPassport("NG")
        setup.advance(); setup.advance(); setup.advance()
        #expect(setup.step == .permits)
        #expect(setup.canSkip)
        setup.skip()
        #expect(setup.step == .airport)
        #expect(setup.draft.permitAnswer == .skipped)
        #expect(!setup.canSkip)
        setup.skip()
        #expect(setup.step == .airport)
        setup.back()
        #expect(setup.step == .permits)
        #expect(!setup.isMovingForward)
    }

    @Test("currency is suggested from residence but never forced")
    func currency() {
        let setup = model()
        setup.update { $0.countryOfResidence = "NG" }
        #expect(setup.suggestedCurrency?.code == "NGN")
        #expect(setup.draft.currency == nil)
        setup.update { $0.currency = "USD" }
        #expect(setup.currency(setup.draft.currency)?.code == "USD")
        #expect(setup.airport("LOS")?.shortLabel == "Lagos · LOS")
    }

    @Test("exclusive preferences clear the others and the others clear the exclusive one")
    func exclusive() {
        let setup = model()
        setup.selectExclusive(.warm, exclusive: .noPreference, in: \.weather)
        setup.selectExclusive(.mild, exclusive: .noPreference, in: \.weather)
        #expect(setup.draft.weather == [.warm, .mild])
        setup.selectExclusive(.noPreference, exclusive: .noPreference, in: \.weather)
        #expect(setup.draft.weather == [.noPreference])
        setup.selectExclusive(.cool, exclusive: .noPreference, in: \.weather)
        #expect(setup.draft.weather == [.cool])
    }

    @Test("progress is persisted and restored after relaunch")
    func restoration() {
        let store = InMemoryTravelSetupStore()
        let first = model(store: store)
        first.advance()
        first.update { $0.countryOfResidence = "FR" }
        first.advance()
        first.addPassport("FR")
        first.advance()
        let second = model(store: store)
        #expect(second.step == .permits)
        #expect(second.draft.countryOfResidence == "FR")
        #expect(second.draft.passportCountries == ["FR"])
    }

    @Test("review lets a line be edited and completion persists the profile and clears the draft")
    func completion() {
        let store = InMemoryTravelSetupStore()
        let setup = model(store: store)
        fill(setup)
        setup.edit(.review)
        #expect(setup.step == .review)
        setup.edit(.currency)
        #expect(setup.step == .currency)
        setup.update { $0.currency = "USD" }
        setup.returnToReview()
        setup.advance()
        #expect(setup.completedProfile?.preferredCurrency == "USD")
        #expect(setup.completedProfile?.countryOfResidence == "NG")
        #expect(setup.completedProfile?.passportCountries == ["NG"])
        #expect(store.loadProfile() == setup.completedProfile)
        #expect(store.loadDraft() == nil)
    }

    @Test("completion is blocked without passport context unless the traveller explicitly declined")
    func passportPolicy() {
        var draft = TravelSetupDraft(profile: TravelProfile(countryOfResidence: "NG", passportCountries: [], preferredDepartureAirport: nil, preferredCurrency: "NGN", paperworkPreference: .keepSimple, interests: [.beach], pace: .relaxed, weather: [.warm], crowds: .quiet, flights: [], stays: [], group: .solo, accessibility: []))
        #expect(draft.profile() == nil)
        draft.hasDeclinedPassport = true
        #expect(draft.profile()?.hasPassportContext == false)
    }

    @Test("edit mode starts on the requested step, hides progress, and saves without a draft")
    func editMode() {
        let profile = TravelProfile(countryOfResidence: "GB", passportCountries: ["GB"], preferredDepartureAirport: "JFK", preferredCurrency: "GBP", paperworkPreference: .evisaFine, interests: [.food], pace: .balanced, weather: [.mild], crowds: .balanced, flights: [], stays: [], group: .friends, accessibility: [])
        let store = InMemoryTravelSetupStore(profile: profile)
        let setup = TravelSetupModel(store: store, airports: airports, editing: .airport, locale: locale)
        #expect(setup.step == .airport)
        #expect(!setup.showsProgress)
        #expect(!setup.canGoBack)
        setup.update { $0.departureAirport = "LOS" }
        setup.advance()
        #expect(store.loadProfile()?.preferredDepartureAirport == "LOS")
        #expect(store.loadProfile()?.preferredCurrency == "GBP")
        #expect(store.loadDraft() == nil)
    }

    @Test("change on a passport or permit swaps it for the next pick and dismissing keeps it")
    func replace() {
        let model = self.model()
        model.addPassport("NG")
        model.replacePassport("NG")
        #expect(model.picker == .passport)
        model.addPassport("US")
        #expect(model.draft.passportCountries == ["US"])
        model.addPermit("GB")
        model.replacePermit("GB")
        model.picker = nil
        model.addPermit("FR")
        #expect(model.draft.residencePermits == ["GB", "FR"])
    }
}

@Suite("TravelCatalog")
struct TravelCatalogTests {
    @Test("regions come from Foundation with flags and search by name or code")
    func regions() {
        let regions = TravelCatalog.regions(locale: Locale(identifier: "en_US"))
        let nigeria = regions.first { $0.code == "NG" }
        #expect(nigeria?.name == "Nigeria")
        #expect(nigeria?.flag == "🇳🇬")
        #expect(regions.count > 200)
        #expect(TravelCatalog.matches(nigeria!, "nig"))
        #expect(TravelCatalog.matches(nigeria!, "NG"))
        #expect(!TravelCatalog.matches(nigeria!, "France"))
        let dominica = regions.first { $0.code == "DM" }!
        #expect(!TravelCatalog.matches(dominica, "ni"))
        let southAfrica = regions.first { $0.code == "ZA" }!
        #expect(TravelCatalog.matches(southAfrica, "afr"))
        #expect(TravelCatalog.matches(regions.first { $0.code == "CI" }!, "cote"))
    }

    @Test("currencies are the ISO list, not a hand-picked few")
    func currencies() {
        let currencies = TravelCatalog.currencies(locale: Locale(identifier: "en_US"))
        #expect(currencies.contains { $0.code == "NGN" && $0.name == "Nigerian Naira" })
        #expect(currencies.contains { $0.code == "KES" })
        #expect(currencies.count > 100)
        #expect(TravelCatalog.suggestedCurrency(forRegion: "FR") == "EUR")
        #expect(TravelCatalog.suggestedCurrency(forRegion: nil) == nil)
    }

    @Test("the summary reads naturally")
    func summary() {
        let summary = TravelProfileSummary(copy: .english, locale: Locale(identifier: "en_US"))
        #expect(summary.weatherLabel([.mild, .warm]) == "Mild to warm")
        var draft = TravelSetupDraft()
        draft.countryOfResidence = "NG"; draft.passportCountries = ["NG"]; draft.currency = "NGN"; draft.paperwork = .keepSimple
        let lines = summary.lines(for: draft, airports: [])
        #expect(lines.first?.value == "Nigeria")
        #expect(lines.contains { $0.key == "Currency" && $0.value.hasPrefix("NGN") })
        #expect(lines.contains { $0.key == "Usually leaves from" && $0.value == "Not added yet" })
    }

}
