public struct PickerCopy: Equatable, Sendable {
    public let dates: String
    public let exactDates: String
    public let flexibleDates: String
    public let startDate: String
    public let endDate: String
    public let month: String
    public let nights: String
    public let anytimeIn: String
    public let about: String
    public let travelers: String
    public let adult: String
    public let adults: String
    public let adultsDetail: String
    public let child: String
    public let children: String
    public let childrenDetail: String
    public let infant: String
    public let infants: String
    public let infantsDetail: String
    public let departureAirport: String
    public let searchAirports: String
    public let noAirports: String
    public let noAirportsBody: String
    public let language: String
    public let searchLanguages: String
    public let noLanguages: String
    public let currency: String
    public let searchCurrencies: String
    public let region: String
    public let searchRegions: String
    public let noMatches: String
    public let noMatchesBody: String

    public static let english = PickerCopy(
        dates: "Travel dates", exactDates: "Exact dates", flexibleDates: "I’m flexible", startDate: "Start", endDate: "End", month: "Month", nights: "Nights", anytimeIn: "Anytime in", about: "about",
        travelers: "Travelers", adult: "adult", adults: "Adults", adultsDetail: "Ages 18+", child: "child", children: "Children", childrenDetail: "Ages 2–17", infant: "infant", infants: "Infants", infantsDetail: "Under 2",
        departureAirport: "Departure airport", searchAirports: "City, airport or code", noAirports: "No airports found", noAirportsBody: "Try the city name or the three-letter code.",
        language: "Language", searchLanguages: "Search languages", noLanguages: "No languages found", currency: "Currency", searchCurrencies: "Search currencies", region: "Region", searchRegions: "Search regions",
        noMatches: "No matches", noMatchesBody: "Try a different spelling."
    )

    public static func resolve(_ language: Language) -> PickerCopy { english }
}
