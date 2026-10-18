import NuvroDesignSystem
import NuvroModels
import SwiftUI

public struct MatchSetupView: View {
    @Environment(\.designMetrics) private var metrics
    @Bindable private var model: MatchSetupModel
    private let onFindMatches: () -> Void

    public init(model: MatchSetupModel, onFindMatches: @escaping () -> Void) {
        self.model = model
        self.onFindMatches = onFindMatches
    }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(20)) {
                VStack(alignment: .leading, spacing: metrics.points(8)) {
                    Text(model.copy.title).typeStyle(Typography.headline).foregroundStyle(Palette.label)
                    Text(model.copy.body).typeStyle(Typography.body).foregroundStyle(Palette.labelAlpha(0.62))
                }
                DepartureAndTravelersCard(model: model)
                BudgetCard(model: model)
                DatesCard(model: model)
                PreferenceGrid(title: model.copy.preferencesTitle, countLabel: model.preferenceCountLabel, isSelected: model.isSelected, onToggle: model.toggle)
                FineTuneCard(model: model)
                InlineActions {
                    CallToActionButton(model.copy.findMatches, action: onFindMatches)
                        .disabled(!model.canSubmit)
                        .opacity(model.canSubmit ? 1 : 0.5)
                }
            }
            .padding(.horizontal, metrics.points(20))
            .padding(.top, metrics.points(8))
            .padding(.bottom, metrics.points(40))
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .sheet(item: $model.presentation) { presentation in
            switch presentation {
            case .airport:
                SearchablePickerSheet(
                    title: model.pickerCopy.departureAirport, searchPrompt: model.pickerCopy.searchAirports, doneLabel: model.stateCopy.done,
                    emptyTitle: model.pickerCopy.noAirports, emptyMessage: model.pickerCopy.noAirportsBody, options: model.airports,
                    selection: Binding(get: { model.departureAirport?.id }, set: { code in model.departureAirport = model.airports.first { $0.id == code } }),
                    label: \.shortLabel, detail: { "\($0.name) · \($0.country)" }, matches: { $0.matches($1) }
                )
            case .dates:
                DateSelectionSheet(selection: $model.dateSelection, copy: model.dateSheetCopy, today: model.today)
            case .travelers:
                CountPickerSheet(title: model.pickerCopy.travelers, doneLabel: model.stateCopy.done, lines: model.travelerLines, counts: $model.travelerCounts)
            }
        }
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Text(model.copy.duration)
                    .typeStyle(Typography.metaEmphasis)
                    .foregroundStyle(Palette.labelAlpha(0.6))
                    .padding(.horizontal, metrics.points(6))
                    .fixedSize()
            }
        }
    }
}
