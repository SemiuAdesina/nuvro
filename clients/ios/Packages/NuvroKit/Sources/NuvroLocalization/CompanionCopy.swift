public struct LanguageCopy: Equatable, Sendable {
    public let title: String
    public let speak: String
    public let type: String
    public let camera: String
    public let phrases: String
    public let saySomething: String
    public let willTranslate: String
    public let listening: String
    public let play: String
    public let showLarge: String
    public let copy: String
    public let tapToReturn: String
    public let usefulNow: String
    public let cameraEntry: String
    public let cameraBody: String
    public let offline: String
    public let typePlaceholder: String
    public let translate: String
    public let spokenHere: String
    public let listeningSub: String
    public let translationPlaceholder: String
    public let showToSomeone: String
    public let changeLanguages: String
    public let phrasesContext: String
    public let copied: String

    public static let english = LanguageCopy(
        title: "Language", speak: "Speak", type: "Type", camera: "Camera", phrases: "Phrases", saySomething: "Say something",
        willTranslate: "NUVRO will translate it into", listening: "Listening…", play: "Play", showLarge: "Show large", copy: "Copy",
        tapToReturn: "Tap to return", usefulNow: "Useful right now", cameraEntry: "Translate a menu or sign", cameraBody: "media · restaurant menu, funchal",
        offline: "Essential phrases available offline", typePlaceholder: "Type in English…", translate: "Translate",
        spokenHere: "is spoken here", listeningSub: "Speak naturally, then pause.", translationPlaceholder: "Translation appears here",
        showToSomeone: "Show to someone", changeLanguages: "Change languages", phrasesContext: "Madeira · Day 3",
        copied: "Copied"
    )

    public static func resolve(_ language: Language) -> LanguageCopy { english }
}

public struct NearbyCopy: Equatable, Sendable {
    public let title: String
    public let aroundYou: String
    public let searchPlaceholder: String
    public let map: String
    public let list: String
    public let openNow: String
    public let closed: String
    public let minWalk: String
    public let go: String
    public let addToToday: String
    public let freeTime: String
    public let fitBefore: String
    public let placesNearby: String
    public let photoBy: String
    public let free: String
    public let until: String
    public let placesNearYou: String
    public let addedToToday: String
    public let theseFit: String
    public let noResultsTitle: String
    public let noResultsBody: String
    public let clearFilters: String

    public static let english = NearbyCopy(
        title: "Explore nearby", aroundYou: "Around you", searchPlaceholder: "What do you feel like doing?", map: "Map", list: "List",
        openNow: "Open now", closed: "Closed", minWalk: "min walk", go: "Go", addToToday: "Add to today", freeTime: "You have", fitBefore: "These places fit before your",
        placesNearby: "places nearby", photoBy: "Photo", free: "free", until: "until your", placesNearYou: "places near you", addedToToday: "Added to today", theseFit: "These fit before your 19:30 dinner",
        noResultsTitle: "Nothing nearby for that", noResultsBody: "Try another filter or clear it to see everything around you.", clearFilters: "Clear filter"
    )

    public static func resolve(_ language: Language) -> NearbyCopy { english }
}

public struct WalletCopy: Equatable, Sendable {
    public let title: String
    public let subtitle: String
    public let neededNext: String
    public let showTicket: String
    public let travelers: String
    public let availableOffline: String
    public let notDownloaded: String
    public let reference: String
    public let tapToReveal: String
    public let addToWallet: String
    public let flights: String
    public let stay: String
    public let activities: String
    public let transfers: String
    public let other: String
    public let scanAtEntrance: String
    public let done: String
    public let documents: String
    public let document: String
    public let downloadAll: String
    public let addAllToWallet: String
    public let footer: String
    public let tapToClose: String
    public let downloaded: String
    public let download: String
    public let removeDownload: String
    public let removeDownloadTitle: String
    public let removeDownloadBody: String
    public let downloadRemoved: String
    public let cancel: String
    public let emptyTitle: String
    public let emptyBody: String
    public let allOffline: String
    public let expired: String
    public let expiredBody: String
    public let noBarcodeTitle: String
    public let noBarcodeBody: String
    public let walletUnavailable: String

    public static let english = WalletCopy(
        title: "Tickets & documents", subtitle: "Your trip", neededNext: "NEEDED NEXT", showTicket: "Show ticket", travelers: "travellers",
        availableOffline: "Available offline", notDownloaded: "Not downloaded", reference: "Ref", tapToReveal: "Tap to reveal", addToWallet: "Add to Wallet",
        flights: "Flights", stay: "Stay", activities: "Activities", transfers: "Transfers", other: "Other", scanAtEntrance: "Scan at the entrance", done: "Done",
        documents: "documents", document: "document", downloadAll: "Download all for offline", addAllToWallet: "Add all to Wallet",
        footer: "Downloaded documents open without a connection. Reference numbers stay masked until you tap them.", tapToClose: "Tap to close",
        downloaded: "Downloaded for offline", download: "Download for offline", removeDownload: "Remove download", removeDownloadTitle: "Remove the offline copy?",
        removeDownloadBody: "You’ll need a connection to open this again. The booking itself is unchanged.", downloadRemoved: "Offline copy removed", cancel: "Cancel",
        emptyTitle: "No documents yet", emptyBody: "Confirmed tickets and bookings for this trip appear here.", allOffline: "Everything is available offline",
        expired: "Expired", expiredBody: "This ticket is no longer valid.", noBarcodeTitle: "No scannable code", noBarcodeBody: "Show the reference number at the entrance.",
        walletUnavailable: "Wallet passes arrive once bookings are issued by the live service"
    )

    public static func resolve(_ language: Language) -> WalletCopy { english }
}

public struct AlertsCopy: Equatable, Sendable {
    public let title: String
    public let subtitle: String
    public let now: String
    public let today: String
    public let later: String
    public let actionNeeded: String
    public let onTime: String
    public let reminder: String
    public let updated: String
    public let needsDecision: String
    public let goodToKnow: String
    public let noActionYet: String
    public let preferences: String
    public let preferencesBody: String
    public let dismiss: String
    public let allClear: String
    public let preferencesTitle: String
    public let done: String
    public let allCaughtUp: String
    public let allCaughtUpBody: String

    public static let english = AlertsCopy(
        title: "Trip updates", subtitle: "Only what matters for your trip", now: "Now", today: "Today", later: "Later",
        actionNeeded: "Action needed", onTime: "On time", reminder: "Reminder", updated: "Updated",
        needsDecision: "Needs a decision", goodToKnow: "Good to know", noActionYet: "No action yet", preferences: "Alert preferences",
        preferencesBody: "Flight changes · Weather · Reservations · Trip reminders · Price tracking", dismiss: "Dismiss", allClear: "Nothing needs your attention right now.",
        preferencesTitle: "Alert preferences", done: "Done", allCaughtUp: "You’re all caught up", allCaughtUpBody: "Nothing needs your attention right now."
    )

    public static func resolve(_ language: Language) -> AlertsCopy { english }
}

public struct SummaryCopy: Equatable, Sendable {
    public let headline: String
    public let days: String
    public let places: String
    public let activities: String
    public let explored: String
    public let placesVisited: String
    public let tripTotal: String
    public let underBudget: String
    public let overBudget: String
    public let learningTitle: String
    public let loved: String
    public let notForMe: String
    public let nextTitle: String
    public let planAnother: String
    public let viewTrip: String
    public let share: String
    public let learningBody: String
    public let footer: String

    public static let english = SummaryCopy(
        headline: "What a week.", days: "days", places: "places", activities: "activities", explored: "explored", placesVisited: "Places you went",
        tripTotal: "Trip total", underBudget: "under plan", overBudget: "over plan",
        learningTitle: "Help us make the next trip better", loved: "Loved", notForMe: "Not for me", nextTitle: "You might love next",
        planAnother: "Plan another trip", viewTrip: "View trip", share: "Share",
        learningBody: "Tap to confirm or remove. NUVRO only learns what you keep.", footer: "Destination photography is selected per trip and season by NUVRO, with provider attribution."
    )

    public static func resolve(_ language: Language) -> SummaryCopy { english }
}
