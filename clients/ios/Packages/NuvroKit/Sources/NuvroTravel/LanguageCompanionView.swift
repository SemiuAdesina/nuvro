import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import NuvroPermissions
import SwiftUI

public struct LanguageCompanionView: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(PermissionCenter.self) private var permissions: PermissionCenter?
    @Environment(TransientFeedback.self) private var feedback: TransientFeedback?
    @Bindable private var model: LanguageCompanionModel
    private let permissionCopy = PermissionCopy.resolve(.english)
    private let dayLabel: String
    private let onNext: () -> Void

    public init(model: LanguageCompanionModel, dayLabel: String, onNext: @escaping () -> Void) {
        self.model = model
        self.dayLabel = dayLabel
        self.onNext = onNext
    }

    private var copy: LanguageCopy { model.copy }

    private func toggleListening() {
        Task {
            let allowed = await permissions?.prepare(.microphone) ?? true
            if allowed { model.toggleListening() }
        }
    }

    public var body: some View {
        ZStack {
            ScrollView {
                VStack(alignment: .leading, spacing: metrics.points(16)) {
                    LanguagePairSelector(source: model.pair.source, target: model.pair.target, swap: model.swapLanguages)
                    SegmentedChoice(options: CompanionMode.allCases, selection: $model.mode) { modeLabel($0) }
                    modeContent
                    phraseList
                    HStack(spacing: metrics.points(8)) {
                        OfflineMark()
                        Text(copy.offline).typeStyle(TypeStyle(size: 12.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
                    }
                    .padding(.horizontal, metrics.points(4))
                    InlineActions {
                        CallToActionButton(copy.showToSomeone) {
                            if model.result == nil { model.showSample() }
                            model.isShowingLarge = true
                        }
                        Button(action: model.swapLanguages) {
                            Text(copy.changeLanguages).typeStyle(Typography.secondaryAction).foregroundStyle(Palette.sky).frame(maxWidth: .infinity).frame(height: metrics.points(34))
                        }
                        .buttonStyle(.plain)
                        Button(action: onNext) {
                            Text("Explore nearby").typeStyle(Typography.secondaryAction).foregroundStyle(Palette.labelAlpha(0.5)).frame(maxWidth: .infinity).frame(height: metrics.points(30))
                        }
                        .buttonStyle(.plain)
                    }
                }
                .padding(.horizontal, metrics.points(20))
                .padding(.top, metrics.points(8))
                .padding(.bottom, metrics.points(40))
                .animation(Motion.fade(0.3), value: model.mode)
                .animation(Motion.settle(0.5), value: model.result)
            }
            .scrollIndicators(.hidden)
            .background(Palette.background.ignoresSafeArea())
            .screenTitle(copy.title, subtitle: "\(model.pair.target == "English" ? "English" : "Portuguese") \(copy.spokenHere)")
            .task(id: model.mode) { if model.mode == .camera { _ = await permissions?.prepare(.camera) } }
            .onChange(of: model.lastEvent) { _, event in if let event { feedback?.show(event, symbol: "doc.on.doc"); model.clearEvent() } }
            .permissionPresentation(
                permissions, for: .microphone,
                content: PermissionExplainerContent(symbol: "mic.fill", title: permissionCopy.microphoneTitle, message: permissionCopy.microphoneBody, deniedTitle: permissionCopy.microphoneDeniedTitle, deniedMessage: permissionCopy.microphoneDeniedBody, allow: permissionCopy.allow, notNow: permissionCopy.notNow, openSettings: permissionCopy.openSettings),
                fallback: (permissionCopy.typeInstead, { model.mode = .type })
            )
            .permissionPresentation(
                permissions, for: .camera,
                content: PermissionExplainerContent(symbol: "camera.fill", title: permissionCopy.cameraTitle, message: permissionCopy.cameraBody, deniedTitle: permissionCopy.cameraDeniedTitle, deniedMessage: permissionCopy.cameraDeniedBody, allow: permissionCopy.allow, notNow: permissionCopy.notNow, openSettings: permissionCopy.openSettings),
                fallback: (permissionCopy.typeInstead, { model.mode = .type })
            )
            .toolbar {
                ToolbarItem(placement: .primaryAction) {
                    HStack(spacing: metrics.points(6)) {
                        Circle().fill(Palette.aqua).frame(width: metrics.points(6), height: metrics.points(6))
                        Text(dayLabel).typeStyle(Typography.statusPill).foregroundStyle(Palette.matchTeal)
                    }
                    .padding(.horizontal, metrics.points(6))
                    .fixedSize()
                }
            }
            if model.isShowingLarge, let result = model.result {
                ShowToLocalView(translation: result, languageLabel: model.pair.target.uppercased(), returnLabel: copy.tapToReturn) { model.isShowingLarge = false }
                    .transition(.opacity)
            }
        }
        .animation(Motion.fade(0.3), value: model.isShowingLarge)
    }

    private func modeLabel(_ mode: CompanionMode) -> String {
        switch mode {
        case .speak: copy.speak
        case .type: copy.type
        case .camera: copy.camera
        case .phrases: copy.phrases
        }
    }

    @ViewBuilder
    private var modeContent: some View {
        switch model.mode {
        case .speak: SpeakPanel(model: model, onToggle: toggleListening)
        case .type: TypeCard(model: model)
        case .camera: CameraPanel(
                copy: copy,
                deniedTitle: permissions?.state(of: .camera).needsSettings == true ? permissionCopy.cameraDeniedTitle : nil,
                deniedMessage: permissionCopy.cameraDeniedBody, typeInstead: permissionCopy.typeInstead, onTypeInstead: { model.mode = .type }
            )
        case .phrases: EmptyView()
        }
    }

    private var phraseList: some View {
        VStack(alignment: .leading, spacing: metrics.points(11)) {
            HStack(alignment: .firstTextBaseline) {
                Text(model.mode == .phrases ? copy.phrases : copy.usefulNow).typeStyle(TypeStyle(size: 18, weight: .bold, tracking: -0.2)).foregroundStyle(Palette.label)
                Spacer()
                Text(copy.phrasesContext).typeStyle(TypeStyle(size: 12.5, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.55))
            }
            VStack(spacing: 0) {
                ForEach(model.phrases) { phrase in
                    Button { model.show(phrase) } label: {
                        HStack(spacing: metrics.points(12)) {
                            PhotoStandIn(palette: .named(phrase.placeholder), label: "", labelSize: 6)
                                .frame(width: metrics.points(30), height: metrics.points(30))
                                .clipShape(RoundedRectangle(cornerRadius: metrics.points(10), style: .continuous))
                            VStack(alignment: .leading, spacing: metrics.points(2)) {
                                Text(phrase.context.uppercased()).typeStyle(TypeStyle(size: 11, weight: .bold, tracking: 0.66)).foregroundStyle(Palette.labelAlpha(0.5))
                                Text("“\(phrase.original)”").typeStyle(TypeStyle(size: 15, weight: .semibold)).foregroundStyle(Palette.label)
                            }
                            Spacer(minLength: 0)
                            Image(systemName: "chevron.forward").font(.system(size: metrics.points(12), weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.3))
                        }
                        .padding(.horizontal, metrics.points(14)).padding(.vertical, metrics.points(12))
                        .contentShape(Rectangle())
                    }
                    .buttonStyle(.plain)
                    if phrase.id != model.phrases.last?.id { Divider().overlay(Palette.labelAlpha(0.07)) }
                }
            }
            .surfaceCard()
        }
    }
}

struct OfflineMark: View {
    @Environment(\.designMetrics) private var metrics

    var body: some View {
        Image(systemName: "arrow.down")
            .font(.system(size: metrics.points(9), weight: .bold))
            .foregroundStyle(Palette.matchTeal)
            .frame(width: metrics.points(16), height: metrics.points(16))
            .background(Palette.aqua.opacity(0.3), in: RoundedRectangle(cornerRadius: metrics.points(5), style: .continuous))
    }
}
