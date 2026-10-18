public struct LibraryCopy: Equatable, Sendable {
    public let title: String
    public let planTrip: String
    public let happeningNow: String
    public let comingUp: String
    public let planning: String
    public let past: String
    public let active: String
    public let next: String
    public let continueTrip: String
    public let continuePlanning: String
    public let daysAway: String
    public let planningTag: String
    public let more: String
    public let downloadOffline: String
    public let downloaded: String
    public let archive: String
    public let archived: String
    public let delete: String
    public let deleteTitle: String
    public let deleteBody: String
    public let deleted: String
    public let cancel: String
    public let emptyTitle: String
    public let emptyBody: String
    public let findVacation: String

    public static let english = LibraryCopy(
        title: "Trips", planTrip: "Plan trip", happeningNow: "HAPPENING NOW", comingUp: "COMING UP", planning: "PLANNING", past: "PAST", active: "Active", next: "NEXT",
        continueTrip: "Continue trip", continuePlanning: "Continue planning", daysAway: "days away", planningTag: "Planning",
        more: "More", downloadOffline: "Download for offline", downloaded: "Downloaded for offline", archive: "Archive trip", archived: "Trip archived",
        delete: "Delete trip", deleteTitle: "Delete this trip?", deleteBody: "The plan, itinerary and budget for this trip are removed. Bookings with partners are not cancelled.", deleted: "Trip deleted", cancel: "Cancel",
        emptyTitle: "No trips yet", emptyBody: "Your next trip starts here.", findVacation: "Find a vacation"
    )

    public static func resolve(_ language: Language) -> LibraryCopy { english }
}

public struct FlightsCopy: Equatable, Sendable {
    public let title: String
    public let bestOverall: String
    public let lowestPrice: String
    public let easiestJourney: String
    public let viewFlight: String
    public let otherOptions: String
    public let disclaimer: String
    public let emptyTitle: String
    public let emptyBody: String
    public let pricesChecked: String
    public let pricesStale: String
    public let providerUnavailable: String
    public let leavingTitle: String
    public let leavingBody: String
    public let continueToPartner: String
    public let cancel: String
    public let linkPending: String
    public let affiliateDisclosure: String

    public static let english = FlightsCopy(
        title: "Flights", bestOverall: "Best overall", lowestPrice: "Lowest price", easiestJourney: "Easiest journey", viewFlight: "View flight",
        otherOptions: "Other options", disclaimer: "Prototype fares. Live prices and fare rules are confirmed before booking.",
        emptyTitle: "No flights for these dates", emptyBody: "Try flexible dates or a nearby airport.", pricesChecked: "Prices checked", pricesStale: "Prices may have changed — refresh before booking",
        providerUnavailable: "Flight search is temporarily unavailable",
        leavingTitle: "You’re leaving NUVRO to book with", leavingBody: "Fares, seats and fare rules are confirmed on the partner’s site. Your trip stays in NUVRO.",
        continueToPartner: "Continue to partner", cancel: "Not now", linkPending: "Booking links arrive with live fares.", affiliateDisclosure: "NUVRO may earn from eligible bookings. Ranking never depends on it."
    )

    public static func resolve(_ language: Language) -> FlightsCopy { english }
}

public struct ExperiencesCopy: Equatable, Sendable {
    public let title: String
    public let rankedFor: String
    public let duration: String
    public let from: String
    public let rating: String
    public let viewExperience: String
    public let addToTrip: String
    public let added: String
    public let moreForDates: String
    public let seeAll: String
    public let disclaimer: String
    public let tabs: [String]
    public let emptyTitle: String
    public let emptyBody: String

    public static let english = ExperiencesCopy(
        title: "Things to do", rankedFor: "Ranked for your 7-day trip", duration: "Duration", from: "From", rating: "Rating", viewExperience: "View experience",
        addToTrip: "Add to trip", added: "Added", moreForDates: "More for your dates", seeAll: "See all",
        disclaimer: "Ranked for your trip, never for commission. NUVRO may earn from eligible bookings.",
        tabs: ["For you", "Nature", "Food", "Culture", "Adventure", "Relax", "Family"],
        emptyTitle: "Nothing to do yet", emptyBody: "Experiences for your dates appear here once NUVRO has ranked them."
    )

    public static func resolve(_ language: Language) -> ExperiencesCopy { english }
}

public struct TransportCopy: Equatable, Sendable {
    public let title: String
    public let bestForTrip: String
    public let fromAirport: String
    public let aroundIsland: String
    public let recommended: String
    public let bestForPlan: String
    public let airport: String
    public let hotel: String
    public let disclaimer: String

    public static let english = TransportCopy(
        title: "Getting around", bestForTrip: "Best for your trip", fromAirport: "From the airport", aroundIsland: "Around the island",
        recommended: "Recommended", bestForPlan: "Best for your plan", airport: "FNC airport", hotel: "Hotel",
        disclaimer: "Costs are typical ranges for your dates, not quotes."
    )

    public static func resolve(_ language: Language) -> TransportCopy { english }
}

public struct ChecklistCopy: Equatable, Sendable {
    public let title: String
    public let daysToGo: String
    public let ready: String
    public let noRush: String
    public let basedOnTrip: String
    public let addItem: String
    public let of: String
    public let itemPlaceholder: String
    public let remove: String
    public let allDone: String
    public let sectionDone: String

    public static let english = ChecklistCopy(
        title: "Get ready", daysToGo: "days to go", ready: "ready", noRush: "No rush — nothing is overdue", basedOnTrip: "Based on your trip", addItem: "Add item", of: "of",
        itemPlaceholder: "New item", remove: "Remove", allDone: "Everything is packed", sectionDone: "Done"
    )

    public static func resolve(_ language: Language) -> ChecklistCopy { english }
}

public struct GroupCopy: Equatable, Sendable {
    public let title: String
    public let invite: String
    public let owner: String
    public let editor: String
    public let viewer: String
    public let decideTogether: String
    public let vote: String
    public let votes: String
    public let groupBudget: String
    public let estimatedFor: String
    public let travellers: String
    public let view: String
    public let recentChanges: String
    public let changeRole: String
    public let removeTraveller: String
    public let removeTitle: String
    public let removeBody: String
    public let removed: String
    public let leaveTrip: String
    public let leaveTitle: String
    public let leaveBody: String
    public let cancel: String
    public let inviteTitle: String
    public let inviteBody: String
    public let inviteEmail: String
    public let sendInvite: String
    public let invitePending: String
    public let inviteSent: String
    public let memberActions: String

    public static let english = GroupCopy(
        title: "Travelling together", invite: "Invite traveller", owner: "Owner", editor: "Editor", viewer: "Viewer", decideTogether: "Decide together",
        vote: "vote", votes: "votes", groupBudget: "Group budget", estimatedFor: "estimated", travellers: "travellers", view: "View", recentChanges: "Recent changes",
        changeRole: "Change role", removeTraveller: "Remove traveller", removeTitle: "Remove from this trip?", removeBody: "They lose access to the plan. Their bookings are not affected.", removed: "Traveller removed",
        leaveTrip: "Leave trip", leaveTitle: "Leave this trip?", leaveBody: "You’ll lose access to the shared plan. The owner can invite you again.", cancel: "Cancel",
        inviteTitle: "Invite a traveller", inviteBody: "They’ll get a link to join this trip’s plan.", inviteEmail: "Email address", sendInvite: "Send invite", invitePending: "Invited · pending", inviteSent: "Invitation sent",
        memberActions: "Traveller options"
    )

    public static func resolve(_ language: Language) -> GroupCopy { english }
}

public struct PracticalCopy: Equatable, Sendable {
    public let title: String
    public let entry: String
    public let passportNote: String
    public let openSource: String
    public let emergency: String
    public let emergencyNote: String

    public static let english = PracticalCopy(
        title: "Know before you go", entry: "Entry",
        passportNote: "Requirements depend on your passport, not your home region. Confirm yours against the official source before you travel.",
        openSource: "Open official source", emergency: "Emergency",
        emergencyNote: "112 works across the EU and connects to police, fire and medical services. English is generally available."
    )

    public static func resolve(_ language: Language) -> PracticalCopy { english }
}
