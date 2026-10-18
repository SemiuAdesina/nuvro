import Foundation

public struct DiscoverCopy: Equatable, Sendable {
    public let goodMorning: String
    public let goodAfternoon: String
    public let goodEvening: String
    public let headline: String
    public let searchPlaceholder: String
    public let heroTitle: String
    public let heroBody: String
    public let heroAction: String
    public let recommendedTitle: String
    public let seeAll: String
    public let priceDisclaimer: String
    public let fromPrice: String
    public let saved: String
    public let removedFromSaved: String
    public let noSuggestionsTitle: String
    public let noSuggestionsBody: String
    public let refreshHint: String

    public static let english = DiscoverCopy(
        goodMorning: "Good morning",
        goodAfternoon: "Good afternoon",
        goodEvening: "Good evening",
        headline: "Where should we take you?",
        searchPlaceholder: "Beach, somewhere warm, under $2,500…",
        heroTitle: "Where can my money take me?",
        heroBody: "Set your budget and see the places you can actually afford.",
        heroAction: "Explore by budget",
        recommendedTitle: "Made for your next escape",
        seeAll: "See all",
        priceDisclaimer: "Prices shown are examples, not live fares.",
        fromPrice: "From"
,
        saved: "Saved", removedFromSaved: "Removed from saved", noSuggestionsTitle: "No picks yet", noSuggestionsBody: "Tell NUVRO what you’re after and picks appear here.", refreshHint: "Pull to refresh"
    )

    public static func resolve(_ language: Language) -> DiscoverCopy { english }
}

public struct MatchSetupCopy: Equatable, Sendable {
    public let duration: String
    public let title: String
    public let body: String
    public let leavingFrom: String
    public let change: String
    public let travelers: String
    public let totalBudget: String
    public let flexible: String
    public let dates: String
    public let exactDates: String
    public let flexibleDates: String
    public let preferencesTitle: String
    public let selected: String
    public let fineTune: String
    public let findMatches: String

    public static let english = MatchSetupCopy(
        duration: "Takes about a minute",
        title: "Build your perfect escape",
        body: "A few details help NUVRO find places that actually fit you.",
        leavingFrom: "Leaving from",
        change: "Change",
        travelers: "Travelers",
        totalBudget: "Total trip budget",
        flexible: "Flexible",
        dates: "Dates",
        exactDates: "Exact dates",
        flexibleDates: "I’m flexible",
        preferencesTitle: "What do you want from this trip?",
        selected: "selected",
        fineTune: "Fine tune",
        findMatches: "Find my matches"
    )

    public static func resolve(_ language: Language) -> MatchSetupCopy { english }
}

public struct MatchResultsCopy: Equatable, Sendable {
    public let editPreferences: String
    public let title: String
    public let body: String
    public let estimatedTrip: String
    public let weather: String
    public let flight: String
    public let whyItFits: String
    public let explore: String
    public let compare: String
    public let comparing: String
    public let otherMatches: String
    public let noMatchesTitle: String
    public let noMatchesBody: String
    public let adjustPreferences: String
    public let ranking: String

    public static let english = MatchResultsCopy(
        editPreferences: "Edit preferences",
        title: "Your best matches",
        body: "Based on your budget, dates, travel style, and preferences.",
        estimatedTrip: "Estimated trip",
        weather: "Weather",
        flight: "Flight",
        whyItFits: "Why it fits you",
        explore: "Explore",
        compare: "Compare",
        comparing: "Comparing",
        otherMatches: "Other strong matches"
,
        noMatchesTitle: "No matches for this search", noMatchesBody: "Widen the budget, dates or preferences to see more places.", adjustPreferences: "Adjust preferences", ranking: "Ranking your matches…"
    )

    public static func resolve(_ language: Language) -> MatchResultsCopy { english }
}

public struct DestinationCopy: Equatable, Sendable {
    public let whyItFits: String
    public let everythingAbout: String
    public let whereToStay: String
    public let thingsToDo: String
    public let onTheMap: String
    public let openMap: String
    public let seeAll: String
    public let buildTrip: String
    public let save: String
    public let saved: String
    public let length: String
    public let trip: String
    public let climate: String
    public let flight: String
    public let pace: String
    public let budget: String
    public let weather: String
    public let styleMatch: String
    public let under: String
    public let stop: String
    public let stops: String

    public static let english = DestinationCopy(
        whyItFits: "Why it fits you",
        everythingAbout: "Everything about",
        whereToStay: "Where should you stay?",
        thingsToDo: "Things to do",
        onTheMap: "On the map",
        openMap: "Open map",
        seeAll: "See all",
        buildTrip: "Build my trip",
        save: "Save",
        saved: "Saved",
        length: "Length",
        trip: "Trip",
        climate: "Climate",
        flight: "Flight",
        pace: "Pace",
        budget: "Budget",
        weather: "Weather",
        styleMatch: "Style match",
        under: "under",
        stop: "stop",
        stops: "stops"
    )

    public static func resolve(_ language: Language) -> DestinationCopy { english }
}

public struct TripPlanCopy: Equatable, Sendable {
    public let planning: String
    public let yourTrip: String
    public let ready: String
    public let estimatedTrip: String
    public let underBudget: String
    public let overBudget: String
    public let viewBudget: String
    public let yourWeek: String
    public let viewItinerary: String
    public let day: String
    public let askNuvro: String
    public let askPlaceholder: String
    public let askSuggestions: [String]
    public let editTrip: String
    public let changeDates: String
    public let changeTravelers: String
    public let changeBudget: String
    public let travellingTogether: String
    public let budgetTitle: String
    public let budgetHint: String
    public let cancel: String
    public let done: String
    public let tripUpdated: String

    public static let english = TripPlanCopy(
        planning: "Planning",
        yourTrip: "Your trip",
        ready: "ready",
        estimatedTrip: "Estimated trip",
        underBudget: "under budget",
        overBudget: "over budget",
        viewBudget: "View budget",
        yourWeek: "Your week",
        viewItinerary: "View itinerary",
        day: "Day",
        askNuvro: "Ask NUVRO",
        askPlaceholder: "Ask anything about this trip…",
        askSuggestions: ["Make Day 3 more relaxing.", "Can we spend less on this trip?"]
,
        editTrip: "Edit trip", changeDates: "Change dates", changeTravelers: "Change travellers", changeBudget: "Change budget", travellingTogether: "Travelling together",
        budgetTitle: "Trip budget", budgetHint: "NUVRO keeps the plan inside this amount.", cancel: "Cancel", done: "Done", tripUpdated: "Trip updated"
    )

    public static func resolve(_ language: Language) -> TripPlanCopy { english }
}
