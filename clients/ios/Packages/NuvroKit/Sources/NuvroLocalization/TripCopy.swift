import Foundation

public struct ItineraryCopy: Equatable, Sendable {
    public let map: String
    public let more: String
    public let day: String
    public let move: String
    public let replace: String
    public let remove: String
    public let viewOnMap: String
    public let askNuvro: String
    public let askExample: String
    public let optimize: String
    public let viewBudget: String
    public let booked: String
    public let reserved: String
    public let ticket: String
    public let weather: String
    public let removed: String
    public let movedTo: String
    public let removeTitle: String
    public let removeBody: String
    public let cancel: String
    public let moveTitle: String
    public let done: String
    public let emptyDayTitle: String
    public let emptyDayBody: String

    public static let english = ItineraryCopy(
        map: "Map", more: "More", day: "Day", move: "Move", replace: "Replace", remove: "Remove", viewOnMap: "View on map",
        askNuvro: "Ask NUVRO", askExample: "“Make this day more relaxing.”", optimize: "Optimize this day", viewBudget: "View budget",
        booked: "Booked", reserved: "Reserved", ticket: "Ticket", weather: "Weather",
        removed: "Removed from", movedTo: "Moved to", removeTitle: "Remove this reservation?", removeBody: "It comes off your plan. The booking itself isn’t cancelled — do that with the provider.",
        cancel: "Cancel", moveTitle: "Move to", done: "Done", emptyDayTitle: "Nothing planned yet", emptyDayBody: "Add something or ask NUVRO to fill this day."
    )

    public static func resolve(_ language: Language) -> ItineraryCopy { english }
}

public struct BudgetCopy: Equatable, Sendable {
    public let title: String
    public let estimatedTotal: String
    public let budget: String
    public let remaining: String
    public let over: String
    public let ofBudget: String
    public let withinBudget: String
    public let overBudget: String
    public let booked: String
    public let estimated: String
    public let saveAbout: String
    public let insight: String
    public let seeOption: String
    public let disclaimer: String
    public let viewStays: String
    public let adjustBudget: String

    public static let english = BudgetCopy(
        title: "Trip budget", estimatedTotal: "Estimated trip total", budget: "Budget", remaining: "remaining", over: "over", ofBudget: "of budget",
        withinBudget: "Within your budget — flights and stay are already covered.", overBudget: "Over your budget — review the largest categories below.",
        booked: "Booked", estimated: "Estimated", saveAbout: "Save about",
        insight: "Staying one neighbourhood farther from the centre could reduce your stay cost without adding much travel time.",
        seeOption: "See option", disclaimer: "Estimates only — not live prices.", viewStays: "View stay options", adjustBudget: "Adjust budget"
    )

    public static func resolve(_ language: Language) -> BudgetCopy { english }
}

public struct StayCopy: Equatable, Sendable {
    public let title: String
    public let list: String
    public let map: String
    public let bestOverall: String
    public let bestValue: String
    public let bestLocation: String
    public let bestPremium: String
    public let perNight: String
    public let nights: String
    public let rating: String
    public let toPlans: String
    public let whyPicked: String
    public let viewStay: String
    public let save: String
    public let saved: String
    public let otherFits: String
    public let yourActivities: String
    public let disclosure: String
    public let compare: String

    public static let english = StayCopy(
        title: "Where to stay", list: "List", map: "Map", bestOverall: "Best overall", bestValue: "Best value", bestLocation: "Best location", bestPremium: "Best premium",
        perNight: "per night", nights: "nights", rating: "Rating", toPlans: "To plans", whyPicked: "Why NUVRO picked this", viewStay: "View stay",
        save: "Save", saved: "Saved", otherFits: "Other good fits", yourActivities: "YOUR ACTIVITIES",
        disclosure: "Recommendations are ranked for your trip. NUVRO may earn from eligible bookings.", compare: "Compare stays"
    )

    public static func resolve(_ language: Language) -> StayCopy { english }
}

public struct TravelCopy: Equatable, Sendable {
    public let goodMorning: String
    public let goodAfternoon: String
    public let goodEvening: String
    public let dayIn: String
    public let next: String
    public let leaveIn: String
    public let minutes: String
    public let navigate: String
    public let ticket: String
    public let tickets: String
    public let translate: String
    public let askNuvro: String
    public let weather: String
    public let transport: String
    public let reservations: String
    public let normal: String
    public let today: String
    public let fullItinerary: String
    public let reviewPlan: String
    public let updated: String

    public static let english = TravelCopy(
        goodMorning: "Good morning", goodAfternoon: "Good afternoon", goodEvening: "Good evening", dayIn: "in", next: "NEXT", leaveIn: "Leave in", minutes: "min",
        navigate: "Navigate", ticket: "Ticket", tickets: "Tickets", translate: "Translate", askNuvro: "Ask NUVRO",
        weather: "Weather", transport: "Transport", reservations: "Reservations", normal: "Normal", today: "Today", fullItinerary: "Full itinerary",
        reviewPlan: "Review plan", updated: "Updated"
    )

    public static func resolve(_ language: Language) -> TravelCopy { english }
}

public struct ReplanCopy: Equatable, Sendable {
    public let title: String
    public let body: String
    public let before: String
    public let suggested: String
    public let applied: String
    public let whyThisChange: String
    public let staysTheSame: String
    public let applyPlan: String
    public let keepOriginal: String
    public let anotherOption: String
    public let showingOption: String
    public let planUpdated: String

    public static let english = ReplanCopy(
        title: "Adjust today", body: "NUVRO can reorganize the plan around the weather.", before: "BEFORE", suggested: "SUGGESTED", applied: "APPLIED",
        whyThisChange: "Why this change", staysTheSame: "What stays the same", applyPlan: "Apply new plan", keepOriginal: "Keep original",
        anotherOption: "Show another option", showingOption: "Showing option 2 of 2", planUpdated: "Plan updated · returning to today"
    )

    public static func resolve(_ language: Language) -> ReplanCopy { english }
}
