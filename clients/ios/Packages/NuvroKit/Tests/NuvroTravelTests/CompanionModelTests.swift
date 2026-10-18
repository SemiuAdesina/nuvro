import Foundation
import Testing

@testable import NuvroModels
@testable import NuvroTravel

@Suite("Companion models")
@MainActor
struct CompanionModelTests {
    private let pair = LanguagePair(source: "English", target: "Português")
    private let phrases = [
        TravelPhrase(context: "Hotel", systemImage: "bed.double.fill", original: "I have a reservation.", translated: "Tenho uma reserva."),
        TravelPhrase(context: "Emergency", systemImage: "cross.case.fill", original: "I need help.", translated: "Preciso de ajuda.")
    ]

    private func language() -> LanguageCompanionModel {
        LanguageCompanionModel(pair: pair, sample: PhraseTranslation(pair: pair, original: "We have a reservation for two.", translated: "Temos uma reserva para duas pessoas."), phrases: phrases)
    }

    @Test("listening produces the sample translation and swapping reverses it")
    func listenAndSwap() {
        let model = language()
        model.beginListening()
        #expect(model.isListening)
        model.finishListening()

        #expect(!model.isListening)
        #expect(model.result?.translated == "Temos uma reserva para duas pessoas.")
        model.swapLanguages()
        #expect(model.pair.source == "Português")
        #expect(model.result?.original == "Temos uma reserva para duas pessoas.")
    }

    @Test("a phrase opens the show-to-local card and typed text matches known phrases")
    func phrasesAndTyping() {
        let model = language()
        model.show(phrases[1])
        #expect(model.isShowingLarge)
        #expect(model.result?.translated == "Preciso de ajuda.")

        model.typedText = "i have a reservation."
        model.translateTyped()
        #expect(model.result?.translated == "Tenho uma reserva.")
        model.typedText = "   "
        model.translateTyped()
        #expect(model.result?.translated == "Tenho uma reserva.")
    }

    private func media(_ name: String) -> DestinationMedia { DestinationMedia(placeholder: "gold", altText: name) }

    private func nearby() -> ExploreNearbyModel {
        ExploreNearbyModel(places: [
            NearbyPlace(id: "m", name: "Mercado", category: .food, categoryLabel: "Market", walkMinutes: 8, travelLabel: "8 min walk", isOpen: true, hoursLabel: "Open now", reason: "", media: media("m"), coordinate: GeoCoordinate(latitude: 0, longitude: 0), markerLetter: "M"),
            NearbyPlace(id: "v", name: "Viewpoint", category: .views, categoryLabel: "Viewpoint", walkMinutes: 45, travelLabel: "45 min by car", isOpen: true, hoursLabel: "Always open", reason: "", media: media("v"), coordinate: GeoCoordinate(latitude: 0, longitude: 0), markerLetter: "V")
        ], freeTime: FreeTimeWindow(minutes: 140, untilLabel: "dinner reservation"), locationName: "Funchal", destinationName: "Madeira")
    }

    @Test("filters, selection, free time and saved places behave")
    func nearbyBehaviour() {
        let model = nearby()
        #expect(model.freeTimeLabel == "2h 20m")
        #expect(model.fits(model.places[0]))
        #expect(!model.fits(model.places[1]))
        #expect(model.selected?.id == "m")

        model.toggleFilter(.views)
        #expect(model.visiblePlaces.map(\.id) == ["v"])
        #expect(model.selected?.id == "v")
        model.toggleFilter(.views)
        #expect(model.visiblePlaces.count == 2)

        model.toggleSaved(model.places[1])
        model.toggleFilter(.saved)
        #expect(model.visiblePlaces.map(\.id) == ["v"])
        model.addToToday(model.places[1])
        #expect(model.isAdded(model.places[1]))
    }

    @Test("the wallet masks references until revealed and groups by category")
    func wallet() {
        let documents = [
            TripDocument(id: "c", category: .activity, title: "Cable Car", subtitle: "", whenLabel: "", travelers: 2, isAvailableOffline: true, reference: "TCF48213", isNeededNext: true),
            TripDocument(id: "f", category: .flight, title: "JFK → FNC", subtitle: "", whenLabel: "", travelers: 2, isAvailableOffline: true, reference: "PNR7Q2LX")
        ]
        let model = TripWalletModel(documents: documents)

        #expect(model.neededNext?.id == "c")
        #expect(model.groups.map(\.category) == [.flight, .activity])
        #expect(model.reference(for: documents[1]) == "••••••••")
        model.toggleReveal(documents[1])
        #expect(model.reference(for: documents[1]) == "PNR7Q2LX")
        model.addToWallet(documents[0])
        #expect(model.lastEvent == nil)
        model.addAllToWallet()
        #expect(model.lastEvent == nil)
        let passable = TripWalletModel(documents: [TripDocument(id: "p", category: .activity, title: "", subtitle: "", whenLabel: "", travelers: 1, isAvailableOffline: true, reference: "R", canAddToWallet: true)])
        passable.addToWallet(passable.documents[0])
        #expect(passable.lastEvent == passable.copy.walletUnavailable)
    }

    @Test("copying a translation reports it and voices match language names")
    func copyAndVoice() {
        let model = language()
        #expect(model.copyTranslation() == nil)
        model.showSample()
        #expect(model.copyTranslation() == "Temos uma reserva para duas pessoas.")
        #expect(model.lastEvent == model.copy.copied)
        #expect(PhraseSpeaker.language(named: "Portuguese", among: ["en-US", "pt-BR", "pt-PT"]) == "pt-BR")
        #expect(PhraseSpeaker.language(named: "Klingon", among: ["en-US"]) == nil)
    }

    @Test("alerts group by priority and dismissals remove them")
    func alerts() {
        let alerts = [
            TripAlert(id: "r", priority: .now, kind: .weather, severity: .actionNeeded, title: "Rain", body: "", reason: "", actionLabel: "Adjust", destination: .replanning),
            TripAlert(id: "f", priority: .today, kind: .flight, severity: .reminder, title: "Flight", body: "", reason: "", actionLabel: "View", destination: .document)
        ]
        let model = TripAlertsModel(alerts: alerts)

        #expect(model.groups.map(\.title) == ["Now", "Today"])
        #expect(model.activeCount == 2)
        model.dismiss(alerts[0])
        #expect(model.groups.map(\.title) == ["Today"])
        model.dismiss(alerts[1])
        #expect(model.isAllClear)
    }

    @Test("the summary computes the budget difference and keeps feedback editable")
    func summary() throws {
        let summary = TripSummary(
            destinationName: "Madeira", dateRangeLabel: "Oct 12–19", days: 7, placesCount: 12, activitiesCount: 5, kilometresExplored: 42, hero: media("hero"), places: [],
            total: try Money(amountMinor: 273000, currency: "USD"), plannedTotal: try Money(amountMinor: 278000, currency: "USD"),
            lovedOptions: ["Coastal places"], notForMeOptions: ["Long transfers"],
            suggestion: NextTripSuggestion(name: "Azores", country: "Portugal", reason: "", media: media("a"))
        )
        let model = TripSummaryModel(summary: summary, loved: ["Coastal places"], notForMe: [], locale: Locale(identifier: "en_US"))

        #expect(model.totalLabel == "$2,730")
        #expect(model.budgetLabel == "$50 under plan")
        #expect(model.facts.map(\.value) == ["7", "12", "5", "42 km"])
        model.toggleLoved("Coastal places")
        #expect(model.loved.isEmpty)
        model.toggleNotForMe("Long transfers")
        #expect(model.notForMe.contains("Long transfers"))
    }
}
