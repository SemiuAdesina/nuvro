import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import SwiftUI

public struct LiveReplanningView: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @State private var isApplied = false
    @State private var showingAlternative = false
    private let model: TravelModeModel
    private let onFinished: () -> Void

    public init(model: TravelModeModel, onFinished: @escaping () -> Void) {
        self.model = model
        self.onFinished = onFinished
    }

    private var adjustment: PlanAdjustment { model.pendingAdjustment }
    private var copy: ReplanCopy { model.replanCopy }

    public var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: metrics.points(16)) {
                HStack(spacing: metrics.points(8)) {
                    Image(systemName: "cloud.rain.fill").foregroundStyle(Palette.sky)
                    Text("“\(adjustment.trigger)”")
                }
                .typeStyle(Typography.control).foregroundStyle(Palette.label)
                .padding(.horizontal, metrics.points(14)).padding(.vertical, metrics.points(9))
                .glassCapsule(.light)
                comparison
                InsightCard(title: copy.whyThisChange) {
                    Text(adjustment.reasoning).typeStyle(Typography.bodyCompact).foregroundStyle(Palette.labelAlpha(0.75))
                }
                VStack(alignment: .leading, spacing: metrics.points(10)) {
                    Text(copy.staysTheSame).typeStyle(TypeStyle(size: 15, weight: .bold)).foregroundStyle(Palette.label)
                    ForEach(adjustment.preserved, id: \.self) { CheckRow($0) }
                }
                .padding(metrics.points(16))
                .frame(maxWidth: .infinity, alignment: .leading)
                .surfaceCard()
                LinkButton(showingAlternative ? copy.showingOption : copy.anotherOption) { showingAlternative.toggle() }
                    .frame(maxWidth: .infinity)
                if isApplied {
                    HStack(spacing: metrics.points(10)) {
                        Image(systemName: "checkmark").font(.system(size: metrics.points(12), weight: .bold)).foregroundStyle(.white)
                            .frame(width: metrics.points(22), height: metrics.points(22)).background(Palette.aqua, in: Circle())
                        Text(copy.planUpdated).typeStyle(Typography.bodyCompactEmphasis).foregroundStyle(Palette.label)
                    }
                    .padding(.horizontal, metrics.points(14)).padding(.vertical, metrics.points(12))
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .background(Palette.aqua.opacity(0.16), in: RoundedRectangle(cornerRadius: metrics.points(Radius.notice), style: .continuous))
                    .overlay(RoundedRectangle(cornerRadius: metrics.points(Radius.notice), style: .continuous).strokeBorder(Palette.aqua.opacity(0.45), lineWidth: 0.5))
                    .transition(.opacity)
                }
                InlineActions {
                    CallToActionButton(copy.applyPlan) { Task { await apply() } }
                        .disabled(isApplied)
                    Button { model.keepOriginal(); onFinished() } label: {
                        Text(copy.keepOriginal).typeStyle(Typography.secondaryAction).foregroundStyle(Palette.labelAlpha(0.7))
                            .frame(maxWidth: .infinity).frame(height: metrics.points(36))
                    }
                    .buttonStyle(.plain)
                    .disabled(isApplied)
                }
            }
            .padding(.horizontal, metrics.points(20))
            .padding(.top, metrics.points(8))
            .padding(.bottom, metrics.points(40))
            .animation(Motion.settle(0.6), value: isApplied)
        }
        .scrollIndicators(.hidden)
        .background(Palette.background.ignoresSafeArea())
        .screenTitle(copy.title, subtitle: copy.body)
        .navigationBarBackButtonHidden(true)
        .toolbar {
            ToolbarItem(placement: .cancellationAction) {
                Button { model.keepOriginal(); onFinished() } label: {
                    Image(systemName: "xmark").font(.system(size: metrics.points(14), weight: .semibold)).foregroundStyle(Palette.label)
                }
                .accessibilityLabel(copy.keepOriginal)
            }
        }
    }

    private var comparison: some View {
        VStack(alignment: .leading, spacing: metrics.points(14)) {
            VStack(alignment: .leading, spacing: metrics.points(8)) {
                Text(copy.before).typeStyle(TypeStyle(size: 11, weight: .bold, tracking: 1.3)).foregroundStyle(Palette.labelAlpha(0.5))
                ComparisonRow(activity: adjustment.before, tone: .before, isStruck: isApplied)
            }
            .opacity(isApplied ? 0.35 : 1)
            .offset(y: isApplied && !reduceMotion ? -metrics.points(6) : 0)
            VStack(alignment: .leading, spacing: metrics.points(8)) {
                Text(isApplied ? copy.applied : copy.suggested).typeStyle(TypeStyle(size: 11, weight: .bold, tracking: 1.3)).foregroundStyle(Palette.sky)
                ComparisonRow(activity: adjustment.after, tone: .after, isStruck: false)
                    .offset(y: isApplied && !reduceMotion ? -metrics.points(4) : 0)
            }
            HStack(spacing: metrics.points(9)) {
                Image(systemName: "arrow.turn.down.right").font(.system(size: metrics.points(12), weight: .semibold)).foregroundStyle(Palette.amber)
                Text(adjustment.relocationNote).typeStyle(TypeStyle(size: 13.5, weight: .regular)).foregroundStyle(Palette.label)
            }
            .padding(.horizontal, metrics.points(12)).padding(.vertical, metrics.points(10))
            .frame(maxWidth: .infinity, alignment: .leading)
            .background(Palette.gold.opacity(0.18), in: RoundedRectangle(cornerRadius: metrics.points(Radius.chip), style: .continuous))
        }
        .padding(metrics.points(16))
        .surfaceCard(radius: 22)
    }

    private func apply() async {
        withAnimation(reduceMotion ? nil : Motion.settle(0.6)) { isApplied = true }
        model.beginApplying()
        try? await Task.sleep(for: .seconds(1.1))
        model.applyPendingAdjustment()
        onFinished()
    }
}

struct ComparisonRow: View {
    enum Tone { case before, after }

    @Environment(\.designMetrics) private var metrics
    let activity: ItineraryActivity
    let tone: Tone
    let isStruck: Bool

    var body: some View {
        HStack(spacing: metrics.points(12)) {
            Text(activity.time).typeStyle(Typography.metaEmphasis).foregroundStyle(tone == .after ? Palette.sky : Palette.labelAlpha(0.6)).frame(width: metrics.points(42), alignment: .leading)
            VStack(alignment: .leading, spacing: metrics.points(2)) {
                Text(activity.title).typeStyle(Typography.secondaryAction).foregroundStyle(tone == .after ? Palette.label : Palette.labelAlpha(0.7)).strikethrough(isStruck)
                Text(activity.detail).typeStyle(Typography.caption).foregroundStyle(Palette.labelAlpha(tone == .after ? 0.6 : 0.55))
            }
            Spacer(minLength: 0)
        }
        .padding(.horizontal, metrics.points(12)).padding(.vertical, metrics.points(11))
        .background(tone == .after ? Palette.sky.opacity(0.08) : Palette.labelAlpha(0.045), in: RoundedRectangle(cornerRadius: metrics.points(16), style: .continuous))
        .overlay {
            if tone == .after {
                RoundedRectangle(cornerRadius: metrics.points(16), style: .continuous).strokeBorder(Palette.sky.opacity(0.28), lineWidth: 0.5)
            }
        }
        .accessibilityElement(children: .combine)
    }
}
