import Foundation
import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import Testing

@testable import NuvroProfile

@Suite("Profile and region pickers")
@MainActor
struct PreferencePickerTests {
    private let profile = TravelerProfile(
        name: "Alex", initial: "A", homeLabel: "New York", styleChips: [],
        preferences: [PreferenceRow(name: "Interests", value: "Beach, Food", placeholder: "gold"), PreferenceRow(name: "Travel pace", value: "Relaxed", placeholder: "sky"), PreferenceRow(name: "Notes", value: "—", placeholder: "sky")],
        learned: [], usesTripHistory: true
    )
    private let choices: [String: PreferenceChoices] = [
        "Interests": PreferenceChoices(options: ["Beach", "Food", "Culture"].map { PickerOption($0) }, allowsMultiple: true),
        "Travel pace": PreferenceChoices(options: ["Relaxed", "Balanced", "Packed"].map { PickerOption($0) })
    ]

    @Test("rows without choices are not editable and values fall back to the profile")
    func editability() {
        let model = ProfileModel(profile: profile, choices: choices)
        #expect(model.canEdit(profile.preferences[0]))
        #expect(!model.canEdit(profile.preferences[2]))
        #expect(model.value(for: profile.preferences[1]) == "Relaxed")
        #expect(model.selection(for: profile.preferences[0]) == ["Beach", "Food"])
    }

    @Test("multi-select updates keep the option order and report the change once")
    func multiSelect() {
        let model = ProfileModel(profile: profile, choices: choices)
        model.update(profile.preferences[0], selection: ["Culture", "Beach"])
        #expect(model.value(for: profile.preferences[0]) == "Beach, Culture")
        #expect(model.lastUpdated == "Interests")
        model.update(profile.preferences[0], selection: [])
        #expect(model.value(for: profile.preferences[0]) == "Beach, Culture")
    }

    @Test("row values summarise long selections the way a settings row does")
    func displayValue() {
        let model = ProfileModel(profile: profile, choices: choices)
        model.update(profile.preferences[0], selection: ["Beach", "Food", "Culture"])
        #expect(model.displayValue(for: profile.preferences[0]) == "Beach, Food +1")
        #expect(model.displayValue(for: profile.preferences[1]) == "Relaxed")
        #expect(model.displayValue(for: profile.preferences[0], visible: 3) == "Beach, Food, Culture")
        let dotted = PreferenceRow(name: "Dotted", value: "Beach · Food · Culture", placeholder: "gold")
        #expect(model.selection(for: dotted) == ["Beach", "Food", "Culture"])
    }

    @Test("the learning card drops its illustration at accessibility text sizes")
    func illustrationVisibility() {
        #expect(LearningPreferenceCard.showsIllustration(for: .large))
        #expect(LearningPreferenceCard.showsIllustration(for: .xxxLarge))
        #expect(!LearningPreferenceCard.showsIllustration(for: .accessibility1))
        #expect(!LearningPreferenceCard.showsIllustration(for: .accessibility5))
    }

    @Test("single-select updates replace the value")
    func singleSelect() {
        let model = ProfileModel(profile: profile, choices: choices)
        model.update(profile.preferences[1], single: "Packed")
        #expect(model.value(for: profile.preferences[1]) == "Packed")
        #expect(model.lastUpdated == "Travel pace")
    }

    @Test("region selections start from the saved settings and expose labels")
    func region() {
        let settings = RegionSettings(appLanguage: "Español", spokenLanguages: "English, Portuguese", region: "United States", regionNote: "", currency: "USD", homeAirport: "New York · JFK")
        let airports = [Airport(code: "JFK", city: "New York", name: "JFK", country: "US"), Airport(code: "LIS", city: "Lisbon", name: "Lisbon", country: "PT")]
        let model = RegionModel(settings: settings, units: UnitPreferences(distance: .miles, temperature: .fahrenheit, clock: .twelveHour, date: .monthFirst), sampleDate: .now, choices: RegionChoices(regions: [], currencies: [], airports: airports))
        #expect(model.appLanguage == .spanish)
        #expect(model.spokenLanguages == ["English", "Portuguese"])
        #expect(model.homeAirportCode == "JFK")
        model.homeAirportCode = "LIS"
        #expect(model.homeAirportLabel == "Lisbon · LIS")
        #expect(model.languageOptions.count == Language.allCases.count)
    }
}
