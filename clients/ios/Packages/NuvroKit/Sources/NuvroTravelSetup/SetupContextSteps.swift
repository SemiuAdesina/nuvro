import NuvroDesignSystem
import NuvroModels
import SwiftUI

struct WelcomeStep: View {
    @Environment(\.designMetrics) private var metrics
    let model: TravelSetupModel

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(18)) {
            RouteArcGlyph(width: metrics.points(54))
            SetupQuestion(question: model.copy.welcomeTitle, supporting: model.copy.welcomeBody)
            Text(model.copy.changeAnytime).typeStyle(Typography.caption).foregroundStyle(Palette.labelAlpha(0.5))
        }
        .padding(.top, metrics.points(40))
    }
}

struct ResidenceStep: View {
    @Environment(\.designMetrics) private var metrics
    let model: TravelSetupModel

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(18)) {
            SetupQuestion(question: model.copy.residenceQuestion, supporting: model.copy.residenceBody, privacy: model.copy.residencePrivacy)
            if let region = model.region(model.draft.countryOfResidence) {
                SelectionCard(leading: region.code, title: region.name, actionLabel: model.copy.change) { model.picker = .residence }
            } else {
                AddRow(title: model.copy.chooseCountry) { model.picker = .residence }
            }
        }
    }
}

struct PassportStep: View {
    @Environment(\.designMetrics) private var metrics
    let model: TravelSetupModel

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(14)) {
            SetupQuestion(question: model.copy.passportQuestion, supporting: model.copy.passportBody, privacy: model.copy.passportPrivacy)
            ForEach(model.draft.passportCountries, id: \.self) { code in
                if let region = model.region(code) {
                    SelectionCard(leading: region.code, title: region.name, detail: model.options.passport.capitalized, actionLabel: model.copy.change, action: { model.replacePassport(code) }, onRemove: { model.removePassport(code) })
                }
            }
            AddRow(title: model.draft.passportCountries.isEmpty ? model.copy.addPassport : model.copy.addAnotherPassport) { model.picker = .passport }
            if model.draft.passportCountries.isEmpty {
                ChoiceRow(title: model.copy.declinePassport, detail: model.copy.passportDeclinedNote, isSelected: model.draft.hasDeclinedPassport) { model.declinePassport() }
            }
        }
        .animation(Motion.fade(0.25), value: model.draft.passportCountries)
    }
}

struct PermitStep: View {
    @Environment(\.designMetrics) private var metrics
    let model: TravelSetupModel

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(14)) {
            SetupQuestion(question: model.copy.permitQuestion, supporting: model.copy.permitBody)
            ChoiceRow(title: model.copy.permitYes, isSelected: model.draft.permitAnswer == .yes) { model.update { $0.permitAnswer = .yes } }
            ChoiceRow(title: model.copy.permitNo, isSelected: model.draft.permitAnswer == .no) { model.update { $0.permitAnswer = .no; $0.residencePermits = [] } }
            if model.draft.permitAnswer == .yes {
                ForEach(model.draft.residencePermits, id: \.self) { code in
                    let region = model.region(code)
                    SelectionCard(leading: region?.code ?? "EU", title: region?.name ?? TravelCatalog.permitZones.first { $0.id == code }?.label ?? code, actionLabel: model.copy.change, action: { model.replacePermit(code) }, onRemove: { model.removePermit(code) })
                }
                AddRow(title: model.copy.addPermit) { model.picker = .permit }
            }
        }
        .animation(Motion.fade(0.25), value: model.draft.permitAnswer)
        .animation(Motion.fade(0.25), value: model.draft.residencePermits)
    }
}

struct AirportStep: View {
    @Environment(\.designMetrics) private var metrics
    let model: TravelSetupModel

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(18)) {
            SetupQuestion(question: model.copy.airportQuestion, supporting: model.copy.airportBody)
            if let airport = model.airport(model.draft.departureAirport) {
                SelectionCard(leading: airport.code, title: airport.city, detail: "\(airport.name) · \(airport.country)", actionLabel: model.copy.change) { model.picker = .airport }
            } else {
                AddRow(title: model.copy.chooseAirport) { model.picker = .airport }
            }
        }
    }
}

struct CurrencyStep: View {
    @Environment(\.designMetrics) private var metrics
    let model: TravelSetupModel

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(14)) {
            SetupQuestion(question: model.copy.currencyQuestion, supporting: model.copy.currencyBody, privacy: model.copy.currencyPrivacy)
            if let currency = model.currency(model.draft.currency) {
                SelectionCard(leading: currency.code, title: currency.name, actionLabel: model.copy.change) { model.picker = .currency }
            } else {
                if let suggested = model.suggestedCurrency {
                    EyebrowLabel(model.copy.suggestedForYou.uppercased())
                    ChoiceRow(title: suggested.code, detail: suggested.name, isSelected: false) { model.update { $0.currency = suggested.code } }
                }
                AddRow(title: model.copy.chooseCurrency) { model.picker = .currency }
            }
        }
        .animation(Motion.fade(0.25), value: model.draft.currency)
    }
}
