import Foundation
import NuvroLocalization
import NuvroModels
import Observation

@Observable
@MainActor
public final class TripsLibraryModel {
    public private(set) var entries: [TripLibraryEntry]
    public var overflow: TripLibraryEntry?
    public var pendingDelete: TripLibraryEntry?
    public private(set) var downloadedIDs: Set<String> = []
    public private(set) var lastEvent: String?

    public func clearEvent() {
        lastEvent = nil
    }
    public let copy: LibraryCopy

    public init(entries: [TripLibraryEntry], language: Language = .english) {
        self.entries = entries
        self.copy = .resolve(language)
    }

    public var isEmpty: Bool { entries.isEmpty }

    public func entries(in phase: TripLibraryEntry.Phase) -> [TripLibraryEntry] {
        entries.filter { $0.phase == phase }
    }

    public func isDownloaded(_ entry: TripLibraryEntry) -> Bool {
        downloadedIDs.contains(entry.id)
    }

    public func download(_ entry: TripLibraryEntry) {
        downloadedIDs.insert(entry.id)
        lastEvent = copy.downloaded
    }

    public func archive(_ entry: TripLibraryEntry) {
        entries.removeAll { $0.id == entry.id }
        lastEvent = copy.archived
    }

    public func requestDelete(_ entry: TripLibraryEntry) {
        overflow = nil
        pendingDelete = entry
    }

    public func confirmDelete() {
        guard let entry = pendingDelete else { return }
        entries.removeAll { $0.id == entry.id }
        pendingDelete = nil
        lastEvent = copy.deleted
    }

    public func cancelDelete() {
        pendingDelete = nil
    }
}
