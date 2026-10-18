import Foundation
import NuvroLocalization
import NuvroModels
import Observation

@Observable
@MainActor
public final class TripWalletModel {
    public let documents: [TripDocument]
    public private(set) var revealedIDs: Set<String> = []
    public private(set) var downloadedIDs: Set<String>
    public var openTicket: TripDocument?
    public var pendingDownloadRemoval: TripDocument?
    public private(set) var lastEvent: String?

    public func clearEvent() {
        lastEvent = nil
    }
    public let copy: WalletCopy

    public init(documents: [TripDocument], language: Language = .english) {
        self.documents = documents
        self.downloadedIDs = Set(documents.filter(\.isAvailableOffline).map(\.id))
        self.copy = .resolve(language)
    }

    public func isAvailableOffline(_ document: TripDocument) -> Bool {
        downloadedIDs.contains(document.id)
    }

    public var allDownloaded: Bool {
        documents.allSatisfy(isAvailableOffline)
    }

    public func download(_ document: TripDocument) {
        downloadedIDs.insert(document.id)
        lastEvent = copy.downloaded
    }

    public func downloadAll() {
        guard !allDownloaded else { return }
        downloadedIDs = Set(documents.map(\.id))
        lastEvent = copy.downloaded
    }

    public func addToWallet(_ document: TripDocument) {
        guard document.canAddToWallet else { return }
        lastEvent = copy.walletUnavailable
    }

    public func addAllToWallet() {
        guard documents.contains(where: \.canAddToWallet) else { return }
        lastEvent = copy.walletUnavailable
    }

    public func requestRemoveDownload(_ document: TripDocument) {
        guard isAvailableOffline(document) else { return }
        pendingDownloadRemoval = document
    }

    public func confirmRemoveDownload() {
        guard let document = pendingDownloadRemoval else { return }
        downloadedIDs.remove(document.id)
        pendingDownloadRemoval = nil
        lastEvent = copy.downloadRemoved
    }

    public func cancelRemoveDownload() {
        pendingDownloadRemoval = nil
    }

    public var neededNext: TripDocument? {
        documents.first(where: \.isNeededNext)
    }

    public var groups: [(category: TripDocument.Category, title: String, documents: [TripDocument])] {
        TripDocument.Category.allCases.compactMap { category in
            let items = documents.filter { $0.category == category }
            return items.isEmpty ? nil : (category, title(for: category), items)
        }
    }

    public func reference(for document: TripDocument) -> String {
        revealedIDs.contains(document.id) ? document.reference : document.maskedReference
    }

    public func toggleReveal(_ document: TripDocument) {
        if revealedIDs.contains(document.id) { revealedIDs.remove(document.id) } else { revealedIDs.insert(document.id) }
    }

    public func isRevealed(_ document: TripDocument) -> Bool {
        revealedIDs.contains(document.id)
    }

    public func title(for category: TripDocument.Category) -> String {
        switch category {
        case .flight: copy.flights
        case .stay: copy.stay
        case .activity: copy.activities
        case .transfer: copy.transfers
        case .other: copy.other
        }
    }

    public func symbol(for category: TripDocument.Category) -> String {
        switch category {
        case .flight: "airplane"
        case .stay: "bed.double.fill"
        case .activity: "ticket.fill"
        case .transfer: "car.fill"
        case .other: "doc.text.fill"
        }
    }
}
