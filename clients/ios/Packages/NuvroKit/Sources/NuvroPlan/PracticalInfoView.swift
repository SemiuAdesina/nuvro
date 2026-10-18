import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import Observation
import SwiftUI

@Observable
@MainActor
public final class PracticalModel {
    public let entryCheckedLabel: String
    public let entryRows: [EntryRequirementRow]
    public let sections: [InfoSection]
    public let emergency: [EmergencyContact]
    public var openSection: String?
    public let copy: PracticalCopy

    public init(entryCheckedLabel: String, entryRows: [EntryRequirementRow], sections: [InfoSection], emergency: [EmergencyContact], openSection: String? = nil, language: Language = .english) {
        self.entryCheckedLabel = entryCheckedLabel
        self.entryRows = entryRows
        self.sections = sections
        self.emergency = emergency
        self.openSection = openSection
        self.copy = .resolve(language)
    }

    public func toggle(_ section: InfoSection) { openSection = openSection == section.id ? nil : section.id }
    public func isOpen(_ section: InfoSection) -> Bool { openSection == section.id }
}

public struct PracticalInfoView: View {
    @Environment(\.designMetrics) private var metrics
    private let model: PracticalModel
    private let subtitle: String

    public init(model: PracticalModel, subtitle: String) {
        self.model = model
        self.subtitle = subtitle
    }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(16)) {
                entryCard
                ForEach(model.sections) { section in InfoSectionCard(section: section, model: model) }
                VStack(alignment: .leading, spacing: metrics.points(11)) {
                    Text(model.copy.emergency).typeStyle(TypeStyle(size: 17, weight: .bold, tracking: -0.2)).foregroundStyle(Palette.label)
                    HStack(spacing: metrics.points(8)) {
                        ForEach(model.emergency) { contact in FactTile(key: contact.key, value: contact.value) }
                    }
                    Text(model.copy.emergencyNote).typeStyle(TypeStyle(size: 12, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.55))
                }
                .padding(metrics.points(16))
                .surfaceCard()
            }
            .padding(.horizontal, metrics.points(20)).padding(.top, metrics.points(8)).padding(.bottom, metrics.points(40))
            .animation(Motion.settle(0.35), value: model.openSection)
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .screenTitle(model.copy.title, subtitle: subtitle)
    }

    private var entryCard: some View {
        VStack(alignment: .leading, spacing: metrics.points(12)) {
            HStack {
                Text(model.copy.entry).typeStyle(TypeStyle(size: 18, weight: .bold, tracking: -0.2)).foregroundStyle(Palette.label)
                Spacer()
                Text(model.entryCheckedLabel).typeStyle(TypeStyle(size: 11, weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.5))
            }
            VStack(alignment: .leading, spacing: metrics.points(9)) {
                ForEach(model.entryRows) { row in
                    HStack(alignment: .top, spacing: metrics.points(10)) {
                        Text(row.state == .met ? "✓" : "i").typeStyle(TypeStyle(size: 11, weight: .bold)).foregroundStyle(row.state == .met ? Palette.matchTeal : Palette.sky)
                            .frame(width: metrics.points(20), height: metrics.points(20))
                            .background(row.state == .met ? Palette.aqua.opacity(0.2) : Palette.sky.opacity(0.14), in: Circle())
                            .padding(.top, 1)
                        VStack(alignment: .leading, spacing: 1) {
                            Text(row.key).typeStyle(TypeStyle(size: 14, weight: .semibold)).foregroundStyle(Palette.label)
                            Text(row.value).typeStyle(TypeStyle(size: 12.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.65))
                        }
                    }
                }
            }
            VStack(alignment: .leading, spacing: metrics.points(6)) {
                Text(model.copy.passportNote).typeStyle(TypeStyle(size: 12.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.78))
                LinkButton(model.copy.openSource) {}
            }
            .padding(.horizontal, metrics.points(12)).padding(.vertical, metrics.points(11))
            .background(Palette.gold.opacity(0.16), in: RoundedRectangle(cornerRadius: metrics.points(14), style: .continuous))
        }
        .padding(metrics.points(16))
        .background(Palette.surface, in: RoundedRectangle(cornerRadius: metrics.points(Radius.card), style: .continuous))
        .overlay(RoundedRectangle(cornerRadius: metrics.points(Radius.card), style: .continuous).strokeBorder(Palette.sky.opacity(0.22), lineWidth: 0.5))
    }
}

struct InfoSectionCard: View {
    @Environment(\.designMetrics) private var metrics
    let section: InfoSection
    let model: PracticalModel

    var body: some View {
        VStack(spacing: 0) {
            Button { model.toggle(section) } label: {
                HStack(spacing: metrics.points(12)) {
                    PhotoStandIn(palette: .named(section.placeholder), label: "", labelSize: 6).frame(width: metrics.points(32), height: metrics.points(32)).clipShape(RoundedRectangle(cornerRadius: metrics.points(10), style: .continuous))
                    VStack(alignment: .leading, spacing: metrics.points(2)) {
                        Text(section.name).typeStyle(TypeStyle(size: 16, weight: .semibold)).foregroundStyle(Palette.label)
                        Text(section.gist).typeStyle(TypeStyle(size: 12.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
                    }
                    Spacer(minLength: 0)
                    Image(systemName: "chevron.down").font(.system(size: metrics.points(12), weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.35)).rotationEffect(.degrees(model.isOpen(section) ? 180 : 0))
                }
                .padding(EdgeInsets(top: metrics.points(14), leading: metrics.points(15), bottom: metrics.points(14), trailing: metrics.points(15)))
                .contentShape(Rectangle())
            }
            .buttonStyle(.plain)
            .accessibilityValue(model.isOpen(section) ? "Expanded" : "Collapsed")
            if model.isOpen(section) {
                VStack(alignment: .leading, spacing: metrics.points(8)) {
                    Divider().overlay(Palette.labelAlpha(0.07))
                    ForEach(section.rows) { row in
                        HStack(alignment: .firstTextBaseline) {
                            Text(row.key).typeStyle(TypeStyle(size: 13, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.62))
                            Spacer()
                            Text(row.value).typeStyle(TypeStyle(size: 13, weight: .semibold)).foregroundStyle(Palette.label).multilineTextAlignment(.trailing)
                        }
                    }
                    Text(section.source).typeStyle(TypeStyle(size: 11, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.45)).padding(.top, metrics.points(4))
                }
                .padding(EdgeInsets(top: 0, leading: metrics.points(59), bottom: metrics.points(14), trailing: metrics.points(15)))
                .transition(.opacity)
            }
        }
        .surfaceCard()
    }
}
