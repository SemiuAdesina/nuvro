import SwiftUI

public struct SearchablePickerSheet<Option: Identifiable & Hashable>: View {
    @Environment(\.dismiss) private var dismiss
    @Environment(\.designMetrics) private var metrics
    @State private var query = ""
    let title: String
    let searchPrompt: String
    let doneLabel: String
    let emptyTitle: String
    let emptyMessage: String
    let options: [Option]
    let label: (Option) -> String
    let detail: (Option) -> String?
    let accessibilityLabel: ((Option) -> String)?
    let code: (Option) -> String?
    let matches: (Option, String) -> Bool
    @Binding var selection: Option.ID?

    public init(title: String, searchPrompt: String, doneLabel: String, emptyTitle: String, emptyMessage: String, options: [Option], selection: Binding<Option.ID?>, label: @escaping (Option) -> String, detail: @escaping (Option) -> String? = { _ in nil }, code: @escaping (Option) -> String? = { _ in nil }, accessibilityLabel: ((Option) -> String)? = nil, matches: @escaping (Option, String) -> Bool) {
        self.title = title
        self.searchPrompt = searchPrompt
        self.doneLabel = doneLabel
        self.emptyTitle = emptyTitle
        self.emptyMessage = emptyMessage
        self.options = options
        self._selection = selection
        self.label = label
        self.detail = detail
        self.accessibilityLabel = accessibilityLabel
        self.code = code
        self.matches = matches
    }

    private var results: [Option] {
        options.filter { matches($0, query) }
    }

    public var body: some View {
        NavigationStack {
            Group {
                if results.isEmpty {
                    NuvroEmptyState(symbol: "magnifyingglass", title: emptyTitle, message: emptyMessage)
                        .background(Palette.background)
                } else {
                    List(results) { option in
                        PickerOptionRow(option: PickerOption(id: "\(option.id)", label: label(option), detail: detail(option), code: code(option)), isSelected: option.id == selection) {
                            selection = option.id
                            dismiss()
                        }
                        .accessibilityLabel(accessibilityLabel?(option) ?? label(option))
                    }
                    .pickerListStyle()
                    .scrollContentBackground(.hidden)
                    .background(Palette.background)
                }
            }
            .navigationTitle(title)
            .pickerSearchable(text: $query, prompt: searchPrompt)
            .pickerToolbar(doneLabel: doneLabel) { dismiss() }
        }
        .presentationDetents([.large])
        .presentationDragIndicator(.visible)
    }
}
