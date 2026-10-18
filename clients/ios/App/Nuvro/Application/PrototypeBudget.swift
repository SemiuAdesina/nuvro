import NuvroModels

enum PrototypeBudget {
    private static func usd(_ major: Int64) -> Money { PrototypeContent.usd(major) }

    static let lines = [
        BudgetLine(category: "Flights", amount: usd(760), booked: usd(760), details: [BudgetDetail(label: "2 × round trip JFK–FNC", amount: usd(720)), BudgetDetail(label: "Seat selection", amount: usd(40))]),
        BudgetLine(category: "Stay", amount: usd(980), booked: usd(940), details: [BudgetDetail(label: "Quinta do Monte · 7 nights", amount: usd(940)), BudgetDetail(label: "City tax", amount: usd(40))]),
        BudgetLine(category: "Food", amount: usd(420), details: [BudgetDetail(label: "Dinners (4 reserved)", amount: usd(240)), BudgetDetail(label: "Daily meals", amount: usd(180))]),
        BudgetLine(category: "Activities", amount: usd(310), details: [BudgetDetail(label: "Cable car · 2 tickets", amount: usd(44)), BudgetDetail(label: "Levada guide", amount: usd(130)), BudgetDetail(label: "Museum + garden entry", amount: usd(136))]),
        BudgetLine(category: "Transport", amount: usd(160), details: [BudgetDetail(label: "Rental car · 2 days", amount: usd(110)), BudgetDetail(label: "Airport transfers", amount: usd(50))]),
        BudgetLine(category: "Other", amount: usd(150), details: [BudgetDetail(label: "eSIM · 7 days", amount: usd(18)), BudgetDetail(label: "Buffer", amount: usd(132))])
    ]

    static let stays = [
        StayOption(id: "quinta", name: "Quinta do Monte Palace", area: "Funchal", distanceNote: "900 m from Old Town", nightlyRate: usd(140), nights: 7, rating: 4.7, minutesToPlans: 9, recommendation: .bestOverall, photoReference: "gold",
                   reasons: ["$980 for 7 nights — inside your stay budget", "9 min average to your planned activities", "Quiet street, matches your calmer pace"], coordinate: GeoCoordinate(latitude: 32.6497, longitude: -16.9086)),
        StayOption(id: "carmo", name: "Hotel do Carmo", area: "Funchal centre", distanceNote: "400 m to market", nightlyRate: usd(96), nights: 7, rating: 4.3, minutesToPlans: 6, recommendation: .bestValue, photoReference: "gold",
                   reasons: ["Lowest total for a central stay"], coordinate: GeoCoordinate(latitude: 32.6488, longitude: -16.9070)),
        StayOption(id: "ponta", name: "Estalagem da Ponta", area: "Câmara de Lobos", distanceNote: "sea view", nightlyRate: usd(128), nights: 7, rating: 4.6, minutesToPlans: 18, recommendation: .bestLocation, photoReference: "curacao",
                   reasons: ["Cliff-top views, quieter village"], coordinate: GeoCoordinate(latitude: 32.6480, longitude: -16.9770)),
        StayOption(id: "reids", name: "Reid’s Cliff House", area: "Estrada Monumental", distanceNote: "spa", nightlyRate: usd(310), nights: 7, rating: 4.9, minutesToPlans: 12, recommendation: .bestPremium, photoReference: "twilight",
                   reasons: ["Landmark hotel with a spa"], coordinate: GeoCoordinate(latitude: 32.6410, longitude: -16.9260))
    ]
}
