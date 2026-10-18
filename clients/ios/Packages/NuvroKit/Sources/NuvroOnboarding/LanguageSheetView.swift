import NuvroDesignSystem
import NuvroLocalization
import SwiftUI

struct LanguageSheetView: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let searchPlaceholder: String
    let selected: Language
    let languages: [Language]
    @Binding var query: String
    let onSelect: (Language) -> Void
    let onClose: () -> Void

    var body: some View {
        VStack(spacing: 0) {
            HStack {
                Text(title).typeStyle(Typography.sheetTitle).foregroundStyle(Palette.label)
                Spacer()
                Button(action: onClose) {
                    Image(systemName: "xmark")
                        .font(.system(size: metrics.points(14), weight: .bold))
                        .foregroundStyle(Palette.label)
                        .frame(width: metrics.points(32), height: metrics.points(32))
                        .background(Palette.labelAlpha(0.08), in: Circle())
                }
                .buttonStyle(.plain)
                .accessibilityLabel("Close")
            }
            .padding(.horizontal, metrics.points(20))
            .padding(.top, metrics.points(8))
            .padding(.bottom, metrics.points(12))

            HStack(spacing: metrics.points(8)) {
                Image(systemName: "magnifyingglass")
                    .font(.system(size: metrics.points(14), weight: .semibold))
                    .foregroundStyle(Palette.labelAlpha(0.45))
                TextField(searchPlaceholder, text: $query)
                    .typeStyle(Typography.listRow)
                    .foregroundStyle(Palette.label)
                    .textFieldStyle(.plain)
                    .autocorrectionDisabled()
            }
            .padding(.horizontal, metrics.points(12))
            .frame(height: metrics.points(40))
            .background(Palette.labelAlpha(0.07), in: RoundedRectangle(cornerRadius: metrics.points(Radius.field), style: .continuous))
            .padding(.horizontal, metrics.points(20))
            .padding(.bottom, metrics.points(10))

            ScrollView {
                VStack(spacing: 0) {
                    ForEach(languages) { language in
                        row(language)
                        if language != languages.last {
                            Divider().overlay(Palette.labelAlpha(0.08))
                        }
                    }
                }
                .background(Color.white.opacity(0.75), in: RoundedRectangle(cornerRadius: metrics.points(Radius.list), style: .continuous))
                .padding(.horizontal, metrics.points(20))
                .padding(.bottom, metrics.points(40))
            }
        }
        .environment(\.colorScheme, .light)
        .presentationDetents([.height(metrics.points(560))])
        .presentationDragIndicator(.visible)
        .presentationCornerRadius(metrics.points(Radius.sheet))
        .presentationBackground {
            Color.clear
                .glass(.sheet, in: Rectangle())
                .environment(\.colorScheme, .light)
        }
    }

    private func row(_ language: Language) -> some View {
        Button { onSelect(language) } label: {
            HStack {
                Text(language.nativeName)
                    .typeStyle(language == selected ? TypeStyle(size: 16, weight: .bold) : Typography.listRow)
                    .foregroundStyle(Palette.label)
                Spacer()
                Image(systemName: "checkmark")
                    .font(.system(size: metrics.points(14), weight: .bold))
                    .foregroundStyle(Palette.sky)
                    .opacity(language == selected ? 1 : 0)
            }
            .padding(.horizontal, metrics.points(16))
            .frame(height: metrics.points(48))
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .accessibilityAddTraits(language == selected ? .isSelected : [])
    }
}
