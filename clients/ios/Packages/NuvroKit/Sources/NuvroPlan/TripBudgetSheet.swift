import NuvroDesignSystem
import NuvroModels
import SwiftUI

struct TripBudgetSheet: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(\.dismiss) private var dismiss
    @Bindable var model: TripPlanModel

    private var label: String {
        (try? Money(amountMinor: Int64(model.budgetMajor * 100), currency: model.plan.budgetCeiling.currency))?.formattedWhole(locale: model.locale) ?? ""
    }

    var body: some View {
        NavigationStack {
            VStack(alignment: .leading, spacing: metrics.points(14)) {
                Text(label).typeStyle(Typography.budgetFigure).foregroundStyle(Palette.label).contentTransition(.numericText())
                Slider(value: $model.budgetMajor, in: 800...12000, step: 100)
                    .tint(Palette.sky)
                    .accessibilityLabel(model.copy.budgetTitle)
                    .accessibilityValue(label)
                Text(model.copy.budgetHint).typeStyle(Typography.bodyCompact).foregroundStyle(Palette.labelAlpha(0.62))
                Spacer()
            }
            .padding(metrics.points(24))
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
            .background(Palette.background)
            .navigationTitle(model.copy.budgetTitle)
            .toolbar { ToolbarItem(placement: .confirmationAction) { Button(model.copy.done) { dismiss() } } }
            .toolbarTitleDisplayMode(.inline)
            .animation(Motion.fade(0.2), value: model.budgetMajor)
        }
        .presentationDetents([.medium])
    }
}
