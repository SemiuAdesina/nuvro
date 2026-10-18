import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import NuvroTravelSetup
import Observation
import SwiftUI

public struct PreferenceChoices: Sendable {
    public let options: [PickerOption]
    public let allowsMultiple: Bool

    public init(options: [PickerOption], allowsMultiple: Bool = false) {
        self.options = options
        self.allowsMultiple = allowsMultiple
    }
}

@Observable
@MainActor
public final class ProfileModel {
    public private(set) var profile: TravelerProfile
    public var usesTripHistory: Bool
    public var editing: PreferenceRow?
    public private(set) var values: [String: String] = [:]
    public private(set) var lastUpdated: String?
    public var loadFailure: LoadFailure?
    public let choices: [String: PreferenceChoices]
    public let copy: ProfileCopy
    public let stateCopy: StateCopy

    public init(profile: TravelerProfile, choices: [String: PreferenceChoices] = [:], language: Language = .english) {
        self.profile = profile
        self.usesTripHistory = profile.usesTripHistory
        self.choices = choices
        self.copy = .resolve(language)
        self.stateCopy = .resolve(language)
    }

    public var historyStateLabel: String { usesTripHistory ? copy.historyOn : copy.historyOff }

    public func apply(_ profile: TravelerProfile) {
        self.profile = profile
        values = [:]
    }

    public func value(for row: PreferenceRow) -> String {
        values[row.name] ?? row.value
    }

    public func displayValue(for row: PreferenceRow, visible: Int = 2) -> String {
        let parts = parts(of: row)
        guard parts.count > visible else { return parts.joined(separator: ", ") }
        return "\(parts.prefix(visible).joined(separator: ", ")) +\(parts.count - visible)"
    }

    public func canEdit(_ row: PreferenceRow) -> Bool {
        choices[row.name] != nil
    }

    public func selection(for row: PreferenceRow) -> Set<String> {
        Set(parts(of: row))
    }

    private func parts(of row: PreferenceRow) -> [String] {
        value(for: row).components(separatedBy: CharacterSet(charactersIn: ",·")).map { $0.trimmingCharacters(in: .whitespaces) }.filter { !$0.isEmpty }
    }

    public func update(_ row: PreferenceRow, selection: Set<String>) {
        let ordered = (choices[row.name]?.options.map(\.id) ?? []).filter(selection.contains)
        guard !ordered.isEmpty else { return }
        let value = ordered.joined(separator: ", ")
        guard value != self.value(for: row) else { return }
        values[row.name] = value
        lastUpdated = row.name
    }

    public func update(_ row: PreferenceRow, single: String) {
        update(row, selection: [single])
    }
}

public struct ProfileView: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(TransientFeedback.self) private var feedback: TransientFeedback?
    @Bindable private var model: ProfileModel
    private let travelProfile: TravelProfile?
    private let airports: [Airport]
    private let onRegion: () -> Void
    private let onSettings: () -> Void
    private let onEditTravelAccess: (TravelSetupStep) -> Void

    public init(model: ProfileModel, travelProfile: TravelProfile? = nil, airports: [Airport] = [], onRegion: @escaping () -> Void, onSettings: @escaping () -> Void, onEditTravelAccess: @escaping (TravelSetupStep) -> Void = { _ in }) {
        self.model = model
        self.travelProfile = travelProfile
        self.airports = airports
        self.onRegion = onRegion
        self.onSettings = onSettings
        self.onEditTravelAccess = onEditTravelAccess
    }

    @ViewBuilder
    private var travelAccess: some View {
        if let travelProfile {
            VStack(alignment: .leading, spacing: metrics.points(9)) {
                EyebrowLabel(model.copy.travelAccess.uppercased())
                VStack(spacing: 0) {
                    let rows = TravelAccessRow.rows(for: travelProfile, airports: airports)
                    ForEach(rows) { row in
                        NuvroPickerRow(row.title, value: row.value) { onEditTravelAccess(row.step) }
                        if row.id != rows.last?.id { Divider().overlay(Palette.labelAlpha(0.07)) }
                    }
                }
                .surfaceCard()
                Text(model.copy.accessNote).typeStyle(Typography.caption).foregroundStyle(Palette.labelAlpha(0.5))
            }
        }
    }

    private var profile: TravelerProfile { model.profile }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(18)) {
                HStack(spacing: metrics.points(14)) {
                    Text(profile.initial).typeStyle(TypeStyle(size: 21, weight: .bold)).foregroundStyle(Palette.cloud)
                        .frame(width: metrics.points(60), height: metrics.points(60))
                        .background(LinearGradient(colors: [Palette.twilight, Palette.navy], startPoint: .topLeading, endPoint: .bottomTrailing), in: Circle())
                    VStack(alignment: .leading, spacing: metrics.points(2)) {
                        Text(profile.name).typeStyle(TypeStyle(size: 19, weight: .bold, tracking: -0.3)).foregroundStyle(Palette.label)
                        Text(profile.homeLabel).typeStyle(TypeStyle(size: 13.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
                    }
                    Spacer()
                    LinkButton(model.copy.edit) {}
                }
                VStack(alignment: .leading, spacing: metrics.points(12)) {
                    Text(model.copy.travelStyle).typeStyle(TypeStyle(size: 17, weight: .bold, tracking: -0.2)).foregroundStyle(Palette.label)
                    StyleChipRow(chips: profile.styleChips)
                    Text(model.copy.styleNote).typeStyle(TypeStyle(size: 12.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
                }
                .padding(metrics.points(16))
                .surfaceCard()
                VStack(spacing: 0) {
                    ForEach(profile.preferences) { row in
                        Button { if model.canEdit(row) { model.editing = row } } label: {
                            HStack(spacing: metrics.points(12)) {
                                PhotoStandIn(palette: .named(row.placeholder), label: "", labelSize: 6).frame(width: metrics.points(30), height: metrics.points(30)).clipShape(RoundedRectangle(cornerRadius: metrics.points(9), style: .continuous))
                                Text(row.name).typeStyle(TypeStyle(size: 15.5, weight: .medium)).foregroundStyle(Palette.label).lineLimit(1).layoutPriority(1)
                                Spacer(minLength: metrics.points(12))
                                Text(model.displayValue(for: row)).typeStyle(TypeStyle(size: 13.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.55)).lineLimit(1).truncationMode(.tail)
                                ChevronGlyph()
                            }
                            .padding(.horizontal, metrics.points(15)).padding(.vertical, metrics.points(13))
                            .contentShape(Rectangle())
                        }
                        .buttonStyle(.plain)
                        if row.id != profile.preferences.last?.id { Divider().overlay(Palette.labelAlpha(0.07)) }
                    }
                }
                .surfaceCard()
                LearningPreferenceCard(model: model, insights: profile.learned)
                travelAccess
                VStack(spacing: 0) {
                    settingsLink("Language & region", action: onRegion)
                    Divider().overlay(Palette.labelAlpha(0.07))
                    settingsLink("Settings", action: onSettings)
                }
                .surfaceCard()
            }
            .padding(.horizontal, metrics.points(20)).padding(.top, metrics.points(8)).padding(.bottom, metrics.points(40))
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .screenTitle(model.copy.title)
        .sheet(item: $model.editing) { row in
            PreferencePickerSheet(model: model, row: row)
        }
        .onChange(of: model.lastUpdated) { _, updated in
            if updated != nil { feedback?.show(model.copy.updated) }
        }
    }

    private func settingsLink(_ title: String, action: @escaping () -> Void) -> some View {
        Button(action: action) {
            HStack {
                Text(title).typeStyle(TypeStyle(size: 15.5, weight: .medium)).foregroundStyle(Palette.label)
                Spacer()
                ChevronGlyph()
            }
            .padding(.horizontal, metrics.points(15)).padding(.vertical, metrics.points(13))
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
    }
}

struct StyleChipRow: View {
    @Environment(\.designMetrics) private var metrics
    let chips: [StyleChip]

    private func colours(_ tone: StyleChip.Tone) -> (Color, Color) {
        switch tone {
        case .aqua: (Palette.aqua.opacity(0.18), Palette.matchTeal)
        case .gold: (Palette.gold.opacity(0.25), Palette.amber)
        case .twilight: (Palette.twilight.opacity(0.12), Palette.twilight)
        case .sky: (Palette.sky.opacity(0.12), Palette.sky)
        }
    }

    var body: some View {
        FlowLayout(spacing: metrics.points(7), rowSpacing: metrics.points(7)) {
            ForEach(chips) { chip in
                let tone = colours(chip.tone)
                Text(chip.name).typeStyle(TypeStyle(size: 13, weight: .semibold)).foregroundStyle(tone.1)
                    .padding(.horizontal, metrics.points(12)).frame(height: metrics.points(32)).background(tone.0, in: Capsule())
            }
        }
    }
}

struct PreferencePickerSheet: View {
    let model: ProfileModel
    let row: PreferenceRow

    var body: some View {
        if let choices = model.choices[row.name] {
            if choices.allowsMultiple {
                MultiOptionPickerSheet(
                    title: row.name, doneLabel: model.stateCopy.done, options: choices.options,
                    selection: Binding(get: { model.selection(for: row) }, set: { model.update(row, selection: $0) }), minimum: 1
                )
            } else {
                OptionPickerSheet(
                    title: row.name, doneLabel: model.stateCopy.done, options: choices.options,
                    selection: Binding(get: { model.value(for: row) }, set: { model.update(row, single: $0) })
                )
            }
        }
    }
}
