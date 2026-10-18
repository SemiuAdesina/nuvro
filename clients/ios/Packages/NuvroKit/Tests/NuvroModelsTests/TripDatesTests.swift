import Foundation
import Testing

@testable import NuvroModels

@Suite("Trip dates, travellers and airports")
struct TripDatesTests {
    private let calendar = Calendar(identifier: .gregorian)
    private var october12: Date { calendar.date(from: DateComponents(year: 2026, month: 10, day: 12))! }
    private var october19: Date { calendar.date(from: DateComponents(year: 2026, month: 10, day: 19))! }

    @Test("exact dates count nights and normalise an inverted range")
    func exact() {
        let selection = DateSelection.exact(start: october12, end: october19)
        #expect(selection.nights == 7)
        #expect(!selection.isFlexible)
        let inverted = DateSelection.exact(start: october19, end: october12).normalised()
        #expect(inverted.nights == 1)
    }

    @Test("flexible dates describe the month and nights")
    func flexible() {
        let selection = DateSelection.flexible(month: october12, nights: 6)
        #expect(selection.isFlexible)
        #expect(selection.label(locale: Locale(identifier: "en_US"), flexibleWord: "Anytime in", aboutWord: "about", nightsWord: "nights") == "Anytime in October · about 6 nights")
    }

    @Test("traveller counts never go below one adult and label naturally")
    func travellers() {
        let counts = TravelerCounts(adults: 0, children: 1, infants: -2)
        #expect(counts.adults == 1)
        #expect(counts.infants == 0)
        #expect(counts.total == 2)
        let label = TravelerCounts(adults: 2, children: 1).label(adultWord: "adult", adultsWord: "adults", childWord: "child", childrenWord: "children", infantWord: "infant", infantsWord: "infants")
        #expect(label == "2 adults, 1 child")
    }

    @Test("airport search matches code, city, name and country case-insensitively")
    func airports() {
        let jfk = Airport(code: "JFK", city: "New York", name: "John F. Kennedy International", country: "United States")
        #expect(jfk.matches("jfk"))
        #expect(jfk.matches("new yo"))
        #expect(jfk.matches("kennedy"))
        #expect(jfk.matches(""))
        #expect(!jfk.matches("lisbon"))
        #expect(jfk.shortLabel == "New York · JFK")
    }
}
