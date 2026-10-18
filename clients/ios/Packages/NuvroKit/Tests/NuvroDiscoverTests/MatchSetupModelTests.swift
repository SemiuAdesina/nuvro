import Foundation
import NuvroModels
import Testing

@testable import NuvroDiscover

@Suite("MatchSetupModel")
@MainActor
struct MatchSetupModelTests {
    private let calendar = Calendar(identifier: .gregorian)
    private var september1: Date { calendar.date(from: DateComponents(year: 2026, month: 9, day: 1))! }
    private let airports = [
        Airport(code: "JFK", city: "New York", name: "John F. Kennedy International", country: "United States"),
        Airport(code: "LHR", city: "London", name: "Heathrow", country: "United Kingdom")
    ]

    private func model() -> MatchSetupModel {
        MatchSetupModel(airports: airports, departureCode: "JFK", locale: Locale(identifier: "en_US"), today: september1, calendar: calendar)
    }

    @Test("defaults mirror the prototype's starting state")
    func defaults() {
        let model = model()
        #expect(model.departure == "New York · JFK")

        #expect(model.budgetLabel == "$3,000")
        #expect(model.isBudgetFlexible)
        #expect(model.dateChoice == .flexible)
        #expect(model.travelersLabel == "2 adults")
        #expect(model.preferenceCountLabel == "3 selected")
        #expect(model.fineTuneSummary == "Warm · Quiet · Balanced · Anywhere")
        #expect(model.canSubmit)
    }

    @Test("the budget hint changes with the amount")
    func budgetHint() {
        let model = model()

        model.budgetMajor = 1200
        #expect(model.budgetHint == "Short trips and nearby escapes")
        model.budgetMajor = 3000
        #expect(model.budgetHint == "Most of Europe and the Caribbean fit")
        model.budgetMajor = 5000
        #expect(model.budgetHint == "Long-haul and premium stays open up")
        model.budgetMajor = 7900
        #expect(model.budgetHint == "Almost anywhere, in comfort")
    }

    @Test("the budget is money in minor units, never a float in the model output")
    func budgetMoney() {
        let model = model()
        model.budgetMajor = 2500

        #expect(model.budget?.amountMinor == 250000)
        #expect(model.budget?.currency == "USD")
    }

    @Test("travelers label reads naturally and never drops below one adult")
    func travelers() {
        let model = model()
        model.adults = 1
        #expect(model.travelersLabel == "1 adult")
        model.travelerCounts = ["adults": 0, "children": 2, "infants": 1]
        #expect(model.travelersLabel == "1 adult, 2 children, 1 infant")
        #expect(model.travelers.total == 4)
    }

    @Test("dates default to a flexible week next month and switch to an exact range")
    func dates() {
        let model = model()
        #expect(model.datesLabel == "Anytime in October · about 7 nights")
        model.dateChoice = .exact
        #expect(model.dateChoice == .exact)
        #expect(model.dateSelection.nights == 7)
        #expect(model.datesLabel.contains("Oct"))
        model.dateChoice = .flexible
        #expect(model.dateSelection.isFlexible)
    }

    @Test("a missing departure airport blocks submission")
    func departure() {
        let model = model()
        #expect(model.canSubmit)
        model.departureAirport = nil
        #expect(model.departure.isEmpty)
        #expect(!model.canSubmit)
        model.presentation = .airport
        #expect(model.presentation == .airport)
    }

    @Test("clearing every preference blocks submission")
    func preferences() {
        let model = model()
        for preference in TravelPreference.allCases where model.isSelected(preference) {
            model.toggle(preference)
        }

        #expect(model.preferenceCountLabel == "0 selected")
        #expect(!model.canSubmit)
    }
}
