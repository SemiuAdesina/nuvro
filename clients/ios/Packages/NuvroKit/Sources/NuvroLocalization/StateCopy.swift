public struct StateCopy: Equatable, Sendable {
    public let loading: String
    public let retry: String
    public let offlineTitle: String
    public let offlineBody: String
    public let offlineCachedBody: String
    public let reconnecting: String
    public let unavailableTitle: String
    public let unavailableBody: String
    public let unexpectedTitle: String
    public let unexpectedBody: String
    public let unauthorizedTitle: String
    public let unauthorizedBody: String
    public let signInAgain: String
    public let fresh: String
    public let cached: String
    public let stale: String
    public let unavailableShort: String
    public let updatedJustNow: String
    public let openSettings: String
    public let notNow: String
    public let continueLabel: String
    public let cancel: String
    public let done: String
    public let search: String
    public let noResults: String
    public let noResultsBody: String
    public let clearSearch: String

    public static let english = StateCopy(
        loading: "Loading…", retry: "Try again",
        offlineTitle: "You’re offline", offlineBody: "Connect to load this.", offlineCachedBody: "Some saved trip details are still available.", reconnecting: "Reconnecting…",
        unavailableTitle: "Couldn’t load this right now", unavailableBody: "NUVRO is temporarily unavailable. Your saved details are still here.",
        unexpectedTitle: "Something went wrong", unexpectedBody: "Give it another try.",
        unauthorizedTitle: "Sign in to continue", unauthorizedBody: "Your session ended. Sign in to load this.", signInAgain: "Sign in",
        fresh: "Live", cached: "Saved copy", stale: "May be out of date", unavailableShort: "Unavailable", updatedJustNow: "Updated just now",
        openSettings: "Open Settings", notNow: "Not now", continueLabel: "Continue", cancel: "Cancel", done: "Done",
        search: "Search", noResults: "No matches for this search", noResultsBody: "Try a different spelling or fewer words.", clearSearch: "Clear search"
    )

    public static func resolve(_ language: Language) -> StateCopy { english }
}
