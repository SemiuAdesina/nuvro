import NuvroDesignSystem
import NuvroModels
import SwiftUI

struct LearningPreferenceCard: View {
    @Environment(\.designMetrics) private var metrics
    @Environment(\.dynamicTypeSize) private var typeSize
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @State private var hasAppeared = false
    @Bindable var model: ProfileModel
    let insights: [LearnedInsight]

    static let assetName = "profile-learning-attendant"

    static func showsIllustration(for size: DynamicTypeSize) -> Bool {
        !size.isAccessibilitySize
    }

    private var showsIllustration: Bool { Self.showsIllustration(for: typeSize) }
    private var shape: RoundedRectangle { RoundedRectangle(cornerRadius: metrics.points(Radius.card), style: .continuous) }

    var body: some View {
        content
            .frame(maxWidth: .infinity, alignment: .leading)
            .frame(minHeight: showsIllustration ? metrics.points(148) : nil)
            .background(alignment: .topTrailing) { if showsIllustration { illustration } }
            .background(
                LinearGradient(colors: [Palette.sky.opacity(0.08), Palette.aqua.opacity(0.10)], startPoint: .topLeading, endPoint: .bottomTrailing),
                in: shape
            )
            .background(Palette.surface, in: shape)
            .clipShape(shape)
            .overlay(shape.strokeBorder(Palette.sky.opacity(0.18), lineWidth: 0.5))
    }

    private var content: some View {
        VStack(alignment: .leading, spacing: metrics.points(10)) {
            Text(model.copy.learned).typeStyle(TypeStyle(size: 15, weight: .bold)).foregroundStyle(Palette.label)
                .padding(.trailing, showsIllustration ? metrics.points(126) : 0)
            ForEach(insights) { insight in
                HStack(spacing: metrics.points(10)) {
                    Circle().fill(Palette.aqua).frame(width: metrics.points(7), height: metrics.points(7))
                    Text(insight.text).typeStyle(TypeStyle(size: 13.5, weight: .regular)).foregroundStyle(Palette.label)
                    Spacer()
                    Text(insight.source).typeStyle(TypeStyle(size: 12, weight: .semibold)).foregroundStyle(Palette.labelAlpha(0.45))
                }
                .padding(.horizontal, metrics.points(12)).padding(.vertical, metrics.points(11))
                .background(Palette.surface.opacity(0.85), in: RoundedRectangle(cornerRadius: metrics.points(14), style: .continuous))
                .padding(.trailing, showsIllustration ? metrics.points(126) : 0)
            }
            Spacer(minLength: metrics.points(6))
            Toggle(isOn: $model.usesTripHistory) {
                VStack(alignment: .leading, spacing: 1) {
                    Text(model.copy.useHistory).typeStyle(TypeStyle(size: 14, weight: .semibold)).foregroundStyle(Palette.label)
                    Text(model.historyStateLabel).typeStyle(TypeStyle(size: 12, weight: .regular)).foregroundStyle(Palette.labelAlpha(0.6))
                }
                .padding(.trailing, showsIllustration ? metrics.points(80) : 0)
            }
            .tint(Palette.sky)
            .padding(.top, metrics.points(2))
        }
        .padding(metrics.points(16))
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
    }

    private var illustration: some View {
        Image(Self.assetName, bundle: .module)
            .resizable()
            .scaledToFit()
            .frame(width: metrics.points(124))
            .mask(LinearGradient(stops: [.init(color: .clear, location: 0), .init(color: .black, location: 0.34)], startPoint: .leading, endPoint: .trailing))
            .padding(.trailing, metrics.points(74))
            .offset(y: metrics.points(22))
            .opacity(hasAppeared ? 1 : 0)
            .offset(x: hasAppeared ? 0 : metrics.points(6))
            .allowsHitTesting(false)
            .accessibilityHidden(true)
            .onAppear {
                if reduceMotion { hasAppeared = true } else { withAnimation(.easeOut(duration: 0.4)) { hasAppeared = true } }
            }
    }
}
