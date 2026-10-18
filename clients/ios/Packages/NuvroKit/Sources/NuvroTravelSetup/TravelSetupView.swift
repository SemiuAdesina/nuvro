import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import SwiftUI

public struct TravelSetupView: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @Bindable private var model: TravelSetupModel
    private let onComplete: (TravelProfile) -> Void
    private let onCancel: (() -> Void)?

    public init(model: TravelSetupModel, onComplete: @escaping (TravelProfile) -> Void, onCancel: (() -> Void)? = nil) {
        self.model = model
        self.onComplete = onComplete
        self.onCancel = onCancel
    }

    private var copy: TravelSetupCopy { model.copy }

    private var transition: AnyTransition {
        if reduceMotion { return .opacity }
        let forward = model.isMovingForward
        return .asymmetric(
            insertion: .move(edge: forward ? .trailing : .leading).combined(with: .opacity),
            removal: .move(edge: forward ? .leading : .trailing).combined(with: .opacity)
        )
    }

    public var body: some View {
        VStack(spacing: 0) {
            header
            ScrollView {
                stepContent
                    .id(model.step)
                    .transition(transition)
                    .padding(.horizontal, metrics.points(20))
                    .padding(.top, metrics.points(8))
                    .padding(.bottom, metrics.points(24))
            }
            .scrollIndicators(.hidden)
            .scrollDismissesKeyboard(.interactively)
            .animation(reduceMotion ? Motion.fade(0.25) : Motion.settle(0.5), value: model.step)
            footer
        }
        .background(Palette.background.ignoresSafeArea())
        .setupPickers(model: model)
        .onChange(of: model.completedProfile) { _, profile in if let profile { onComplete(profile) } }
    }

    private var header: some View {
        VStack(spacing: metrics.points(10)) {
            HStack {
                if model.canGoBack {
                    Button(action: model.back) {
                        Label(copy.back, systemImage: "chevron.backward").labelStyle(.iconOnly)
                            .font(.system(size: metrics.points(17), weight: .semibold)).foregroundStyle(Palette.label)
                            .frame(width: metrics.points(38), height: metrics.points(38))
                    }
                    .buttonStyle(GlassPillButtonStyle(style: .light))
                    .accessibilityLabel(copy.back)
                } else if let onCancel {
                    Button(copy.done, action: onCancel).typeStyle(Typography.link).foregroundStyle(Palette.sky).buttonStyle(.plain)
                }
                Spacer()
                if model.showsProgress {
                    Text("\(model.progressIndex) \(copy.progressOf) \(model.progressSteps.count)")
                        .typeStyle(Typography.metaEmphasis).foregroundStyle(Palette.labelAlpha(0.55))
                        .contentTransition(.numericText())
                        .accessibilityLabel("Step \(model.progressIndex) of \(model.progressSteps.count)")
                }
            }
            if model.showsProgress {
                GradientProgressBar(fraction: model.progressFraction, height: 4)
                    .animation(reduceMotion ? nil : Motion.settle(0.4), value: model.progressFraction)
                    .accessibilityHidden(true)
            }
        }
        .padding(.horizontal, metrics.points(20))
        .padding(.top, metrics.points(10))
        .padding(.bottom, metrics.points(6))
    }

    @ViewBuilder
    private var stepContent: some View {
        switch model.step {
        case .welcome: WelcomeStep(model: model)
        case .residence: ResidenceStep(model: model)
        case .passports: PassportStep(model: model)
        case .permits: PermitStep(model: model)
        case .airport: AirportStep(model: model)
        case .currency: CurrencyStep(model: model)
        case .paperwork: PaperworkStep(model: model)
        case .interests: InterestsStep(model: model)
        case .pace: PaceStep(model: model)
        case .weather: WeatherStep(model: model)
        case .crowds: CrowdsStep(model: model)
        case .flights: FlightsStep(model: model)
        case .stays: StaysStep(model: model)
        case .group: GroupStep(model: model)
        case .accessibility: AccessibilityStep(model: model)
        case .review: ReviewStep(model: model)
        }
    }

    private var footer: some View {
        VStack(spacing: metrics.points(8)) {
            CallToActionButton(ctaLabel, action: model.advance)
                .disabled(!model.canContinue)
                .opacity(model.canContinue ? 1 : 0.55)
            if model.canSkip {
                Button(copy.skip, action: model.skip).typeStyle(Typography.secondaryAction).foregroundStyle(Palette.labelAlpha(0.6)).buttonStyle(.plain)
                    .frame(height: metrics.points(34))
            }
        }
        .padding(.horizontal, metrics.points(24))
        .padding(.top, metrics.points(10))
        .padding(.bottom, metrics.points(12))
        .background(Palette.background)
    }

    private var ctaLabel: String {
        if model.isEditing { return copy.done }
        return model.step == .review ? copy.finish : copy.continueLabel
    }
}
