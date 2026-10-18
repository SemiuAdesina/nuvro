import Foundation
import NuvroAPI
import NuvroDiscover
import NuvroModels

enum DiscoverBriefs {
    static func home(profile: TravelProfile?, today: Date = .now, calendar: Calendar = .current) -> DiscoverBrief {
        let currency = profile?.preferredCurrency ?? "USD"
        let start = calendar.date(byAdding: .day, value: 30, to: today) ?? today
        let end = calendar.date(byAdding: .day, value: 7, to: start) ?? start
        return DiscoverBrief(
            budgetMinor: Int64(PrototypeContent.budgetScale(for: currency).initial * 100), currency: currency,
            startDate: start, endDate: end, adults: 2,
            departureAirport: profile?.preferredDepartureAirport,
            residenceCountry: profile?.countryOfResidence,
            passportCountries: profile?.passportCountries ?? [],
            interests: profile.map { $0.interests.map { $0.rawValue.lowercased() } } ?? []
        )
    }

    @MainActor
    static func search(_ setup: MatchSetupModel, profile: TravelProfile?, calendar: Calendar = .current) -> DiscoverBrief {
        let start = setup.dateSelection.exactStart ?? setup.dateSelection.flexibleMonth ?? setup.today
        let end = calendar.date(byAdding: .day, value: max(1, setup.dateSelection.nights), to: start) ?? start
        return DiscoverBrief(
            budgetMinor: setup.budget?.amountMinor ?? 0, currency: setup.currency,
            startDate: start, endDate: end, adults: setup.adults,
            departureAirport: setup.departureAirport?.code ?? profile?.preferredDepartureAirport,
            residenceCountry: profile?.countryOfResidence,
            passportCountries: profile?.passportCountries ?? [],
            interests: setup.preferences.map { $0.rawValue.lowercased() }
        )
    }
}
