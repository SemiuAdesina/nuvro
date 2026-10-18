import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import SwiftUI

public struct FullItineraryView: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(TransientFeedback.self) private var feedback: TransientFeedback?
    @Bindable private var model: ItineraryModel
    private let statusText: String
    private let onOptimize: () -> Void
    private let onViewBudget: () -> Void

    public init(model: ItineraryModel, statusText: String, onOptimize: @escaping () -> Void, onViewBudget: @escaping () -> Void) {
        self.model = model
        self.statusText = statusText
        self.onOptimize = onOptimize
        self.onViewBudget = onViewBudget
    }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(18)) {
                TripIdentityHeader(destination: model.currentPlan.destination, subtitle: "\(model.currentPlan.dateRangeLabel) · \(model.currentPlan.destination.durationDays) days", compact: true)
                DaySelector(days: model.days, selected: $model.selectedDay).padding(.horizontal, -metrics.points(20))
                if let day = model.day {
                    HStack(alignment: .firstTextBaseline) {
                        Text(day.title).typeStyle(Typography.title).foregroundStyle(Palette.label)
                        Spacer()
                        Text(model.daySubtitle).typeStyle(Typography.meta).foregroundStyle(Palette.labelAlpha(0.55))
                    }
                    VStack(spacing: 0) {
                        ForEach(Array(day.activities.enumerated()), id: \.element.id) { index, activity in
                            TimelineRow(time: activity.time, isFirst: index == 0, isLast: index == day.activities.count - 1, travelToNext: activity.travelToNext) {
                                ItineraryActivityRow(activity: activity, statusLabel: activity.status.map(model.statusLabel), copy: model.copy, onMove: { model.moving = activity }, onRemove: { model.requestRemove(activity) })
                            }
                        }
                    }
                    .id(day.dayNumber)
                    .transition(.opacity)
                }
                InsightCard(title: model.copy.askNuvro) {
                    Text(model.copy.askExample).typeStyle(Typography.meta).foregroundStyle(Palette.labelAlpha(0.6))
                }
                InlineActions {
                    CallToActionButton(model.copy.optimize, action: onOptimize)
                    Button(action: onViewBudget) {
                        Text(model.copy.viewBudget).typeStyle(Typography.secondaryAction).foregroundStyle(Palette.sky)
                            .frame(maxWidth: .infinity).frame(height: metrics.points(34))
                    }
                    .buttonStyle(.plain)
                }
            }
            .padding(.horizontal, metrics.points(20))
            .padding(.top, metrics.points(8))
            .padding(.bottom, metrics.points(40))
            .animation(Motion.fade(0.3), value: model.selectedDay)
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .alert(model.copy.removeTitle, isPresented: Binding(get: { model.pendingRemoval != nil }, set: { if !$0 { model.cancelRemoval() } })) {
            Button(model.copy.remove, role: .destructive) { if let activity = model.pendingRemoval { model.remove(activity) } }
            Button(model.copy.cancel, role: .cancel) { model.cancelRemoval() }
        } message: {
            Text(model.copy.removeBody)
        }
        .sheet(item: $model.moving) { activity in
            OptionPickerSheet(title: "\(model.copy.moveTitle)…", doneLabel: model.copy.done, options: model.dayOptions, selection: Binding(get: { "" }, set: { if let day = Int($0) { model.move(activity, toDay: day) } }))
        }
        .onChange(of: model.lastEvent) { _, event in if let event { feedback?.show(event); model.clearEvent() } }
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                HStack(spacing: metrics.points(6)) {
                    Circle().fill(Palette.gold).frame(width: metrics.points(6), height: metrics.points(6))
                    Text(statusText).typeStyle(Typography.statusPill).foregroundStyle(Palette.amber)
                }
                .padding(.horizontal, metrics.points(6))
                .fixedSize()
            }
        }
    }
}

struct ItineraryActivityRow: View {
    @Environment(\.designMetrics) private var metrics
    let activity: ItineraryActivity
    let statusLabel: String?
    let copy: ItineraryCopy
    let onMove: () -> Void
    let onRemove: () -> Void

    var body: some View {
        HStack(spacing: metrics.points(11)) {
            if let photo = activity.photoReference {
                PhotoStandIn(palette: .named(photo), label: "", labelSize: 8)
                    .frame(width: metrics.points(46), height: metrics.points(42))
                    .clipShape(RoundedRectangle(cornerRadius: metrics.points(Radius.stop), style: .continuous))
            }
            VStack(alignment: .leading, spacing: metrics.points(2)) {
                Text(activity.title).typeStyle(TypeStyle(size: 15.5, weight: .semibold, tracking: -0.1)).foregroundStyle(Palette.label)
                Text(activity.detail).typeStyle(TypeStyle(size: 12.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
            }
            Spacer(minLength: 0)
            if let statusLabel, let status = activity.status {
                let isGold = status == .ticket || status == .weather
                Text(statusLabel)
                    .typeStyle(TypeStyle(size: 11, weight: .semibold))
                    .foregroundStyle(isGold ? Palette.amber : Palette.matchTeal)
                    .padding(.horizontal, metrics.points(8)).padding(.vertical, metrics.points(4))
                    .background((isGold ? Palette.gold : Palette.aqua).opacity(0.22), in: Capsule())
                    .fixedSize()
            }
            Menu {
                Button(copy.move, systemImage: "arrow.up.arrow.down", action: onMove)
                Button(copy.remove, systemImage: "trash", role: .destructive, action: onRemove)
            } label: {
                Image(systemName: "ellipsis")
                    .font(.system(size: metrics.points(15), weight: .semibold))
                    .foregroundStyle(Palette.labelAlpha(0.3))
                    .frame(width: metrics.points(24), height: metrics.points(28))
            }
            .accessibilityLabel("\(copy.more), \(activity.title)")
        }
        .contentShape(Rectangle())
    }
}
