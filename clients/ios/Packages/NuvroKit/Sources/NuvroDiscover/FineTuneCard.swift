import NuvroDesignSystem
import SwiftUI

struct FineTuneCard: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @Bindable var model: MatchSetupModel

    var body: some View {
        VStack(spacing: 0) {
            Button {
                withAnimation(reduceMotion ? nil : Motion.settle(0.35)) { model.isFineTuneExpanded.toggle() }
            } label: {
                HStack {
                    VStack(alignment: .leading, spacing: metrics.points(2)) {
                        Text(model.copy.fineTune).typeStyle(Typography.cardTitleLarge).foregroundStyle(Palette.label)
                        Text(model.fineTuneSummary).typeStyle(Typography.caption).foregroundStyle(Palette.labelAlpha(0.55))
                    }
                    Spacer()
                    Image(systemName: "chevron.down")
                        .font(.system(size: metrics.points(13), weight: .semibold))
                        .foregroundStyle(Palette.labelAlpha(0.5))
                        .rotationEffect(.degrees(model.isFineTuneExpanded ? 180 : 0))
                }
                .padding(EdgeInsets(top: metrics.points(14), leading: metrics.points(16), bottom: metrics.points(14), trailing: metrics.points(16)))
                .contentShape(Rectangle())
            }
            .buttonStyle(.plain)
            .accessibilityAddTraits(.isButton)
            .accessibilityValue(model.isFineTuneExpanded ? "Expanded" : "Collapsed")

            if model.isFineTuneExpanded {
                VStack(alignment: .leading, spacing: metrics.points(14)) {
                    Divider().overlay(Palette.labelAlpha(0.08))
                    segment("Climate", options: Climate.allCases, selection: $model.climate) { $0.rawValue }
                    segment("Pace", options: Pace.allCases, selection: $model.pace) { $0.rawValue }
                    segment("Tier", options: Tier.allCases, selection: $model.tier) { $0.rawValue }
                    segment("Scope", options: Scope.allCases, selection: $model.scope) { $0.rawValue }
                    VStack(spacing: metrics.points(10)) {
                        toggleRow("Direct flights preferred", isOn: $model.prefersDirectFlights)
                        toggleRow("Easier entry preferred", isOn: $model.prefersEasierEntry)
                    }
                    .padding(.top, metrics.points(6))
                }
                .padding(EdgeInsets(top: 0, leading: metrics.points(16), bottom: metrics.points(16), trailing: metrics.points(16)))
                .transition(.opacity)
            }
        }
        .surfaceCard()
    }

    private func segment<Option: Hashable>(_ label: String, options: [Option], selection: Binding<Option>, title: @escaping (Option) -> String) -> some View {
        VStack(alignment: .leading, spacing: metrics.points(8)) {
            FieldLabel(label)
            SegmentedChoice(options: options, selection: selection, compact: true, label: title)
        }
    }

    private func toggleRow(_ label: String, isOn: Binding<Bool>) -> some View {
        Toggle(isOn: isOn) {
            Text(label).typeStyle(Typography.body).foregroundStyle(Palette.label)
        }
        .toggleStyle(.switch)
        .tint(Palette.sky)
    }
}
