import Testing

@testable import NuvroModels
@testable import NuvroSaved

@Suite("SavedModel")
@MainActor
struct SavedModelTests {
    private func media(_ name: String) -> DestinationMedia { DestinationMedia(placeholder: "gold", altText: name) }

    private func model() -> SavedModel {
        SavedModel(
            destinations: [
                SavedDestination(id: "madeira", name: "Madeira", country: "Portugal", matchPercent: 94, reason: "", note: "", media: media("m")),
                SavedDestination(id: "lisbon", name: "Lisbon", country: "Portugal", matchPercent: 91, reason: "", note: "", media: media("l")),
                SavedDestination(id: "kyoto", name: "Kyoto", country: "Japan", matchPercent: 91, reason: "", note: "", media: media("k"))
            ],
            places: [],
            comparison: DestinationComparison(subtitle: "", columns: [
                ComparisonColumn(name: "Madeira", matchPercent: 94, gist: "", media: media("m")),
                ComparisonColumn(name: "Lisbon", matchPercent: 91, gist: "", media: media("l"))
            ], metrics: [], differences: [])
        )
    }

    @Test("compare unlocks at two selected destinations and the labels follow the count")
    func selection() {
        let saved = model()
        #expect(!saved.canCompare)
        #expect(saved.compareLabel == "Compare")
        #expect(saved.selectionHint == "Tap two destinations to compare them")

        saved.toggle(saved.destinations[0])
        #expect(!saved.canCompare)
        saved.toggle(saved.destinations[1])
        #expect(saved.canCompare)
        #expect(saved.compareLabel == "Compare 2")
        #expect(saved.selectionHint == "2 selected — tap Compare")

        saved.toggle(saved.destinations[1])
        #expect(!saved.canCompare)
    }

    @Test("choosing a column is reversible and changes its label")
    func choose() {
        let saved = model()
        let column = saved.comparison.columns[0]
        #expect(saved.columnLabel(column) == "Choose Madeira")
        saved.choose(column)
        #expect(saved.columnLabel(column) == "Chosen · Madeira")
        saved.choose(column)
        #expect(saved.chosenColumn == nil)
    }

    @Test("the destinations tab is the first tab")
    func tabs() {
        let saved = model()
        #expect(saved.isDestinationsTab)
        saved.selectedTab = "Places"
        #expect(!saved.isDestinationsTab)
    }
}
