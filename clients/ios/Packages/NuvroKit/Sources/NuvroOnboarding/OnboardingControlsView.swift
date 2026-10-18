import NuvroDesignSystem
import NuvroLocalization
import SwiftUI

struct TopControlsView: View {
    @Environment(\.designMetrics) private var metrics
    let language: Language
    let skipLabel: String
    let isLanguageShown: Bool
    let isSkipShown: Bool
    let onLanguage: () -> Void
    let onSkip: () -> Void

    var body: some View {
        GlassGroup(spacing: metrics.points(12)) {
            HStack {
                if isLanguageShown {
                    languagePill.transition(.opacity)
                }
                Spacer()
                if isSkipShown {
                    skipPill.transition(.opacity)
                }
            }
        }
        .padding(.horizontal, metrics.points(20))
        .frame(width: metrics.size.width, height: metrics.points(38))
        .offset(y: metrics.belowSafeTop(66))
        .animation(Motion.fade(Motion.controlFadeDuration), value: isSkipShown)
        .animation(Motion.fade(Motion.controlFadeDuration), value: isLanguageShown)
    }

    private var languagePill: some View {
        Button(action: onLanguage) {
            HStack(spacing: metrics.points(6)) {
                Image(systemName: "globe").font(.system(size: metrics.points(14)))
                Text(language.nativeName)
                Image(systemName: "chevron.down").font(.system(size: metrics.points(10), weight: .semibold)).opacity(0.7)
            }
            .typeStyle(Typography.control)
            .foregroundStyle(Palette.cloud)
            .padding(.leading, metrics.points(12))
            .padding(.trailing, metrics.points(14))
            .frame(height: metrics.points(38))
        }
        .buttonStyle(GlassPillButtonStyle(style: .control))
        .accessibilityLabel("Language, \(language.nativeName)")
    }

    private var skipPill: some View {
        Button(action: onSkip) {
            Text(skipLabel)
                .typeStyle(Typography.control)
                .foregroundStyle(Palette.cloud)
                .padding(.horizontal, metrics.points(16))
                .frame(height: metrics.points(38))
        }
        .buttonStyle(GlassPillButtonStyle(style: .control))
    }
}

struct PrimaryActionsView: View {
    @Environment(\.designMetrics) private var metrics
    let callToAction: String
    let secondaryAction: String
    let isShown: Bool
    let isSecondaryShown: Bool
    let reduceMotion: Bool
    let onPrimary: () -> Void
    let onSecondary: () -> Void

    var body: some View {
        ZStack(alignment: .top) {
            Button(action: onPrimary) {
                Text(callToAction)
                    .typeStyle(Typography.callToAction)
                    .foregroundStyle(.white)
                    .frame(maxWidth: .infinity)
                    .frame(height: metrics.points(54))
            }
            .buttonStyle(GlassCallToActionStyle())
            .padding(.horizontal, metrics.points(24))
            .frame(width: metrics.size.width)
            .position(x: metrics.horizontalCenter, y: metrics.aboveSafeBottom(92) - metrics.points(27) + (isShown || reduceMotion ? 0 : metrics.points(16)))
            .opacity(isShown ? 1 : 0)
            .disabled(!isShown)
            .animation(reduceMotion ? nil : Motion.settle(Motion.textDuration), value: isShown)
            .animation(Motion.fade(Motion.textFadeDuration), value: isShown)

            Button(action: onSecondary) {
                Text(secondaryAction)
                    .typeStyle(Typography.secondaryAction)
                    .foregroundStyle(Palette.cloudAlpha(0.85))
                    .frame(maxWidth: .infinity)
                    .frame(height: metrics.points(36))
            }
            .buttonStyle(.plain)
            .padding(.horizontal, metrics.points(24))
            .frame(width: metrics.size.width)
            .position(x: metrics.horizontalCenter, y: metrics.aboveSafeBottom(50) - metrics.points(18))
            .opacity(isSecondaryShown ? 1 : 0)
            .disabled(!isSecondaryShown)
            .animation(Motion.fade(Motion.controlFadeDuration), value: isSecondaryShown)
        }
    }
}
