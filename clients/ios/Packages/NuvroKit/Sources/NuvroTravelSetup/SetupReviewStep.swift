import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import SwiftUI

public struct TravelProfileSummary {
    public let copy: TravelOptionCopy
    public let locale: Locale

    public init(copy: TravelOptionCopy, locale: Locale = .current) {
        self.copy = copy
        self.locale = locale
    }

    public typealias SummaryLine = (step: TravelSetupStep, key: String, value: String)

    public func lines(for draft: TravelSetupDraft, airports: [Airport]) -> [SummaryLine] {
        var lines: [SummaryLine] = []
        let residence: String = draft.countryOfResidence.map { TravelCatalog.regionName($0, locale: locale) } ?? copy.notAdded
        lines.append((.residence, copy.livesIn, residence))
        let passportNames: [String] = draft.passportCountries.map { TravelCatalog.regionName($0, locale: locale) }
        lines.append((.passports, copy.travelsWith, passportNames.isEmpty ? copy.notAdded : passportNames.joined(separator: ", ")))
        let permits: [String] = draft.residencePermits.map { code in TravelCatalog.permitZones.first { $0.id == code }?.label ?? TravelCatalog.regionName(code, locale: locale) }
        if !permits.isEmpty { lines.append((.permits, copy.permits, permits.joined(separator: ", "))) }
        let airport: String = airports.first { $0.code == draft.departureAirport }?.shortLabel ?? copy.notAdded
        lines.append((.airport, copy.leavesFrom, airport))
        let currency: String = draft.currency.map { "\($0) — \(TravelCatalog.currencyName($0, locale: locale))" } ?? copy.notAdded
        lines.append((.currency, copy.currency, currency))
        lines.append((.paperwork, copy.entryPreference, draft.paperwork.map(paperworkLabel) ?? copy.notAdded))
        let interests: [String] = TravelPreference.allCases.filter { draft.interests.contains($0) }.map(\.title)
        lines.append((.interests, copy.travelStyle, interests.isEmpty ? copy.notAdded : interests.joined(separator: " · ")))
        lines.append((.pace, copy.pace, draft.pace.map(paceLabel) ?? copy.notAdded))
        lines.append((.weather, copy.weather, weatherLabel(draft.weather)))
        lines.append((.crowds, copy.crowds, draft.crowds.map(crowdsLabel) ?? copy.notAdded))
        let flights: [String] = FlightPreference.allCases.filter { draft.flights.contains($0) }.map(flightLabel)
        lines.append((.flights, copy.flights, flights.isEmpty ? copy.noPreference : flights.joined(separator: " · ")))
        let stays: [String] = StayPreference.allCases.filter { draft.stays.contains($0) }.map(stayLabel)
        lines.append((.stays, copy.stays, stays.isEmpty ? copy.noPreference : stays.joined(separator: " · ")))
        lines.append((.group, copy.group, draft.group.map(groupLabel) ?? copy.notAdded))
        let access: [String] = AccessibilityPreference.allCases.filter { draft.accessibility.contains($0) }.map(accessibilityLabel)
        lines.append((.accessibility, copy.accessibility, access.isEmpty ? copy.noRequirements : access.joined(separator: " · ")))
        return lines
    }

    public func paperworkLabel(_ value: PaperworkPreference) -> String {
        switch value {
        case .keepSimple: copy.keepSimple
        case .evisaFine: copy.evisaFine
        case .anyDestination: copy.anyDestination
        }
    }

    public func paceLabel(_ value: TravelPace) -> String {
        switch value {
        case .relaxed: copy.relaxed
        case .balanced: copy.balanced
        case .full: copy.full
        }
    }

    public func weatherLabel(_ values: Set<WeatherPreference>) -> String {
        let ordered = WeatherPreference.allCases.filter(values.contains)
        if ordered.isEmpty { return copy.notAdded }
        if ordered == [.mild, .warm] || ordered == [.warm, .mild] { return "\(copy.mild) to \(copy.warm.lowercased())" }
        return ordered.map { value in
            switch value {
            case .warm: copy.warm
            case .mild: copy.mild
            case .cool: copy.cool
            case .cold: copy.cold
            case .noPreference: copy.noPreference
            }
        }.joined(separator: " · ")
    }

    public func crowdsLabel(_ value: CrowdPreference) -> String {
        switch value {
        case .quiet: copy.preferQuiet
        case .balanced: copy.crowdsBalanced
        case .fine: copy.crowdsFine
        }
    }

    public func flightLabel(_ value: FlightPreference) -> String {
        switch value {
        case .fewerStops: copy.fewerStops
        case .avoidOvernight: copy.avoidOvernight
        case .shortest: copy.shortest
        case .lowestPrice: copy.lowestPrice
        case .noPreference: copy.noPreference
        }
    }

    public func stayLabel(_ value: StayPreference) -> String {
        switch value {
        case .quietNeighbourhood: copy.quietNeighbourhoods
        case .central: copy.central
        case .nearNightlife: copy.nearNightlife
        case .nearBeach: copy.nearBeach
        case .hotel: copy.hotel
        case .apartment: copy.apartment
        case .resort: copy.resort
        case .budgetFriendly: copy.budgetFriendly
        case .premium: copy.premium
        }
    }

    public func groupLabel(_ value: TravelGroup) -> String {
        switch value {
        case .solo: copy.solo
        case .partner: copy.partner
        case .friends: copy.friends
        case .family: copy.family
        case .changes: copy.itChanges
        }
    }

    public func accessibilityLabel(_ value: AccessibilityPreference) -> String {
        switch value {
        case .none: copy.noRequirements
        case .mobility: copy.mobility
        case .wheelchair: copy.wheelchair
        case .reducedWalking: copy.reducedWalking
        case .vision: copy.vision
        case .hearing: copy.hearing
        case .other: copy.other
        }
    }
}

struct ReviewStep: View {
    @Environment(\.designMetrics) private var metrics
    let model: TravelSetupModel

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(18)) {
            SetupQuestion(question: model.copy.reviewTitle, supporting: model.copy.reviewBody)
            VStack(spacing: 0) {
                let lines = TravelProfileSummary(copy: model.options).lines(for: model.draft, airports: model.airports)
                ForEach(Array(lines.enumerated()), id: \.offset) { index, line in
                    NuvroPickerRow(line.key, value: line.value) { model.edit(line.step) }
                    if index < lines.count - 1 { Divider().overlay(Palette.labelAlpha(0.07)) }
                }
            }
            .surfaceCard()
        }
    }
}
