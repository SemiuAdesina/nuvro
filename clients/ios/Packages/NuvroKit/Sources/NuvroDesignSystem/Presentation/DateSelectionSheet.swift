import NuvroModels
import SwiftUI

public struct DateSelectionCopy: Sendable {
    public let title: String
    public let exact: String
    public let flexible: String
    public let start: String
    public let end: String
    public let month: String
    public let nights: String
    public let done: String

    public init(title: String, exact: String, flexible: String, start: String, end: String, month: String, nights: String, done: String) {
        self.title = title
        self.exact = exact
        self.flexible = flexible
        self.start = start
        self.end = end
        self.month = month
        self.nights = nights
        self.done = done
    }
}

public struct DateSelectionSheet: View {
    private enum Mode: String, CaseIterable { case exact, flexible }

    @Environment(\.dismiss) private var dismiss
    @Environment(\.designMetrics) private var metrics
    @State private var mode: Mode
    @State private var start: Date
    @State private var end: Date
    @State private var month: Date
    @State private var nights: Int
    @Binding var selection: DateSelection
    let copy: DateSelectionCopy
    let calendar: Calendar
    let today: Date

    public init(selection: Binding<DateSelection>, copy: DateSelectionCopy, calendar: Calendar = .current, today: Date = .now) {
        self._selection = selection
        self.copy = copy
        self.calendar = calendar
        self.today = today
        switch selection.wrappedValue {
        case .exact(let start, let end):
            _mode = State(initialValue: .exact); _start = State(initialValue: start); _end = State(initialValue: end)
            _month = State(initialValue: start); _nights = State(initialValue: 7)
        case .flexible(let month, let nights):
            _mode = State(initialValue: .flexible); _month = State(initialValue: month); _nights = State(initialValue: nights)
            _start = State(initialValue: today); _end = State(initialValue: calendar.date(byAdding: .day, value: nights, to: today) ?? today)
        }
    }

    private var months: [Date] {
        (0..<12).compactMap { calendar.date(byAdding: .month, value: $0, to: calendar.date(from: calendar.dateComponents([.year, .month], from: today)) ?? today) }
    }

    public var body: some View {
        NavigationStack {
            VStack(alignment: .leading, spacing: metrics.points(16)) {
                SegmentedChoice(options: Mode.allCases, selection: $mode, isPill: true) { $0 == .exact ? copy.exact : copy.flexible }
                if mode == .exact {
                    VStack(spacing: 0) {
                        DatePicker(copy.start, selection: $start, in: today..., displayedComponents: .date)
                            .padding(.horizontal, metrics.points(16)).padding(.vertical, metrics.points(10))
                        Divider().overlay(Palette.labelAlpha(0.07))
                        DatePicker(copy.end, selection: $end, in: (calendar.date(byAdding: .day, value: 1, to: start) ?? start)..., displayedComponents: .date)
                            .padding(.horizontal, metrics.points(16)).padding(.vertical, metrics.points(10))
                    }
                    .typeStyle(Typography.listRow)
                    .tint(Palette.sky)
                    .surfaceCard()
                } else {
                    VStack(spacing: 0) {
                        Picker(copy.month, selection: $month) {
                            ForEach(months, id: \.self) { Text($0.formatted(.dateTime.month(.wide).year())).tag($0) }
                        }
                        .pickerStyle(.menu)
                        .tint(Palette.sky)
                        .padding(.horizontal, metrics.points(16)).padding(.vertical, metrics.points(10))
                        Divider().overlay(Palette.labelAlpha(0.07))
                        HStack {
                            Text(copy.nights).typeStyle(Typography.listRow).foregroundStyle(Palette.label)
                            Spacer()
                            CountStepper(value: $nights, in: 1...30, unit: copy.nights)
                        }
                        .padding(.horizontal, metrics.points(16)).padding(.vertical, metrics.points(10))
                    }
                    .surfaceCard()
                }
            }
            .padding(metrics.points(20))
            .frame(maxHeight: .infinity, alignment: .top)
            .background(Palette.background)
            .navigationTitle(copy.title)
            .pickerToolbar(doneLabel: copy.done) { commit(); dismiss() }
            .onChange(of: start) { _, value in if end <= value { end = calendar.date(byAdding: .day, value: 1, to: value) ?? value } }
        }
        .presentationDetents([.medium, .large])
        .presentationDragIndicator(.visible)
    }

    private func commit() {
        selection = mode == .exact ? DateSelection.exact(start: start, end: end).normalised() : .flexible(month: month, nights: nights)
    }
}
