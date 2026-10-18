public struct SavedCopy: Equatable, Sendable {
    public let title: String
    public let compare: String
    public let tabs: [String]
    public let tapTwo: String
    public let selectedHint: String
    public let match: String
    public let emptyTitle: String
    public let emptyBody: String
    public let explore: String
    public let emptyPlacesTitle: String
    public let emptyPlacesBody: String

    public static let english = SavedCopy(
        title: "Saved", compare: "Compare", tabs: ["Destinations", "Stays", "Things to do", "Places"],
        tapTwo: "Tap two destinations to compare them", selectedHint: "selected — tap Compare", match: "MATCH",
        emptyTitle: "Nothing saved yet", emptyBody: "Save destinations you like and compare them here.", explore: "Explore destinations",
        emptyPlacesTitle: "No saved places", emptyPlacesBody: "Places you save during a trip appear here."
    )

    public static func resolve(_ language: Language) -> SavedCopy { english }
}

public struct CompareCopy: Equatable, Sendable {
    public let title: String
    public let better: String
    public let differences: String
    public let choose: String
    public let chosen: String
    public let unavailable: String

    public static let english = CompareCopy(title: "Compare", better: "▲ better", differences: "Biggest differences for you", choose: "Choose", chosen: "Chosen", unavailable: "Not available right now")

    public static func resolve(_ language: Language) -> CompareCopy { english }
}

public struct ProfileCopy: Equatable, Sendable {
    public let title: String
    public let edit: String
    public let travelStyle: String
    public let styleNote: String
    public let learned: String
    public let review: String
    public let useHistory: String
    public let historyOn: String
    public let historyOff: String
    public let trips: String
    public let updated: String
    public let travelAccess: String
    public let residence: String
    public let passport: String
    public let departureAirport: String
    public let paperwork: String
    public let travellerFallback: String
    public let accessNote: String

    public static let english = ProfileCopy(
        title: "Profile", edit: "Edit", travelStyle: "Your travel style", styleNote: "This is what NUVRO uses to rank destinations. Change any of it below.",
        learned: "What NUVRO has learned", review: "Review", useHistory: "Use past trips to improve recommendations",
        historyOn: "On — turn off any time", historyOff: "Off — recommendations use this trip only", trips: "trips", updated: "Preference updated",
        travelAccess: "Travel access", residence: "Residence", passport: "Passport", departureAirport: "Departure airport", paperwork: "Travel paperwork",
        travellerFallback: "Traveller", accessNote: "Used for entry guidance and nearby departures. Never your passport number."
    )

    public static func resolve(_ language: Language) -> ProfileCopy { english }
}

public struct RegionCopy: Equatable, Sendable {
    public let title: String
    public let appLanguage: String
    public let appLanguageNote: String
    public let spoken: String
    public let spokenNote: String
    public let region: String
    public let currency: String
    public let homeAirport: String
    public let units: String
    public let distance: String
    public let temperature: String
    public let time: String
    public let date: String
    public let preview: String

    public static let english = RegionCopy(
        title: "Language & region", appLanguage: "App language", appLanguageNote: "Menus, recommendations and NUVRO’s replies",
        spoken: "Languages I speak", spokenNote: "Used for destination fit and Travel Mode", region: "Region", currency: "Currency", homeAirport: "Home airport",
        units: "Units & formats", distance: "Distance", temperature: "Temperature", time: "Time", date: "Date", preview: "How things will look"
    )

    public static func resolve(_ language: Language) -> RegionCopy { english }
}

public struct SettingsCopy: Equatable, Sendable {
    public let title: String
    public let notifications: String
    public let permissions: String
    public let openSettings: String
    public let legal: [String]
    public let version: String
    public let signOut: String
    public let signOutTitle: String
    public let signOutBody: String
    public let cancel: String
    public let deleteAccount: String
    public let deleteTitle: String
    public let deleteExplainer: String
    public let deleteKeeps: String
    public let deleteConfirmTitle: String
    public let deleteConfirmBody: String
    public let deleteFailed: String
    public let deleting: String
    public let granted: String
    public let denied: String
    public let notRequested: String
    public let restricted: String
    public let appearance: String
    public let appearanceSystem: String
    public let appearanceLight: String
    public let appearanceDark: String
    public let appearanceNote: String

    public static let english = SettingsCopy(
        title: "Settings", notifications: "NOTIFICATIONS", permissions: "PERMISSIONS", openSettings: "Open iOS Settings",
        legal: ["Help", "Privacy Policy", "Terms", "About NUVRO"], version: "NUVRO 1.0 · prototype build",
        signOut: "Sign out", signOutTitle: "Sign out of NUVRO?", signOutBody: "Your trips stay in your account. Saved trip details on this phone remain available offline.", cancel: "Cancel",
        deleteAccount: "Delete account", deleteTitle: "Delete your account",
        deleteExplainer: "Deleting removes your profile, trips, saved places, preferences and downloaded documents from NUVRO. This cannot be undone.",
        deleteKeeps: "Bookings made with partners are not cancelled by deleting your NUVRO account.",
        deleteConfirmTitle: "Delete your account?", deleteConfirmBody: "This permanently deletes everything in your NUVRO account.",
        deleteFailed: "NUVRO couldn’t delete your account right now. Nothing was changed.", deleting: "Deleting…",
        granted: "Allowed", denied: "Off", notRequested: "Ask when needed", restricted: "Restricted",
        appearance: "APPEARANCE", appearanceSystem: "System", appearanceLight: "Light", appearanceDark: "Dark", appearanceNote: "System follows your iPhone’s setting."
    )

    public static func resolve(_ language: Language) -> SettingsCopy { english }
}
