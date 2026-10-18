import NuvroDesignSystem
import NuvroModels
import SwiftUI

struct PaperworkStep: View {
    @Environment(\.designMetrics) private var metrics
    let model: TravelSetupModel

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(14)) {
            SetupQuestion(question: model.copy.paperworkQuestion, supporting: nil, privacy: model.copy.paperworkNote)
            ChoiceRow(title: model.options.keepSimple, detail: model.options.keepSimpleDetail, isSelected: model.draft.paperwork == .keepSimple) { model.update { $0.paperwork = .keepSimple } }
            ChoiceRow(title: model.options.evisaFine, detail: model.options.evisaFineDetail, isSelected: model.draft.paperwork == .evisaFine) { model.update { $0.paperwork = .evisaFine } }
            ChoiceRow(title: model.options.anyDestination, detail: model.options.anyDestinationDetail, isSelected: model.draft.paperwork == .anyDestination) { model.update { $0.paperwork = .anyDestination } }
        }
    }
}

struct InterestsStep: View {
    @Environment(\.designMetrics) private var metrics
    let model: TravelSetupModel

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(18)) {
            SetupQuestion(question: model.copy.interestsQuestion, supporting: nil)
            PreferenceGrid(title: "", countLabel: "\(model.draft.interests.count)", options: TravelPreference.allCases, isSelected: { model.draft.interests.contains($0) }, onToggle: { model.toggle($0, in: \.interests) })
        }
    }
}

struct PaceStep: View {
    @Environment(\.designMetrics) private var metrics
    let model: TravelSetupModel

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(14)) {
            SetupQuestion(question: model.copy.paceQuestion, supporting: nil)
            ChoiceRow(title: model.options.relaxed, detail: model.options.relaxedDetail, isSelected: model.draft.pace == .relaxed) { model.update { $0.pace = .relaxed } }
            ChoiceRow(title: model.options.balanced, detail: model.options.balancedDetail, isSelected: model.draft.pace == .balanced) { model.update { $0.pace = .balanced } }
            ChoiceRow(title: model.options.full, detail: model.options.fullDetail, isSelected: model.draft.pace == .full) { model.update { $0.pace = .full } }
        }
    }
}

struct WeatherStep: View {
    @Environment(\.designMetrics) private var metrics
    let model: TravelSetupModel

    private func label(_ value: WeatherPreference) -> String {
        switch value {
        case .warm: model.options.warm
        case .mild: model.options.mild
        case .cool: model.options.cool
        case .cold: model.options.cold
        case .noPreference: model.options.noPreference
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(14)) {
            SetupQuestion(question: model.copy.weatherQuestion, supporting: nil)
            ForEach(WeatherPreference.allCases, id: \.self) { value in
                ChoiceRow(title: label(value), isSelected: model.draft.weather.contains(value), allowsMultiple: true) { model.selectExclusive(value, exclusive: .noPreference, in: \.weather) }
            }
        }
    }
}

struct CrowdsStep: View {
    @Environment(\.designMetrics) private var metrics
    let model: TravelSetupModel

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(14)) {
            SetupQuestion(question: model.copy.crowdsQuestion, supporting: nil)
            ChoiceRow(title: model.options.preferQuiet, isSelected: model.draft.crowds == .quiet) { model.update { $0.crowds = .quiet } }
            ChoiceRow(title: model.options.crowdsBalanced, isSelected: model.draft.crowds == .balanced) { model.update { $0.crowds = .balanced } }
            ChoiceRow(title: model.options.crowdsFine, isSelected: model.draft.crowds == .fine) { model.update { $0.crowds = .fine } }
        }
    }
}

struct FlightsStep: View {
    @Environment(\.designMetrics) private var metrics
    let model: TravelSetupModel

    private func label(_ value: FlightPreference) -> String {
        switch value {
        case .fewerStops: model.options.fewerStops
        case .avoidOvernight: model.options.avoidOvernight
        case .shortest: model.options.shortest
        case .lowestPrice: model.options.lowestPrice
        case .noPreference: model.options.noPreference
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(14)) {
            SetupQuestion(question: model.copy.flightsQuestion, supporting: nil)
            ForEach(FlightPreference.allCases, id: \.self) { value in
                ChoiceRow(title: label(value), isSelected: model.draft.flights.contains(value), allowsMultiple: true) { model.selectExclusive(value, exclusive: .noPreference, in: \.flights) }
            }
        }
    }
}

struct StaysStep: View {
    @Environment(\.designMetrics) private var metrics
    let model: TravelSetupModel

    private func label(_ value: StayPreference) -> String {
        switch value {
        case .quietNeighbourhood: model.options.quietNeighbourhoods
        case .central: model.options.central
        case .nearNightlife: model.options.nearNightlife
        case .nearBeach: model.options.nearBeach
        case .hotel: model.options.hotel
        case .apartment: model.options.apartment
        case .resort: model.options.resort
        case .budgetFriendly: model.options.budgetFriendly
        case .premium: model.options.premium
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(18)) {
            SetupQuestion(question: model.copy.staysQuestion, supporting: nil)
            FlowLayout(spacing: metrics.points(8), rowSpacing: metrics.points(8)) {
                ForEach(StayPreference.allCases, id: \.self) { value in
                    ChoiceChip(title: label(value), isSelected: model.draft.stays.contains(value)) { model.toggle(value, in: \.stays) }
                }
            }
        }
    }
}

struct GroupStep: View {
    @Environment(\.designMetrics) private var metrics
    let model: TravelSetupModel

    private func label(_ value: TravelGroup) -> String {
        switch value {
        case .solo: model.options.solo
        case .partner: model.options.partner
        case .friends: model.options.friends
        case .family: model.options.family
        case .changes: model.options.itChanges
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(14)) {
            SetupQuestion(question: model.copy.groupQuestion, supporting: nil)
            ForEach(TravelGroup.allCases, id: \.self) { value in
                ChoiceRow(title: label(value), isSelected: model.draft.group == value) { model.update { $0.group = value } }
            }
        }
    }
}

struct AccessibilityStep: View {
    @Environment(\.designMetrics) private var metrics
    let model: TravelSetupModel

    private func label(_ value: AccessibilityPreference) -> String {
        switch value {
        case .none: model.options.noRequirements
        case .mobility: model.options.mobility
        case .wheelchair: model.options.wheelchair
        case .reducedWalking: model.options.reducedWalking
        case .vision: model.options.vision
        case .hearing: model.options.hearing
        case .other: model.options.other
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(14)) {
            SetupQuestion(question: model.copy.accessibilityQuestion, supporting: model.copy.accessibilityBody)
            ForEach(AccessibilityPreference.allCases, id: \.self) { value in
                ChoiceRow(title: label(value), isSelected: model.draft.accessibility.contains(value), allowsMultiple: true) { model.selectExclusive(value, exclusive: .none, in: \.accessibility) }
            }
        }
    }
}
