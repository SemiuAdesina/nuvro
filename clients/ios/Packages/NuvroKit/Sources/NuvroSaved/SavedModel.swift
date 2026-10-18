import Foundation
import NuvroLocalization
import NuvroModels
import Observation

@Observable
@MainActor
public final class SavedModel {
    public let destinations: [SavedDestination]
    public let places: [SavedPlace]
    public let comparison: DestinationComparison
    public var selectedTab: String
    public private(set) var selectedIDs: Set<String>
    public var chosenColumn: String?
    public let copy: SavedCopy
    public let compareCopy: CompareCopy

    public init(destinations: [SavedDestination], places: [SavedPlace], comparison: DestinationComparison, preselected: Set<String> = [], language: Language = .english) {
        self.destinations = destinations
        self.places = places
        self.comparison = comparison
        self.selectedIDs = preselected
        self.copy = .resolve(language)
        self.compareCopy = .resolve(language)
        self.selectedTab = SavedCopy.resolve(language).tabs.first ?? ""
    }

    public var isDestinationsTab: Bool { selectedTab == copy.tabs.first }
    public var selectedCount: Int { selectedIDs.count }
    public var canCompare: Bool { selectedCount > 1 }
    public var compareLabel: String { canCompare ? "\(copy.compare) \(selectedCount)" : copy.compare }
    public var selectionHint: String { canCompare ? "\(selectedCount) \(copy.selectedHint)" : copy.tapTwo }

    public func isSelected(_ destination: SavedDestination) -> Bool { selectedIDs.contains(destination.id) }

    public func toggle(_ destination: SavedDestination) {
        if selectedIDs.contains(destination.id) { selectedIDs.remove(destination.id) } else { selectedIDs.insert(destination.id) }
    }

    public func choose(_ column: ComparisonColumn) { chosenColumn = chosenColumn == column.id ? nil : column.id }

    public func columnLabel(_ column: ComparisonColumn) -> String {
        chosenColumn == column.id ? "\(compareCopy.chosen) · \(column.name)" : "\(compareCopy.choose) \(column.name)"
    }
}
