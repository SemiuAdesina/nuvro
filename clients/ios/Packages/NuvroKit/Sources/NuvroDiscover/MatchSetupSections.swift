import NuvroDesignSystem
import SwiftUI

struct DepartureAndTravelersCard: View {
    @Environment(\.designMetrics) private var metrics
    @Bindable var model: MatchSetupModel

    var body: some View {
        VStack(spacing: 0) {
            HStack {
                VStack(alignment: .leading, spacing: metrics.points(2)) {
                    FieldLabel(model.copy.leavingFrom)
                    Text(model.departure).typeStyle(Typography.fieldValue).foregroundStyle(Palette.label)
                }
                Spacer()
                LinkButton(model.copy.change) { model.presentation = .airport }
            }
            .padding(EdgeInsets(top: metrics.points(14), leading: metrics.points(16), bottom: metrics.points(14), trailing: metrics.points(16)))
            Divider().overlay(Palette.labelAlpha(0.08))
            HStack {
                VStack(alignment: .leading, spacing: metrics.points(2)) {
                    FieldLabel(model.copy.travelers)
                    Text(model.travelersLabel).typeStyle(Typography.fieldValue).foregroundStyle(Palette.label)
                        .contentTransition(.numericText())
                        .animation(Motion.fade(0.2), value: model.travelers)
                }
                Spacer()
                LinkButton(model.copy.change) { model.presentation = .travelers }
            }
            .padding(EdgeInsets(top: metrics.points(14), leading: metrics.points(16), bottom: metrics.points(14), trailing: metrics.points(16)))
        }
        .surfaceCard()
    }
}

struct BudgetCard: View {
    @Environment(\.designMetrics) private var metrics
    @Bindable var model: MatchSetupModel

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(10)) {
            HStack(alignment: .firstTextBaseline) {
                FieldLabel(model.copy.totalBudget)
                Spacer()
                Text(model.currency).typeStyle(Typography.caption).foregroundStyle(Palette.labelAlpha(0.5))
            }
            Text(model.budgetLabel)
                .typeStyle(Typography.budgetFigure)
                .foregroundStyle(Palette.label)
                .contentTransition(.numericText())
                .animation(Motion.fade(0.2), value: model.budgetMajor)
            Slider(value: $model.budgetMajor, in: model.budgetRange, step: model.budgetStep)
                .tint(Palette.sky)
                .accessibilityLabel(model.copy.totalBudget)
                .accessibilityValue(model.budgetLabel)
            HStack {
                Text(model.budgetHint).typeStyle(Typography.meta).foregroundStyle(Palette.labelAlpha(0.6))
                Spacer()
                Toggle(isOn: $model.isBudgetFlexible) {
                    Text(model.copy.flexible).typeStyle(Typography.metaEmphasis).foregroundStyle(Palette.label)
                }
                .toggleStyle(.switch)
                .tint(Palette.sky)
                .fixedSize()
            }
        }
        .padding(metrics.points(16))
        .surfaceCard()
    }
}

struct DatesCard: View {
    @Environment(\.designMetrics) private var metrics
    @Bindable var model: MatchSetupModel

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(12)) {
            FieldLabel(model.copy.dates)
            SegmentedChoice(options: DateChoice.allCases, selection: $model.dateChoice) {
                $0 == .exact ? model.copy.exactDates : model.copy.flexibleDates
            }
            Button { model.presentation = .dates } label: {
                HStack {
                    Text(model.datesLabel)
                        .typeStyle(Typography.secondaryAction)
                        .foregroundStyle(Palette.label)
                        .contentTransition(.opacity)
                        .animation(Motion.fade(0.25), value: model.dateSelection)
                    Spacer()
                    ChevronGlyph()
                }
                .contentShape(Rectangle())
            }
            .buttonStyle(.plain)
            .accessibilityHint(model.pickerCopy.dates)
        }
        .padding(metrics.points(16))
        .surfaceCard()
    }
}
