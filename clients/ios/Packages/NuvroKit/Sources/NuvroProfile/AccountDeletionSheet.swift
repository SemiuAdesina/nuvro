import NuvroDesignSystem
import SwiftUI

struct AccountDeletionSheet: View {
    @Environment(\.designMetrics) private var metrics
    @Bindable var model: SettingsModel

    var body: some View {
        NavigationStack {
            VStack(alignment: .leading, spacing: metrics.points(16)) {
                IconTile(palette: .gold, systemImage: "person.crop.circle.badge.xmark", size: 34)
                Text(model.copy.deleteExplainer).typeStyle(Typography.body).foregroundStyle(Palette.label)
                Text(model.copy.deleteKeeps).typeStyle(Typography.bodyCompact).foregroundStyle(Palette.labelAlpha(0.65))
                if model.deleteFailed {
                    NuvroInlineError(model.copy.deleteFailed)
                }
                Spacer()
                Button(role: .destructive) { model.proceedToDeleteConfirmation() } label: {
                    ZStack {
                        Text(model.copy.deleteAccount).typeStyle(Typography.callToAction).foregroundStyle(.white).opacity(model.isDeleting ? 0 : 1)
                        if model.isDeleting { ProgressView().tint(.white) }
                    }
                    .frame(maxWidth: .infinity).frame(height: metrics.points(50))
                }
                .buttonStyle(GlassPillButtonStyle(style: .tinted(Palette.destructive)))
                .disabled(model.isDeleting)
                .accessibilityValue(model.isDeleting ? model.copy.deleting : "")
            }
            .padding(metrics.points(24))
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
            .background(Palette.background)
            .navigationTitle(model.copy.deleteTitle)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button(model.copy.cancel) { model.dismissPresentation() }.disabled(model.isDeleting)
                }
            }
            .toolbarTitleDisplayMode(.inline)
            .alert(model.copy.deleteConfirmTitle, isPresented: Binding(get: { model.presentation == .deleteConfirmation }, set: { if !$0 && model.presentation == .deleteConfirmation { model.presentation = .deleteExplainer } })) {
                Button(model.copy.deleteAccount, role: .destructive) { Task { await model.confirmDelete() } }
                Button(model.copy.cancel, role: .cancel) { model.presentation = .deleteExplainer }
            } message: {
                Text(model.copy.deleteConfirmBody)
            }
        }
        .presentationDetents([.large])
        .interactiveDismissDisabled(model.isDeleting)
    }
}
