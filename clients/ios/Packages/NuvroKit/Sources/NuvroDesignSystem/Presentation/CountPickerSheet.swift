import SwiftUI

public struct CountPickerLine: Identifiable, Sendable {
    public let id: String
    public let title: String
    public let detail: String
    public let range: ClosedRange<Int>

    public init(id: String, title: String, detail: String, range: ClosedRange<Int>) {
        self.id = id
        self.title = title
        self.detail = detail
        self.range = range
    }
}

public struct CountPickerSheet: View {
    @Environment(\.dismiss) private var dismiss
    @Environment(\.designMetrics) private var metrics
    let title: String
    let doneLabel: String
    let lines: [CountPickerLine]
    @Binding var counts: [String: Int]

    public init(title: String, doneLabel: String, lines: [CountPickerLine], counts: Binding<[String: Int]>) {
        self.title = title
        self.doneLabel = doneLabel
        self.lines = lines
        self._counts = counts
    }

    public var body: some View {
        NavigationStack {
            VStack(spacing: 0) {
                ForEach(lines) { line in
                    HStack(spacing: metrics.points(12)) {
                        VStack(alignment: .leading, spacing: 2) {
                            Text(line.title).typeStyle(Typography.listRow).foregroundStyle(Palette.label)
                            Text(line.detail).typeStyle(Typography.caption).foregroundStyle(Palette.labelAlpha(0.55))
                        }
                        Spacer(minLength: 0)
                        CountStepper(value: binding(for: line), in: line.range, unit: line.title)
                    }
                    .padding(.horizontal, metrics.points(16)).padding(.vertical, metrics.points(12))
                    if line.id != lines.last?.id { Divider().overlay(Palette.labelAlpha(0.07)) }
                }
            }
            .surfaceCard()
            .padding(metrics.points(20))
            .frame(maxHeight: .infinity, alignment: .top)
            .background(Palette.background)
            .navigationTitle(title)
            .pickerToolbar(doneLabel: doneLabel) { dismiss() }
        }
        .presentationDetents([.medium])
        .presentationDragIndicator(.visible)
    }

    private func binding(for line: CountPickerLine) -> Binding<Int> {
        Binding(
            get: { counts[line.id] ?? line.range.lowerBound },
            set: { counts[line.id] = min(max($0, line.range.lowerBound), line.range.upperBound) }
        )
    }
}
