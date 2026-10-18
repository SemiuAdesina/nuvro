import Foundation
import NuvroDiscover
import NuvroModels

enum PrototypeContent {
    static let locale = Locale(identifier: "en_US")

    static func usd(_ major: Int64) -> Money {
        try! Money(amountMinor: major * 100, currency: "USD")
    }

    static let budgetCeiling = usd(3000)

    static let madeira = try! DestinationMatch(
        id: "madeira", name: "Madeira", country: "Portugal", region: "Atlantic island", photoReference: "madeira",
        score: MatchScore(percent: 94),
        estimate: TripEstimate(low: usd(2640), high: usd(2940)),
        weather: WeatherSummary(condition: "Warm", temperatureCelsius: 23),
        flight: FlightSummary(stops: 1, approximateHours: 9),
        durationDays: 7, pace: "Easygoing", tags: ["Coastal", "Food", "Relaxing"],
        reasons: [
            MatchReason(text: "Within your budget", verdict: .fits),
            MatchReason(text: "Great weather for your dates", verdict: .fits),
            MatchReason(text: "Strong beach + food match", verdict: .fits),
            MatchReason(text: "Good for couples", verdict: .fits),
            MatchReason(text: "Longer flight than preferred", verdict: .tradeoff)
        ]
    )

    static let lisbon = try! DestinationMatch(
        id: "lisbon", name: "Lisbon", country: "Portugal", region: "Capital city", photoReference: "gold",
        score: MatchScore(percent: 91), estimate: TripEstimate(low: usd(2400), high: usd(2700)),
        weather: WeatherSummary(condition: "Mild", temperatureCelsius: 21), flight: FlightSummary(stops: 0, approximateHours: 7),
        durationDays: 6, pace: "Lively", tags: ["City", "Food"],
        reasons: [MatchReason(text: "Direct flight", verdict: .fits), MatchReason(text: "Busier than your usual pace", verdict: .tradeoff)]
    )

    static let curacao = try! DestinationMatch(
        id: "curacao", name: "Curaçao", country: "Caribbean", region: "Dutch Caribbean", photoReference: "curacao",
        score: MatchScore(percent: 88), estimate: TripEstimate(low: usd(2700), high: usd(3100)),
        weather: WeatherSummary(condition: "Hot", temperatureCelsius: 29), flight: FlightSummary(stops: 1, approximateHours: 6),
        durationDays: 7, pace: "Relaxed", tags: ["Beach", "Relaxed"],
        reasons: [MatchReason(text: "Warmest option for October", verdict: .fits), MatchReason(text: "Top of your budget", verdict: .tradeoff)]
    )

    static let kyoto = try! DestinationMatch(
        id: "kyoto", name: "Kyoto", country: "Japan", region: "Kansai", photoReference: "kyoto",
        score: MatchScore(percent: 85), estimate: TripEstimate(low: usd(2300), high: usd(2900)),
        weather: WeatherSummary(condition: "Mild", temperatureCelsius: 19), flight: FlightSummary(stops: 1, approximateHours: 16),
        durationDays: 9, pace: "Balanced", tags: ["Culture", "Food"],
        reasons: [MatchReason(text: "Autumn colours peak in your window", verdict: .fits), MatchReason(text: "Long flight", verdict: .tradeoff)]
    )

    static let guanacaste = try! DestinationMatch(
        id: "guanacaste", name: "Guanacaste", country: "Costa Rica", region: "Pacific coast", photoReference: "guanacaste",
        score: MatchScore(percent: 86), estimate: TripEstimate(low: usd(1800), high: usd(2300)),
        weather: WeatherSummary(condition: "Hot", temperatureCelsius: 30), flight: FlightSummary(stops: 1, approximateHours: 8),
        durationDays: 6, pace: "Adventurous", tags: ["Nature", "Adventure"],
        reasons: [MatchReason(text: "Quiet season, good rates", verdict: .fits), MatchReason(text: "Rainy afternoons in October", verdict: .tradeoff)]
    )

    static let suggestions = [madeira, kyoto, guanacaste]
    static let matches = [madeira, lisbon, curacao, guanacaste]

    static let madeiraGuide = DestinationGuide(
        destinationID: "madeira",
        centre: GeoCoordinate(latitude: 32.66, longitude: -16.95),
        areas: [
            StayArea(name: "Funchal", summary: "Capital · harbour · old town", highlight: "Best overall", photoReference: "madeira"),
            StayArea(name: "Câmara de Lobos", summary: "Fishing village · cliffs", highlight: "Quieter & scenic", photoReference: "twilight"),
            StayArea(name: "Calheta", summary: "Golden sand · west coast", highlight: "Best for beaches", photoReference: "curacao")
        ],
        activities: [
            Activity(name: "Levada das 25 Fontes", summary: "Half-day walk · waterfalls", photoReference: "guanacaste"),
            Activity(name: "Cabo Girão skywalk", summary: "1h · Europe’s highest sea cliff", photoReference: "sky"),
            Activity(name: "Mercado dos Lavradores", summary: "Morning · fruit & fish market", photoReference: "gold")
        ],
        places: [
            MapPlace(name: "Stay", kind: .stay, coordinate: GeoCoordinate(latitude: 32.6497, longitude: -16.9086)),
            MapPlace(name: "Cabo Girão", kind: .attraction, coordinate: GeoCoordinate(latitude: 32.6553, longitude: -17.0055)),
            MapPlace(name: "Market", kind: .restaurant, coordinate: GeoCoordinate(latitude: 32.6478, longitude: -16.9049)),
            MapPlace(name: "Saved", kind: .saved, coordinate: GeoCoordinate(latitude: 32.6939, longitude: -16.8310))
        ]
    )

    static let tripStart: Date = {
        var components = DateComponents()
        components.year = 2026; components.month = 10; components.day = 12
        return Calendar(identifier: .gregorian).date(from: components) ?? .now
    }()

    static let madeiraPlan = TripPlan(
        id: "trip-madeira", destination: madeira, dateRangeLabel: "Oct 12–19", travelersLabel: "2 adults", startDate: tripStart, budgetCeiling: budgetCeiling,
        readiness: [
            ReadinessItem(name: "Destination", isReady: true), ReadinessItem(name: "Dates", isReady: true),
            ReadinessItem(name: "Stay", isReady: true), ReadinessItem(name: "Basic itinerary", isReady: true),
            ReadinessItem(name: "Budget", isReady: true), ReadinessItem(name: "Flight", isReady: false),
            ReadinessItem(name: "Activities", isReady: false), ReadinessItem(name: "Stay connected", isReady: false)
        ],
        budgetLines: PrototypeBudget.lines,
        days: PrototypeItinerary.days
    )
}

extension PrototypeContent {
    static let airports = [
        Airport(code: "JFK", city: "New York", name: "John F. Kennedy International", country: "United States"),
        Airport(code: "EWR", city: "Newark", name: "Newark Liberty International", country: "United States"),
        Airport(code: "BOS", city: "Boston", name: "Logan International", country: "United States"),
        Airport(code: "LAX", city: "Los Angeles", name: "Los Angeles International", country: "United States"),
        Airport(code: "YYZ", city: "Toronto", name: "Toronto Pearson", country: "Canada"),
        Airport(code: "LHR", city: "London", name: "Heathrow", country: "United Kingdom"),
        Airport(code: "CDG", city: "Paris", name: "Charles de Gaulle", country: "France"),
        Airport(code: "LIS", city: "Lisbon", name: "Humberto Delgado", country: "Portugal"),
        Airport(code: "MAD", city: "Madrid", name: "Adolfo Suárez Madrid–Barajas", country: "Spain"),
        Airport(code: "FRA", city: "Frankfurt", name: "Frankfurt Airport", country: "Germany"),
        Airport(code: "DXB", city: "Dubai", name: "Dubai International", country: "United Arab Emirates"),
        Airport(code: "LOS", city: "Lagos", name: "Murtala Muhammed International", country: "Nigeria"),
        Airport(code: "NRT", city: "Tokyo", name: "Narita International", country: "Japan"),
        Airport(code: "SYD", city: "Sydney", name: "Kingsford Smith", country: "Australia")
    ]
}

extension PrototypeContent {
    static func budgetScale(for currency: String) -> BudgetScale {
        switch currency {
        case "NGN": BudgetScale(range: 500_000...12_000_000, step: 50_000, initial: 3_000_000)
        case "KES": BudgetScale(range: 60_000...1_200_000, step: 5_000, initial: 300_000)
        case "GHS": BudgetScale(range: 6_000...120_000, step: 500, initial: 30_000)
        case "ZAR": BudgetScale(range: 8_000...160_000, step: 500, initial: 45_000)
        case "JPY": BudgetScale(range: 80_000...1_500_000, step: 10_000, initial: 400_000)
        case "INR": BudgetScale(range: 40_000...800_000, step: 5_000, initial: 200_000)
        case "BRL": BudgetScale(range: 3_000...50_000, step: 250, initial: 12_000)
        case "AED": BudgetScale(range: 2_500...35_000, step: 250, initial: 10_000)
        case "GBP": BudgetScale(range: 600...6_500, step: 50, initial: 2_400)
        case "EUR": BudgetScale(range: 700...7_500, step: 50, initial: 2_800)
        case "CAD", "AUD": BudgetScale(range: 1_000...11_000, step: 100, initial: 4_000)
        default: BudgetScale(range: 800...8_000, step: 100, initial: 3_000)
        }
    }
}
