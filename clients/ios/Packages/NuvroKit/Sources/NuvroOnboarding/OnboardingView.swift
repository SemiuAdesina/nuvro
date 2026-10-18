import NuvroDesignSystem
import NuvroLocalization
import SwiftUI

public struct OnboardingView: View {
    @Bindable private var model: OnboardingFlowModel
    @Environment(\.accessibilityReduceMotion) private var systemReduceMotion
    @State private var lightProgress: CGFloat = 0

    public init(model: OnboardingFlowModel) {
        self.model = model
    }

    private var reduceMotion: Bool { model.reduceMotion || systemReduceMotion }

    public var body: some View {
        DesignCanvas { metrics in
            ZStack(alignment: .topLeading) {
                SkyGradient.dawn
                world(metrics)
                SkyGradient.groundScrim
                    .frame(height: metrics.points(420))
                    .offset(y: metrics.size.height - metrics.points(420))
                    .allowsHitTesting(false)
                SplashLogoView(stage: model.stage, hasBooted: model.hasBooted, reduceMotion: reduceMotion, tagline: model.copy.tagline)
                DestinationCardsView(stage: model.stage, reduceMotion: reduceMotion, matchLabel: model.copy.match)
                PlanningChipsView(stage: model.stage, reduceMotion: reduceMotion)
                LiveTripView(stage: model.stage, hasReplanned: model.hasReplanned, reduceMotion: reduceMotion, copy: model.copy)
                CopyBlockView(stage: model.stage, dragOffset: model.dragOffset, reduceMotion: reduceMotion, panels: panels)
                PageIndicatorView(stage: model.stage, isShown: model.isShowingScreenControls)
                PrimaryActionsView(
                    callToAction: model.callToActionLabel,
                    secondaryAction: model.copy.secondaryAction3,
                    isShown: model.isShowingScreenControls,
                    isSecondaryShown: model.stage == .travel && model.exit == nil,
                    reduceMotion: reduceMotion,
                    onPrimary: model.advance,
                    onSecondary: model.chooseExistingTrip
                )
                TopControlsView(
                    language: model.language,
                    skipLabel: model.copy.skip,
                    isLanguageShown: model.isShowingLanguagePill,
                    isSkipShown: model.isShowingScreenControls,
                    onLanguage: model.presentLanguageSheet,
                    onSkip: model.skip
                )
            }
            .frame(width: metrics.size.width, height: metrics.size.height, alignment: .topLeading)
            .clipped()
            .contentShape(Rectangle())
            .gesture(swipe)
        }
        .environment(\.layoutDirection, model.language.isRightToLeft ? .rightToLeft : .leftToRight)
        .task { model.boot(); withAnimation(reduceMotion ? nil : Motion.plane) { lightProgress = 1 } }
        .task(id: model.stage) { await runStageTimer() }
        .sheet(isPresented: $model.isLanguageSheetPresented) {
            LanguageSheetView(
                title: model.copy.languageSheetTitle,
                searchPlaceholder: model.copy.languageSearch,
                selected: model.language,
                languages: model.filteredLanguages,
                query: $model.languageQuery,
                onSelect: model.select,
                onClose: model.dismissLanguageSheet
            )
        }
    }

    private func world(_ metrics: DesignMetrics) -> some View {
        let pose = CameraPose.pose(for: model.stage, reduceMotion: reduceMotion)
        let anchor = UnitPoint(x: CameraPose.anchor.x, y: CameraPose.anchor.y)
        return ZStack(alignment: .topLeading) {
            SkyLayer().modifier(CameraLayer(pose: pose, depth: LayerDepth.far, anchor: anchor, scale: metrics.scale))
            GlobeLayer(lightProgress: lightProgress).modifier(CameraLayer(pose: pose, depth: LayerDepth.mid, anchor: anchor, scale: metrics.scale))
            LandmarkLayer().modifier(CameraLayer(pose: pose, depth: LayerDepth.near, anchor: anchor, scale: metrics.scale))
        }
        .opacity(StagePoses.worldOpacity(model.stage))
        .animation(reduceMotion ? nil : Motion.settle(Motion.cameraDuration), value: pose)
        .animation(Motion.fade(reduceMotion ? Motion.reducedWorldFadeDuration : Motion.worldFadeDuration), value: model.stage)
    }

    private var panels: [CopyPanel] {
        let copy = model.copy
        return [
            CopyPanel(stage: .discover, headline: copy.headline1, supporting: copy.supporting1, body: copy.body1),
            CopyPanel(stage: .plan, headline: copy.headline2, supporting: nil, body: copy.body2),
            CopyPanel(stage: .travel, headline: model.headline3, supporting: model.supporting3, body: copy.body3)
        ]
    }

    private var swipe: some Gesture {
        DragGesture(minimumDistance: 4)
            .onChanged { model.dragChanged($0.translation.width) }
            .onEnded { model.dragEnded($0.translation.width) }
    }

    private func runStageTimer() async {
        switch model.stage {
        case .splash:
            try? await Task.sleep(for: .seconds(Motion.splashHold))
            guard !Task.isCancelled else { return }
            model.splashDidElapse()
        case .travel:
            try? await Task.sleep(for: .seconds(Motion.replanDelay))
            guard !Task.isCancelled else { return }
            model.replanDidElapse()
        case .discover, .plan:
            return
        }
    }
}

struct CameraLayer: ViewModifier {
    let pose: CameraPose
    let depth: CGFloat
    let anchor: UnitPoint
    let scale: CGFloat

    func body(content: Content) -> some View {
        content
            .scaleEffect(pose.layerScale(depth: depth), anchor: anchor)
            .offset(y: pose.layerOffset(depth: depth) * scale)
    }
}
