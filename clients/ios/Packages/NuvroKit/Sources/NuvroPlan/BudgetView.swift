import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import SwiftUI

public struct BudgetView: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @State private var expandedCategory: String?
    private let model: TripPlanModel
    private let copy: BudgetCopy
    private let onViewStays: () -> Void

    public init(model: TripPlanModel, language: Language = .english, onViewStays: @escaping () -> Void) {
        self.model = model
        self.copy = .resolve(language)
        self.onViewStays = onViewStays
    }

    private var plan: TripPlan { model.plan }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(18)) {
                totalCard
                categories
                InsightCard(title: "\(copy.saveAbout) \(savings)") {
                    Text(copy.insight).typeStyle(Typography.bodyCompact).foregroundStyle(Palette.labelAlpha(0.75))
                    SurfaceLinkButton(copy.seeOption, action: onViewStays)
                }
                Text(copy.disclaimer).typeStyle(Typography.small).foregroundStyle(Palette.labelAlpha(0.45)).frame(maxWidth: .infinity)
                InlineActions {
                    CallToActionButton(copy.viewStays, action: onViewStays)
                    Button { model.editing = .budget } label: {
                        Text(copy.adjustBudget).typeStyle(Typography.secondaryAction).foregroundStyle(Palette.sky)
                            .frame(maxWidth: .infinity).frame(height: metrics.points(34))
                    }
                    .buttonStyle(.plain)
                }
            }
            .padding(.horizontal, metrics.points(20))
            .padding(.top, metrics.points(8))
            .padding(.bottom, metrics.points(40))
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .screenTitle(copy.title, subtitle: "\(plan.destination.name) · \(plan.destination.durationDays) days")
        .sheet(isPresented: Binding(get: { model.editing == .budget }, set: { if !$0 { model.editing = nil } }), onDismiss: model.applyEdits) {
            TripBudgetSheet(model: model)
        }
    }

    private var savings: String {
        (try? Money(amountMinor: 18000, currency: plan.budgetCeiling.currency))?.formattedWhole(locale: model.locale) ?? ""
    }

    private var totalCard: some View {
        VStack(alignment: .leading, spacing: metrics.points(14)) {
            HStack(alignment: .bottom, spacing: metrics.points(14)) {
                VStack(alignment: .leading, spacing: metrics.points(3)) {
                    FieldLabel(copy.estimatedTotal)
                    Text(model.estimateLabel).typeStyle(TypeStyle(size: 42, weight: .bold, tracking: -1.2)).foregroundStyle(Palette.label)
                    HStack(spacing: 0) {
                        Text("\(copy.budget) \(plan.budgetCeiling.formattedWhole(locale: model.locale)) · ").typeStyle(Typography.meta).foregroundStyle(Palette.labelAlpha(0.6))
                        Text(remainingLabel).typeStyle(Typography.metaEmphasis).foregroundStyle(model.isWithinBudget ? Palette.matchTeal : Palette.amber)
                    }
                }
                Spacer(minLength: 0)
                BudgetRing(
                    segments: plan.budgetLines.map { (BudgetCategoryTone.color(for: $0.category), model.share(of: $0)) },
                    centreTitle: "\(Int((plan.budgetFraction * 100).rounded()))%",
                    centreLabel: copy.ofBudget
                )
            }
            HStack(spacing: metrics.points(9)) {
                Image(systemName: model.isWithinBudget ? "checkmark" : "exclamationmark")
                    .font(.system(size: metrics.points(11), weight: .bold))
                    .foregroundStyle(model.isWithinBudget ? Palette.matchTeal : Palette.amber)
                    .frame(width: metrics.points(20), height: metrics.points(20))
                    .background((model.isWithinBudget ? Palette.aqua : Palette.gold).opacity(0.35), in: Circle())
                Text(model.isWithinBudget ? copy.withinBudget : copy.overBudget).typeStyle(TypeStyle(size: 13.5, weight: .regular)).foregroundStyle(Palette.label)
            }
            .padding(.horizontal, metrics.points(12))
            .padding(.vertical, metrics.points(10))
            .frame(maxWidth: .infinity, alignment: .leading)
            .background((model.isWithinBudget ? Palette.aqua : Palette.gold).opacity(0.14), in: RoundedRectangle(cornerRadius: metrics.points(Radius.chip), style: .continuous))
        }
        .padding(metrics.points(18))
        .surfaceCard(radius: 22)
    }

    private var remainingLabel: String {
        guard let remaining = plan.remainingBudget, let magnitude = try? Money(amountMinor: abs(remaining.amountMinor), currency: remaining.currency) else { return "" }
        return "\(magnitude.formattedWhole(locale: model.locale)) \(remaining.amountMinor >= 0 ? copy.remaining : copy.over)"
    }

    private var categories: some View {
        VStack(spacing: 0) {
            ForEach(plan.budgetLines) { line in
                BudgetCategoryRow(
                    line: line, copy: copy, locale: model.locale,
                    isExpanded: expandedCategory == line.id,
                    onToggle: { withAnimation(reduceMotion ? nil : Motion.settle(0.3)) { expandedCategory = expandedCategory == line.id ? nil : line.id } }
                )
                if line.id != plan.budgetLines.last?.id {
                    Divider().overlay(Palette.labelAlpha(0.07))
                }
            }
        }
        .surfaceCard(radius: 22)
    }
}

struct BudgetCategoryRow: View {
    @Environment(\.designMetrics) private var metrics
    let line: BudgetLine
    let copy: BudgetCopy
    let locale: Locale
    let isExpanded: Bool
    let onToggle: () -> Void

    private var stateLabel: String {
        if let booked = line.booked, booked != line.amount {
            return "\(copy.estimated) \(line.amount.formattedWhole(locale: locale)) · \(copy.booked) \(booked.formattedWhole(locale: locale))"
        }
        return line.isBooked ? copy.booked : copy.estimated
    }

    var body: some View {
        VStack(spacing: 0) {
            Button(action: onToggle) {
                HStack(spacing: metrics.points(12)) {
                    Circle().fill(BudgetCategoryTone.color(for: line.category)).frame(width: metrics.points(10), height: metrics.points(10))
                    VStack(alignment: .leading, spacing: metrics.points(2)) {
                        Text(line.category).typeStyle(TypeStyle(size: 15.5, weight: .semibold)).foregroundStyle(Palette.label)
                        Text(stateLabel).typeStyle(Typography.caption).foregroundStyle(Palette.labelAlpha(0.55))
                    }
                    Spacer(minLength: 0)
                    Text(line.amount.formattedWhole(locale: locale)).typeStyle(TypeStyle(size: 16, weight: .bold, tracking: -0.2)).foregroundStyle(Palette.label)
                    Image(systemName: "chevron.down")
                        .font(.system(size: metrics.points(12), weight: .semibold))
                        .foregroundStyle(Palette.labelAlpha(0.35))
                        .rotationEffect(.degrees(isExpanded ? 180 : 0))
                }
                .padding(.horizontal, metrics.points(16))
                .padding(.vertical, metrics.points(13))
                .contentShape(Rectangle())
            }
            .buttonStyle(.plain)
            .accessibilityValue(isExpanded ? "Expanded" : "Collapsed")
            if isExpanded {
                VStack(spacing: metrics.points(7)) {
                    ForEach(line.details) { detail in
                        HStack {
                            Text(detail.label).typeStyle(Typography.meta).foregroundStyle(Palette.labelAlpha(0.65))
                            Spacer()
                            Text(detail.amount.formattedWhole(locale: locale)).typeStyle(Typography.metaEmphasis).foregroundStyle(Palette.label)
                        }
                    }
                }
                .padding(EdgeInsets(top: 0, leading: metrics.points(38), bottom: metrics.points(14), trailing: metrics.points(16)))
                .transition(.opacity)
            }
        }
    }
}
