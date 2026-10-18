import Foundation
import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import Observation
import SwiftUI

public enum RegionPresentation: String, Identifiable, Sendable {
    case appLanguage
    case spokenLanguages
    case region
    case currency
    case homeAirport

    public var id: String { rawValue }
}

public struct RegionChoices: Sendable {
    public let regions: [PickerOption]
    public let currencies: [PickerOption]
    public let airports: [Airport]

    public init(regions: [PickerOption], currencies: [PickerOption], airports: [Airport]) {
        self.regions = regions
        self.currencies = currencies
        self.airports = airports
    }
}

@Observable
@MainActor
public final class RegionModel {
    public let settings: RegionSettings
    public var units: UnitPreferences
    public var appLanguage: Language
    public var spokenLanguages: Set<String>
    public var region: String
    public var currency: String
    public var homeAirportCode: String?
    public var presentation: RegionPresentation?
    public let choices: RegionChoices
    public let copy: RegionCopy
    public let pickerCopy: PickerCopy
    public let stateCopy: StateCopy
    private let sampleDate: Date
    private let calendar: Calendar

    public init(settings: RegionSettings, units: UnitPreferences, sampleDate: Date, choices: RegionChoices = RegionChoices(regions: [], currencies: [], airports: []), language: Language = .english, calendar: Calendar = .current) {
        self.settings = settings
        self.units = units
        self.sampleDate = sampleDate
        self.choices = choices
        self.copy = .resolve(language)
        self.pickerCopy = .resolve(language)
        self.stateCopy = .resolve(language)
        self.calendar = calendar
        self.appLanguage = Language.allCases.first { $0.nativeName == settings.appLanguage } ?? language
        self.spokenLanguages = Set(settings.spokenLanguages.split(separator: ",").map { $0.trimmingCharacters(in: .whitespaces) })
        self.region = settings.region
        self.currency = settings.currency
        self.homeAirportCode = choices.airports.first { settings.homeAirport.contains($0.code) }?.code
    }

    public var appLanguageLabel: String { appLanguage.nativeName }
    public var spokenLanguagesLabel: String { spokenLanguages.sorted().joined(separator: ", ") }
    public var homeAirportLabel: String { choices.airports.first { $0.code == homeAirportCode }?.shortLabel ?? settings.homeAirport }
    public var languageOptions: [PickerOption] { Language.allCases.map { PickerOption(id: $0.nativeName, label: $0.nativeName) } }

    public var preview: [String] {
        [
            units.temperature == .fahrenheit ? "72°F" : "22°C",
            units.distance == .miles ? "3.2 mi" : "5.1 km",
            "$2,780",
            dateLabel(units.date),
            units.clock == .twelveHour ? "7:30 PM" : "19:30"
        ]
    }

    public func dateLabel(_ style: DateStyle) -> String {
        var formatter = Date.FormatStyle(date: .abbreviated, time: .omitted, locale: Locale(identifier: style == .monthFirst ? "en_US" : "en_GB"), calendar: calendar, timeZone: calendar.timeZone)
        formatter.locale = Locale(identifier: style == .monthFirst ? "en_US" : "en_GB")
        return sampleDate.formatted(formatter)
    }

    public func label(_ unit: DistanceUnit) -> String { unit == .miles ? "Miles" : "Kilometres" }
    public func label(_ unit: TemperatureUnit) -> String { unit == .fahrenheit ? "Fahrenheit" : "Celsius" }
    public func label(_ format: ClockFormat) -> String { format == .twelveHour ? "12-hour" : "24-hour" }
}

public struct RegionView: View {
    @Environment(\.designMetrics) private var metrics
    @Bindable private var model: RegionModel

    public init(model: RegionModel) {
        self.model = model
    }

    private var copy: RegionCopy { model.copy }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(16)) {
                VStack(spacing: 0) {
                    settingRow(copy.appLanguage, note: copy.appLanguageNote, value: model.appLanguageLabel) { model.presentation = .appLanguage }
                    Divider().overlay(Palette.labelAlpha(0.07))
                    settingRow(copy.spoken, note: copy.spokenNote, value: model.spokenLanguagesLabel) { model.presentation = .spokenLanguages }
                }
                .surfaceCard()
                VStack(spacing: 0) {
                    settingRow(copy.region, note: model.settings.regionNote, value: model.region) { model.presentation = .region }
                    Divider().overlay(Palette.labelAlpha(0.07))
                    settingRow(copy.currency, note: nil, value: model.currency) { model.presentation = .currency }
                    Divider().overlay(Palette.labelAlpha(0.07))
                    settingRow(copy.homeAirport, note: nil, value: model.homeAirportLabel) { model.presentation = .homeAirport }
                }
                .surfaceCard()
                VStack(alignment: .leading, spacing: metrics.points(11)) {
                    Text(copy.units).typeStyle(TypeStyle(size: 17, weight: .bold, tracking: -0.2)).foregroundStyle(Palette.label)
                    VStack(alignment: .leading, spacing: metrics.points(14)) {
                        unitRow(copy.distance) { SegmentedChoice(options: DistanceUnit.allCases, selection: $model.units.distance, compact: true) { model.label($0) } }
                        unitRow(copy.temperature) { SegmentedChoice(options: TemperatureUnit.allCases, selection: $model.units.temperature, compact: true) { model.label($0) } }
                        unitRow(copy.time) { SegmentedChoice(options: ClockFormat.allCases, selection: $model.units.clock, compact: true) { model.label($0) } }
                        unitRow(copy.date) { SegmentedChoice(options: DateStyle.allCases, selection: $model.units.date, compact: true) { model.dateLabel($0) } }
                    }
                    .padding(metrics.points(15))
                    .surfaceCard()
                }
                InsightCard(title: copy.preview, showsMark: false) {
                    FlowLayout(spacing: metrics.points(7), rowSpacing: metrics.points(7)) {
                        ForEach(model.preview, id: \.self) { value in
                            Text(value).typeStyle(TypeStyle(size: 14, weight: .semibold)).foregroundStyle(Palette.label)
                                .padding(.horizontal, metrics.points(13)).frame(height: metrics.points(34))
                                .background(Palette.surface.opacity(0.9), in: Capsule())
                                .overlay(Capsule().strokeBorder(Palette.labelAlpha(0.1), lineWidth: 0.5))
                        }
                    }
                }
            }
            .padding(.horizontal, metrics.points(20)).padding(.top, metrics.points(8)).padding(.bottom, metrics.points(40))
            .animation(Motion.fade(0.25), value: model.units)
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .screenTitle(copy.title)
        .sheet(item: $model.presentation) { presentation in
            RegionPickerSheet(model: model, presentation: presentation)
        }
    }

    private func settingRow(_ title: String, note: String?, value: String, action: @escaping () -> Void) -> some View {
        Button(action: action) {
            HStack(spacing: metrics.points(12)) {
                VStack(alignment: .leading, spacing: metrics.points(2)) {
                    Text(title).typeStyle(TypeStyle(size: 15.5, weight: .medium)).foregroundStyle(Palette.label)
                    if let note, !note.isEmpty { Text(note).typeStyle(TypeStyle(size: 12, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.55)) }
                }
                Spacer()
                Text(value).typeStyle(TypeStyle(size: 13.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
                ChevronGlyph()
            }
            .padding(.horizontal, metrics.points(15)).padding(.vertical, metrics.points(13))
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
    }

    private func unitRow<Content: View>(_ label: String, @ViewBuilder content: () -> Content) -> some View {
        VStack(alignment: .leading, spacing: metrics.points(8)) {
            Text(label).typeStyle(TypeStyle(size: 12, weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.55))
            content()
        }
    }
}

struct RegionPickerSheet: View {
    @Bindable var model: RegionModel
    let presentation: RegionPresentation

    var body: some View {
        switch presentation {
        case .appLanguage:
            SearchablePickerSheet(
                title: model.copy.appLanguage, searchPrompt: model.pickerCopy.searchLanguages, doneLabel: model.stateCopy.done,
                emptyTitle: model.pickerCopy.noLanguages, emptyMessage: model.pickerCopy.noMatchesBody, options: Language.allCases,
                selection: Binding(get: { model.appLanguage.id }, set: { id in if let language = Language(rawValue: id ?? "") { model.appLanguage = language } }),
                label: \.nativeName, matches: { language, query in Language.matching(query).contains(language) }
            )
        case .spokenLanguages:
            MultiOptionPickerSheet(title: model.copy.spoken, doneLabel: model.stateCopy.done, options: model.languageOptions, selection: $model.spokenLanguages, minimum: 1)
        case .region:
            SearchablePickerSheet(
                title: model.copy.region, searchPrompt: model.pickerCopy.searchRegions, doneLabel: model.stateCopy.done,
                emptyTitle: model.pickerCopy.noMatches, emptyMessage: model.pickerCopy.noMatchesBody, options: model.choices.regions,
                selection: Binding(get: { model.region }, set: { if let value = $0 { model.region = value } }),
                label: \.label, detail: \.detail, matches: { option, query in query.isEmpty || option.label.localizedCaseInsensitiveContains(query) }
            )
        case .currency:
            SearchablePickerSheet(
                title: model.copy.currency, searchPrompt: model.pickerCopy.searchCurrencies, doneLabel: model.stateCopy.done,
                emptyTitle: model.pickerCopy.noMatches, emptyMessage: model.pickerCopy.noMatchesBody, options: model.choices.currencies,
                selection: Binding(get: { model.currency }, set: { if let value = $0 { model.currency = value } }),
                label: \.label, detail: \.detail, matches: { option, query in query.isEmpty || option.label.localizedCaseInsensitiveContains(query) || (option.detail ?? "").localizedCaseInsensitiveContains(query) }
            )
        case .homeAirport:
            SearchablePickerSheet(
                title: model.copy.homeAirport, searchPrompt: model.pickerCopy.searchAirports, doneLabel: model.stateCopy.done,
                emptyTitle: model.pickerCopy.noAirports, emptyMessage: model.pickerCopy.noAirportsBody, options: model.choices.airports,
                selection: $model.homeAirportCode, label: \.shortLabel, detail: { "\($0.name) · \($0.country)" }, matches: { $0.matches($1) }
            )
        }
    }
}
