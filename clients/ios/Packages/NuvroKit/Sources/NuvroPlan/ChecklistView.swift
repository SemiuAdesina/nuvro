import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import Observation
import SwiftUI

@Observable
@MainActor
public final class ChecklistModel {
    public private(set) var groups: [ChecklistGroup]
    public let insight: String
    public let copy: ChecklistCopy

    public init(groups: [ChecklistGroup], insight: String, language: Language = .english) {
        self.groups = groups
        self.insight = insight
        self.copy = .resolve(language)
    }

    public var totalCount: Int { groups.reduce(0) { $0 + $1.items.count } }
    public var doneCount: Int { groups.reduce(0) { $0 + $1.doneCount } }
    public var fraction: Double { totalCount == 0 ? 0 : Double(doneCount) / Double(totalCount) }
    public var readyLabel: String { "\(doneCount) \(copy.of) \(totalCount) \(copy.ready)" }

    public func countLabel(_ group: ChecklistGroup) -> String { "\(group.doneCount) \(copy.of) \(group.items.count)" }

    public var addingTo: ChecklistGroup?
    public var newItemName = ""

    public var isAllDone: Bool { totalCount > 0 && doneCount == totalCount }

    public func isComplete(_ group: ChecklistGroup) -> Bool { !group.items.isEmpty && group.doneCount == group.items.count }

    public func toggle(_ item: ChecklistItem, in group: ChecklistGroup) {
        groups = groups.map { current in
            guard current.id == group.id else { return current }
            return ChecklistGroup(name: current.name, items: current.items.map { $0.id == item.id ? $0.toggled() : $0 })
        }
    }

    public func beginAdding(to group: ChecklistGroup) {
        addingTo = group
        newItemName = ""
    }

    public func commitNewItem() {
        let name = newItemName.trimmingCharacters(in: .whitespaces)
        guard let group = addingTo, !name.isEmpty else { addingTo = nil; return }
        groups = groups.map { current in
            guard current.id == group.id, !current.items.contains(where: { $0.name.caseInsensitiveCompare(name) == .orderedSame }) else { return current }
            return ChecklistGroup(name: current.name, items: current.items + [ChecklistItem(name: name)])
        }
        addingTo = nil
        newItemName = ""
    }

    public func remove(_ item: ChecklistItem, from group: ChecklistGroup) {
        groups = groups.map { current in
            guard current.id == group.id else { return current }
            return ChecklistGroup(name: current.name, items: current.items.filter { $0.id != item.id })
        }
    }
}

public struct ChecklistView: View {
    @Environment(\.designMetrics) private var metrics
    @FocusState private var isAddingFocused: Bool
    @Bindable private var model: ChecklistModel
    private let subtitle: String

    public init(model: ChecklistModel, subtitle: String) {
        self.model = model
        self.subtitle = subtitle
    }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(16)) {
                VStack(alignment: .leading, spacing: metrics.points(10)) {
                    HStack(alignment: .firstTextBaseline) {
                        Text(model.readyLabel).typeStyle(TypeStyle(size: 17, weight: .bold, tracking: -0.2)).foregroundStyle(Palette.label)
                        Spacer()
                        Text(model.isAllDone ? model.copy.allDone : model.copy.noRush).typeStyle(TypeStyle(size: 13, weight: .regular)).foregroundStyle(model.isAllDone ? Palette.matchTeal : Palette.labelAlpha(0.55))
                    }
                    GradientProgressBar(fraction: model.fraction, height: 6)
                }
                .padding(metrics.points(16))
                .surfaceCard()
                InsightCard(title: model.copy.basedOnTrip) {
                    Text(model.insight).typeStyle(TypeStyle(size: 14, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.75))
                }
                ForEach(model.groups) { group in
                    VStack(alignment: .leading, spacing: metrics.points(10)) {
                        HStack(alignment: .firstTextBaseline, spacing: metrics.points(8)) {
                            Text(group.name).typeStyle(TypeStyle(size: 17, weight: .bold, tracking: -0.2)).foregroundStyle(Palette.label)
                            Text(model.countLabel(group)).typeStyle(TypeStyle(size: 12.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.5))
                            if model.isComplete(group) { TagChip(model.copy.sectionDone, tone: .aqua) }
                        }
                        VStack(spacing: 0) {
                            ForEach(group.items) { item in
                                ChecklistRow(item: item) { model.toggle(item, in: group) }
                                    .contextMenu { Button(model.copy.remove, systemImage: "trash", role: .destructive) { model.remove(item, from: group) } }
                                Divider().overlay(Palette.labelAlpha(0.07))
                            }
                            if model.addingTo?.id == group.id {
                                HStack(spacing: metrics.points(12)) {
                                    Circle().strokeBorder(Palette.labelAlpha(0.25), lineWidth: 1.5).frame(width: metrics.points(22), height: metrics.points(22))
                                    TextField(model.copy.itemPlaceholder, text: $model.newItemName)
                                        .typeStyle(TypeStyle(size: 15, weight: .medium))
                                        .foregroundStyle(Palette.label)
                                        .submitLabel(.done)
                                        .focused($isAddingFocused)
                                        .onSubmit(model.commitNewItem)
                                        .accessibilityLabel(model.copy.addItem)
                                }
                                .padding(.horizontal, metrics.points(14)).padding(.vertical, metrics.points(12))
                                .onAppear { isAddingFocused = true }
                            } else {
                                Button { model.beginAdding(to: group) } label: {
                                    HStack(spacing: metrics.points(12)) {
                                        Text("+").typeStyle(TypeStyle(size: 14, weight: .semibold)).foregroundStyle(Palette.sky)
                                            .frame(width: metrics.points(22), height: metrics.points(22)).background(Palette.sky.opacity(0.1), in: Circle())
                                        Text(model.copy.addItem).typeStyle(TypeStyle(size: 15, weight: .semibold)).foregroundStyle(Palette.sky)
                                        Spacer()
                                    }
                                    .padding(.horizontal, metrics.points(14)).padding(.vertical, metrics.points(12))
                                    .contentShape(Rectangle())
                                }
                                .buttonStyle(.plain)
                            }
                        }
                        .surfaceCard()
                    }
                }
            }
            .padding(.horizontal, metrics.points(20)).padding(.top, metrics.points(8)).padding(.bottom, metrics.points(40))
            .animation(Motion.fade(0.25), value: model.groups)
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .scrollDismissesKeyboard(.interactively)
        .onChange(of: isAddingFocused) { _, focused in if !focused, model.addingTo != nil { model.commitNewItem() } }
        .screenTitle(model.copy.title, subtitle: subtitle)
    }
}

struct ChecklistRow: View {
    @Environment(\.designMetrics) private var metrics
    let item: ChecklistItem
    let toggle: () -> Void

    var body: some View {
        Button(action: toggle) {
            HStack(spacing: metrics.points(12)) {
                ZStack {
                    Circle().fill(item.isDone ? Palette.aqua : .clear)
                    if !item.isDone { Circle().strokeBorder(Palette.labelAlpha(0.25), lineWidth: 1.5) }
                    if item.isDone { Image(systemName: "checkmark").font(.system(size: metrics.points(11), weight: .bold)).foregroundStyle(.white) }
                }
                .frame(width: metrics.points(22), height: metrics.points(22))
                VStack(alignment: .leading, spacing: 1) {
                    Text(item.name).typeStyle(TypeStyle(size: 15, weight: .medium)).foregroundStyle(item.isDone ? Palette.labelAlpha(0.5) : Palette.label).strikethrough(item.isDone)
                    if let note = item.note, !note.isEmpty { Text(note).typeStyle(TypeStyle(size: 11.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.5)) }
                }
                Spacer(minLength: 0)
            }
            .padding(.horizontal, metrics.points(14)).padding(.vertical, metrics.points(12))
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .accessibilityAddTraits(item.isDone ? .isSelected : [])
    }
}
