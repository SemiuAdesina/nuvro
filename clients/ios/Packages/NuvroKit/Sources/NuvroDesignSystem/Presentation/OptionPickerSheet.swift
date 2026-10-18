import SwiftUI

public struct PickerOption: Identifiable, Hashable, Sendable {
    public let id: String
    public let label: String
    public let detail: String?
    public let code: String?

    public init(id: String, label: String, detail: String? = nil, code: String? = nil) {
        self.id = id
        self.label = label
        self.detail = detail
        self.code = code
    }

    public init(_ label: String, detail: String? = nil) {
        self.init(id: label, label: label, detail: detail)
    }
}

public struct OptionPickerSheet: View {
    @Environment(\.dismiss) private var dismiss
    @Environment(\.designMetrics) private var metrics
    let title: String
    let doneLabel: String
    let options: [PickerOption]
    @Binding var selection: String
    let dismissesOnSelect: Bool

    public init(title: String, doneLabel: String, options: [PickerOption], selection: Binding<String>, dismissesOnSelect: Bool = true) {
        self.title = title
        self.doneLabel = doneLabel
        self.options = options
        self._selection = selection
        self.dismissesOnSelect = dismissesOnSelect
    }

    public var body: some View {
        NavigationStack {
            List(options) { option in
                PickerOptionRow(option: option, isSelected: option.id == selection) {
                    selection = option.id
                    if dismissesOnSelect { dismiss() }
                }
            }
            .pickerListStyle()
            .scrollContentBackground(.hidden)
            .background(Palette.background)
            .navigationTitle(title)
            .pickerToolbar(doneLabel: doneLabel) { dismiss() }
        }
        .presentationDetents(options.count > 6 ? [.large] : [.medium, .large])
        .presentationDragIndicator(.visible)
    }
}

public struct MultiOptionPickerSheet: View {
    @Environment(\.dismiss) private var dismiss
    let title: String
    let doneLabel: String
    let options: [PickerOption]
    @Binding var selection: Set<String>
    let minimum: Int

    public init(title: String, doneLabel: String, options: [PickerOption], selection: Binding<Set<String>>, minimum: Int = 0) {
        self.title = title
        self.doneLabel = doneLabel
        self.options = options
        self._selection = selection
        self.minimum = minimum
    }

    public var body: some View {
        NavigationStack {
            List(options) { option in
                let isSelected = selection.contains(option.id)
                PickerOptionRow(option: option, isSelected: isSelected) {
                    if isSelected {
                        guard selection.count > minimum else { return }
                        selection.remove(option.id)
                    } else {
                        selection.insert(option.id)
                    }
                }
            }
            .pickerListStyle()
            .scrollContentBackground(.hidden)
            .background(Palette.background)
            .navigationTitle(title)
            .pickerToolbar(doneLabel: doneLabel) { dismiss() }
        }
        .presentationDetents([.medium, .large])
        .presentationDragIndicator(.visible)
    }
}

struct PickerOptionRow: View {
    @Environment(\.designMetrics) private var metrics
    let option: PickerOption
    let isSelected: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            HStack(spacing: metrics.points(12)) {
                if let code = option.code {
                    CodeTile(code)
                }
                VStack(alignment: .leading, spacing: 2) {
                    Text(option.label).typeStyle(Typography.listRow).foregroundStyle(Palette.label)
                    if let detail = option.detail { Text(detail).typeStyle(Typography.caption).foregroundStyle(Palette.labelAlpha(0.55)) }
                }
                Spacer(minLength: 0)
                Image(systemName: "checkmark")
                    .font(.system(size: metrics.points(14), weight: .bold))
                    .foregroundStyle(Palette.sky)
                    .opacity(isSelected ? 1 : 0)
                    .accessibilityHidden(true)
            }
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .listRowBackground(Palette.surface)
        .accessibilityAddTraits(isSelected ? .isSelected : [])
    }
}

extension View {
    func pickerToolbar(doneLabel: String, done: @escaping () -> Void) -> some View {
        toolbar {
            ToolbarItem(placement: .confirmationAction) {
                Button(doneLabel, action: done).typeStyle(Typography.secondaryAction).foregroundStyle(Palette.sky)
            }
        }
        .toolbarTitleDisplayMode(.inline)
    }
}

extension View {
    func pickerListStyle() -> some View {
        #if os(iOS)
        listStyle(.insetGrouped)
        #else
        listStyle(.inset)
        #endif
    }

    func pickerSearchable(text: Binding<String>, prompt: String) -> some View {
        #if os(iOS)
        searchable(text: text, placement: .navigationBarDrawer(displayMode: .always), prompt: prompt)
        #else
        searchable(text: text, prompt: prompt)
        #endif
    }
}

public struct CodeTile: View {
    @Environment(\.designMetrics) private var metrics
    let code: String

    public init(_ code: String) {
        self.code = code
    }

    public var body: some View {
        Text(code)
            .font(.system(size: metrics.points(code.count > 2 ? 11 : 13), weight: .bold, design: .rounded))
            .tracking(0.4)
            .foregroundStyle(Palette.twilightText)
            .frame(width: metrics.points(40), height: metrics.points(30))
            .background(Palette.labelAlpha(0.07), in: RoundedRectangle(cornerRadius: metrics.points(8), style: .continuous))
            .accessibilityHidden(true)
    }
}
