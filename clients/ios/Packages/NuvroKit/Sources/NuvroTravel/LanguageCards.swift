import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import SwiftUI

struct SpeakPanel: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @State private var rippling = false
    let model: LanguageCompanionModel
    let onToggle: () -> Void

    var body: some View {
        VStack(spacing: metrics.points(16)) {
            VStack(spacing: metrics.points(12)) {
                Button(action: onToggle) {
                    ZStack {
                        Circle().fill(Palette.sky.opacity(0.25)).scaleEffect(rippling ? 1.9 : 1).opacity(rippling ? 0 : 0.5)
                        Circle().fill(Palette.sky.opacity(model.isListening ? 0.98 : 0.88))
                            .overlay(Circle().strokeBorder(Color.white.opacity(0.5), lineWidth: 0.5))
                            .shadow(color: Palette.sky.opacity(0.16), radius: metrics.points(8), y: metrics.points(4))
                        Image(systemName: model.isListening ? "waveform" : "mic.fill").font(.system(size: metrics.points(30), weight: .semibold)).foregroundStyle(.white)
                    }
                    .frame(width: metrics.points(96), height: metrics.points(96))
                }
                .buttonStyle(PressScaleButtonStyle(scale: 0.96))
                .accessibilityLabel(model.copy.saySomething)
                .onAppear { if !reduceMotion { withAnimation(.easeOut(duration: 2).repeatForever(autoreverses: false)) { rippling = true } } }
                VStack(spacing: metrics.points(4)) {
                    Text(model.isListening ? model.copy.listening : model.copy.saySomething).typeStyle(TypeStyle(size: 20, weight: .bold, tracking: -0.3)).foregroundStyle(Palette.label)
                    Text(model.isListening ? model.copy.listeningSub : "\(model.copy.willTranslate) \(model.targetLanguage == "Português" ? "Portuguese" : model.targetLanguage).")
                        .typeStyle(TypeStyle(size: 14, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.62)).multilineTextAlignment(.center)
                }
            }
            .frame(maxWidth: .infinity)
            .padding(.top, metrics.points(14)).padding(.bottom, metrics.points(4))
            if let result = model.result {
                TranslationResultCard(translation: result, model: model)
            }
        }
    }
}

struct TypeCard: View {
    @Environment(\.designMetrics) private var metrics
    @Bindable var model: LanguageCompanionModel

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(12)) {
            TextField(model.copy.typePlaceholder, text: $model.typedText)
                .typeStyle(TypeStyle(size: 17, weight: .regular))
                .foregroundStyle(Palette.label)
                .onSubmit(model.translateTyped)
            Divider().overlay(Palette.labelAlpha(0.1))
            Text(model.result.map { "Tradução: \($0.translated)" } ?? model.copy.translationPlaceholder)
                .typeStyle(TypeStyle(size: 20, weight: .semibold))
                .foregroundStyle(model.result == nil ? Palette.labelAlpha(0.35) : Palette.label)
            Button { if model.result == nil { model.translateTyped() }; if model.result != nil { model.isShowingLarge = true } } label: {
                Text(model.copy.showLarge).typeStyle(TypeStyle(size: 14, weight: .semibold)).foregroundStyle(.white)
                    .padding(.horizontal, metrics.points(16)).frame(height: metrics.points(38))
                    .background(Palette.sky.opacity(0.88), in: Capsule())
            }
            .buttonStyle(PressScaleButtonStyle(scale: 0.97))
        }
        .padding(metrics.points(16))
        .surfaceCard(radius: 22)
    }
}

struct CameraPanel: View {
    @Environment(\.designMetrics) private var metrics
    let copy: LanguageCopy
    var deniedTitle: String? = nil
    var deniedMessage: String = ""
    var typeInstead: String = ""
    var onTypeInstead: () -> Void = {}

    var body: some View {
        if let deniedTitle {
            NuvroPermissionExplainer(symbol: "camera.fill", palette: .gold, title: deniedTitle, message: deniedMessage, primary: (typeInstead, onTypeInstead))
        } else {
            cameraEntry
        }
    }

    private var cameraEntry: some View {
        PhotoStandIn(palette: .gold, label: copy.cameraBody, labelAlignment: .topLeading)
            .frame(height: metrics.points(180))
            .overlay {
                RoundedRectangle(cornerRadius: metrics.points(12), style: .continuous)
                    .strokeBorder(style: StrokeStyle(lineWidth: 2, dash: [6, 4]))
                    .foregroundStyle(Color.white.opacity(0.8))
                    .frame(width: metrics.points(120), height: metrics.points(84))
            }
            .overlay(alignment: .bottom) {
                Text(copy.cameraEntry).typeStyle(TypeStyle(size: 15, weight: .semibold)).foregroundStyle(Palette.cloud)
                    .frame(maxWidth: .infinity).frame(height: metrics.points(46))
                    .background(Palette.navy.opacity(0.45), in: Capsule())
                    .overlay(Capsule().strokeBorder(Color.white.opacity(0.35), lineWidth: 0.5))
                    .padding(.horizontal, metrics.points(16)).padding(.bottom, metrics.points(14))
            }
            .clipShape(RoundedRectangle(cornerRadius: metrics.points(22), style: .continuous))
    }
}

struct TranslationResultCard: View {
    @Environment(\.designMetrics) private var metrics
    @State private var speaker = PhraseSpeaker()
    let translation: PhraseTranslation
    let model: LanguageCompanionModel

    var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(12)) {
            VStack(alignment: .leading, spacing: metrics.points(3)) {
                Text(translation.pair.source.uppercased()).typeStyle(TypeStyle(size: 11, weight: .bold, tracking: 1.3)).foregroundStyle(Palette.labelAlpha(0.5))
                Text("“\(translation.original)”").typeStyle(TypeStyle(size: 15, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.7))
            }
            Divider().overlay(Palette.labelAlpha(0.1))
            VStack(alignment: .leading, spacing: metrics.points(4)) {
                Text(translation.pair.target.uppercased()).typeStyle(TypeStyle(size: 11, weight: .bold, tracking: 1.3)).foregroundStyle(Palette.sky)
                Text(translation.translated).typeStyle(TypeStyle(size: 21, weight: .semibold, tracking: -0.3)).foregroundStyle(Palette.label)
            }
            HStack(spacing: metrics.points(8)) {
                Button { speaker.speak(translation.translated, languageName: translation.pair.target) } label: {
                    HStack(spacing: metrics.points(7)) {
                        Image(systemName: "play.fill").font(.system(size: metrics.points(11))).foregroundStyle(Palette.sky)
                        Text(model.copy.play).typeStyle(TypeStyle(size: 14, weight: .semibold)).foregroundStyle(Palette.label)
                    }
                    .frame(maxWidth: .infinity).frame(height: metrics.points(42)).glassCapsule(.light)
                }
                .buttonStyle(PressScaleButtonStyle(scale: 0.97))
                Button { model.isShowingLarge = true } label: {
                    Text(model.copy.showLarge).typeStyle(TypeStyle(size: 14.5, weight: .semibold)).foregroundStyle(.white)
                        .frame(maxWidth: .infinity).frame(height: metrics.points(42))
                }
                .buttonStyle(GlassCallToActionStyle())
                .layoutPriority(1)
                Button { copyToPasteboard(model.copyTranslation()) } label: {
                    Image(systemName: "doc.on.doc").font(.system(size: metrics.points(13), weight: .semibold)).foregroundStyle(Palette.label)
                        .frame(width: metrics.points(46), height: metrics.points(42)).glassCapsule(.light)
                }
                .buttonStyle(PressScaleButtonStyle(scale: 0.97))
                .accessibilityLabel(model.copy.copy)
            }
            .padding(.top, metrics.points(2))
        }
        .padding(metrics.points(16))
        .surfaceCard(radius: 22)
        .transition(.opacity.combined(with: .offset(y: 14)))
    }
}

private func copyToPasteboard(_ text: String?) {
    #if canImport(UIKit)
    guard let text else { return }
    UIPasteboard.general.string = text
    #endif
}

struct ShowToLocalView: View {
    @Environment(\.designMetrics) private var metrics
    let translation: PhraseTranslation
    let languageLabel: String
    let returnLabel: String
    let dismiss: () -> Void

    var body: some View {
        Button(action: dismiss) {
            VStack(spacing: metrics.points(28)) {
                Spacer()
                Text(languageLabel).typeStyle(TypeStyle(size: 11, weight: .bold, tracking: 1.5)).foregroundStyle(Palette.labelAlpha(0.45))
                Text(translation.translated)
                    .typeStyle(TypeStyle(size: 34, weight: .semibold, tracking: -0.6))
                    .foregroundStyle(Palette.label)
                    .multilineTextAlignment(.center)
                    .padding(.horizontal, metrics.points(28))
                Text("“\(translation.original)”").typeStyle(Typography.body).foregroundStyle(Palette.labelAlpha(0.5))
                Spacer()
                Text(returnLabel).typeStyle(Typography.metaEmphasis).foregroundStyle(Palette.labelAlpha(0.45)).padding(.bottom, metrics.points(40))
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
            .background(Palette.surface.ignoresSafeArea())
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .accessibilityLabel("\(translation.translated). \(returnLabel)")
    }
}
