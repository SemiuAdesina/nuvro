import NuvroDesignSystem
import NuvroModels
import SwiftUI

struct SetupPickers: ViewModifier {
    @Bindable var model: TravelSetupModel

    func body(content: Content) -> some View {
        content.sheet(item: $model.picker) { picker in
            switch picker {
            case .residence:
                regionPicker(title: model.copy.residenceQuestion, selection: Binding(get: { model.draft.countryOfResidence }, set: { code in if let code { model.update { $0.countryOfResidence = code } } }))
            case .passport:
                regionPicker(title: model.copy.passportQuestion, selection: Binding(get: { nil }, set: { code in if let code { model.addPassport(code) } }))
            case .permit:
                SearchablePickerSheet(
                    title: model.copy.addPermit, searchPrompt: model.copy.searchCountries, doneLabel: model.copy.done,
                    emptyTitle: model.copy.noMatches, emptyMessage: model.copy.noMatchesBody, options: permitOptions,
                    selection: Binding(get: { nil }, set: { id in if let id { model.addPermit(id) } }),
                    label: \.label, matches: { option, query in query.isEmpty || option.label.localizedCaseInsensitiveContains(query) }
                )
            case .airport:
                SearchablePickerSheet(
                    title: model.copy.airportQuestion, searchPrompt: model.copy.searchAirports, doneLabel: model.copy.done,
                    emptyTitle: model.copy.noMatches, emptyMessage: model.copy.noMatchesBody, options: model.airports,
                    selection: Binding(get: { model.draft.departureAirport }, set: { code in if let code { model.update { $0.departureAirport = code } } }),
                    label: \.city, detail: { "\($0.name) · \($0.country)" }, code: \.code, matches: { $0.matches($1) }
                )
            case .currency:
                SearchablePickerSheet(
                    title: model.copy.currencyQuestion, searchPrompt: model.copy.searchCurrencies, doneLabel: model.copy.done,
                    emptyTitle: model.copy.noMatches, emptyMessage: model.copy.noMatchesBody, options: model.currencies,
                    selection: Binding(get: { model.draft.currency }, set: { code in if let code { model.update { $0.currency = code } } }),
                    label: \.name, code: \.code, accessibilityLabel: { "\($0.name), \($0.code)" }, matches: { TravelCatalog.matches($0, $1) }
                )
            }
        }
    }

    private var permitOptions: [PickerOption] {
        TravelCatalog.permitZones + model.regions.map { PickerOption(id: $0.code, label: $0.name, code: $0.code) }
    }

    private func regionPicker(title: String, selection: Binding<String?>) -> some View {
        SearchablePickerSheet(
            title: title, searchPrompt: model.copy.searchCountries, doneLabel: model.copy.done,
            emptyTitle: model.copy.noMatches, emptyMessage: model.copy.noMatchesBody, options: model.regions,
            selection: selection, label: \.name, code: \.code, accessibilityLabel: \.name, matches: { TravelCatalog.matches($0, $1) }
        )
    }
}

extension View {
    func setupPickers(model: TravelSetupModel) -> some View {
        modifier(SetupPickers(model: model))
    }
}
