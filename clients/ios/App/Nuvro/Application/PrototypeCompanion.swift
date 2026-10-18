import NuvroModels

enum PrototypeCompanion {
    private static func media(_ placeholder: String, _ alt: String) -> DestinationMedia {
        DestinationMedia(placeholder: placeholder, altText: alt, attribution: "Photo selected by NUVRO · provider attribution")
    }

    static let languagePair = LanguagePair(source: "English", target: "Português")

    static let sampleTranslation = PhraseTranslation(pair: languagePair, original: "We have a reservation for two.", translated: "Temos uma reserva para duas pessoas.")

    static let phrases = [
        TravelPhrase(context: "Hotel", systemImage: "bed.double.fill", original: "I have a reservation.", translated: "Tenho uma reserva."),
        TravelPhrase(context: "Restaurant", systemImage: "fork.knife", original: "Does this contain peanuts?", translated: "Isto contém amendoins?"),
        TravelPhrase(context: "Transport", systemImage: "bus.fill", original: "How much is the fare?", translated: "Quanto custa o bilhete?"),
        TravelPhrase(context: "Directions", systemImage: "map.fill", original: "Can you show me where this is?", translated: "Pode mostrar-me onde fica isto?"),
        TravelPhrase(context: "Emergency", systemImage: "cross.case.fill", original: "I need help.", translated: "Preciso de ajuda.")
    ]

    static let nearbyPlaces = [
        NearbyPlace(id: "mercado", name: "Mercado dos Lavradores", category: .food, categoryLabel: "Market", walkMinutes: 8, travelLabel: "8 min walk", isOpen: true, hoursLabel: "Open now", reason: "Great match for your food interests", media: media("gold", "Market hall with fruit and flower stalls"), coordinate: GeoCoordinate(latitude: 32.6489, longitude: -16.9056), markerLetter: "F"),
        NearbyPlace(id: "cafe", name: "Barreirinha Bar Café", category: .coffee, categoryLabel: "Coffee", walkMinutes: 6, travelLabel: "6 min walk", isOpen: true, hoursLabel: "Open until 20:00", reason: "Quiet terrace, ocean view", media: media("curacao", "Café terrace above the sea"), coordinate: GeoCoordinate(latitude: 32.6465, longitude: -16.9015), markerLetter: "C"),
        NearbyPlace(id: "museum", name: "Museu de Arte Sacra", category: .indoor, categoryLabel: "Indoor", walkMinutes: 11, travelLabel: "11 min walk", isOpen: true, hoursLabel: "Open now", reason: "Indoor option if the rain arrives", media: media("twilight", "Museum gallery with Flemish paintings"), coordinate: GeoCoordinate(latitude: 32.6477, longitude: -16.9088), markerLetter: "M"),
        NearbyPlace(id: "viewpoint", name: "Miradouro do Pináculo", category: .views, categoryLabel: "Viewpoint", walkMinutes: 18, travelLabel: "18 min by car", isOpen: true, hoursLabel: "Always open", reason: "Best sunset near your route", media: media("madeira", "Viewpoint over Funchal bay at sunset"), coordinate: GeoCoordinate(latitude: 32.6440, longitude: -16.8830), markerLetter: "V")
    ]

    static let freeTime = FreeTimeWindow(minutes: 140, untilLabel: "dinner reservation")

    static let documents = [
        TripDocument(id: "cable", category: .activity, title: "Funchal Cable Car", subtitle: "Today · 10:00", whenLabel: "Today · 10:00", travelers: 2, isAvailableOffline: true, reference: "CBL-4471", canAddToWallet: true, isNeededNext: true),
        TripDocument(id: "out", category: .flight, title: "JFK → FNC", subtitle: "Oct 12 · TP1691", whenLabel: "Oct 12", travelers: 2, isAvailableOffline: true, reference: "TP1691-8842", canAddToWallet: true),
        TripDocument(id: "back", category: .flight, title: "FNC → JFK", subtitle: "Oct 19 · TP1692", whenLabel: "Oct 19", travelers: 2, isAvailableOffline: true, reference: "TP1692-8843", canAddToWallet: true),
        TripDocument(id: "hotel", category: .stay, title: "Quinta do Monte Palace", subtitle: "Oct 12–19 · 2 guests", whenLabel: "Oct 12–19", travelers: 2, isAvailableOffline: true, reference: "QMP-77214"),
        TripDocument(id: "tour", category: .activity, title: "West Island Tour", subtitle: "Oct 15 · 09:00", whenLabel: "Oct 15 · 09:00", travelers: 2, isAvailableOffline: false, reference: "WIT-2038"),
        TripDocument(id: "transfer", category: .transfer, title: "Airport transfer", subtitle: "Oct 12 · 17:05", whenLabel: "Oct 12 · 17:05", travelers: 2, isAvailableOffline: true, reference: "TRF-6610", hasBarcode: false),
        TripDocument(id: "museum", category: .activity, title: "Madeira Story Centre", subtitle: "Oct 13 · used", whenLabel: "Oct 13", travelers: 2, isAvailableOffline: true, reference: "MSC-1180", isExpired: true),
        TripDocument(id: "insurance", category: .other, title: "Travel insurance", subtitle: "Policy · 12–19 Oct", whenLabel: "Oct 12–19", travelers: 2, isAvailableOffline: true, reference: "INS-90112")
    ]

    static let alerts = [
        TripAlert(id: "rain", priority: .now, kind: .weather, severity: .actionNeeded, title: "Rain expected from 3 PM", body: "Your Botanical Garden visit may be affected.", reason: "Affects your Day 3 itinerary", actionLabel: "Adjust plan", destination: .replanning),
        TripAlert(id: "cable", priority: .now, kind: .activity, severity: .onTime, title: "Cable Car in 45 minutes", body: "Leave in about 22 minutes to arrive on time.", reason: "Next item on today’s plan", actionLabel: "Navigate", destination: .travelMode),
        TripAlert(id: "flight", priority: .today, kind: .flight, severity: .updated, title: "Flight time changed", body: "Your return flight now departs 35 minutes later.", reason: "Affects your Oct 19 departure", actionLabel: "View flight", destination: .document),
        TripAlert(id: "entry", priority: .later, kind: .document, severity: .reminder, title: "Check entry requirements", body: "Confirm what your passport needs before departure.", reason: "Based on your route and nationality", actionLabel: "Review", destination: .document),
        TripAlert(id: "summary", priority: .later, kind: .summary, severity: .reminder, title: "Your trip summary will be ready on Oct 19", body: "NUVRO will gather your places, spending, and what you loved.", reason: "Available after your return flight", actionLabel: "Preview summary", destination: .summary)
    ]

    static let summary = TripSummary(
        destinationName: "Madeira", dateRangeLabel: "Oct 12–19", days: 7, placesCount: 12, activitiesCount: 5, kilometresExplored: 42,
        hero: media("madeira", "Madeira coastline at golden hour"),
        places: [
            PlaceMemory(name: "Old Town", media: media("madeira", "Painted doors of Funchal Old Town")),
            PlaceMemory(name: "Botanical Garden", media: media("guanacaste", "Terraced botanical garden")),
            PlaceMemory(name: "Câmara de Lobos", media: media("curacao", "Fishing boats in Câmara de Lobos")),
            PlaceMemory(name: "Cable Car", media: media("sky", "Cable car over Funchal")),
            PlaceMemory(name: "Mercado", media: media("gold", "Market stalls"))
        ],
        total: PrototypeContent.usd(2730), plannedTotal: PrototypeContent.usd(2780),
        lovedOptions: ["Coastal places", "Local food", "Relaxed mornings"], notForMeOptions: ["Long transfers", "Early starts"],
        suggestion: NextTripSuggestion(name: "Azores", country: "Portugal", reason: "Same Atlantic coastline and food scene, fewer visitors in spring", media: media("guanacaste", "Green crater lakes in the Azores"))
    )
}
