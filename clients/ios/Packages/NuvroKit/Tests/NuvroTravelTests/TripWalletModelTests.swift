import NuvroModels
import Testing

@testable import NuvroTravel

@Suite("TripWalletModel downloads")
@MainActor
struct TripWalletModelTests {
    private let documents = [
        TripDocument(id: "flight", category: .flight, title: "JFK → FNC", subtitle: "", whenLabel: "", travelers: 2, isAvailableOffline: true, reference: "TP1691-8842"),
        TripDocument(id: "tour", category: .activity, title: "West Island Tour", subtitle: "", whenLabel: "", travelers: 2, isAvailableOffline: false, reference: "WIT-2038")
    ]

    @Test("offline state starts from the contract and can be added or removed")
    func offlineState() {
        let model = TripWalletModel(documents: documents)
        #expect(model.isAvailableOffline(documents[0]))
        #expect(!model.isAvailableOffline(documents[1]))
        #expect(!model.allDownloaded)
        model.download(documents[1])
        #expect(model.allDownloaded)
        #expect(model.lastEvent == "Downloaded for offline")
        model.requestRemoveDownload(documents[0])
        #expect(model.pendingDownloadRemoval?.id == "flight")
        model.confirmRemoveDownload()
        #expect(!model.isAvailableOffline(documents[0]))
        #expect(model.lastEvent == "Offline copy removed")
    }

    @Test("download all covers every document and is idempotent")
    func downloadAll() {
        let model = TripWalletModel(documents: documents)
        model.downloadAll()
        #expect(model.allDownloaded)
        model.downloadAll()
        #expect(model.allDownloaded)
    }

    @Test("references stay masked until revealed")
    func masking() {
        let model = TripWalletModel(documents: documents)
        #expect(model.reference(for: documents[0]) == "••••••-••••")
        model.toggleReveal(documents[0])
        #expect(model.reference(for: documents[0]) == "TP1691-8842")
    }
}
