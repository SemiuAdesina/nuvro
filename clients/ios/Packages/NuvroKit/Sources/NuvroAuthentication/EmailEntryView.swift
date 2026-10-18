import NuvroDesignSystem
import NuvroLocalization
import SwiftUI

public struct EmailEntryView: View {
    @Environment(\.designMetrics) private var metrics
    @FocusState private var isFocused: Bool
    @Bindable private var model: EmailEntryModel
    private let copy: AuthenticationCopy
    private let onCodeSent: (EmailCodeRequest) -> Void

    public init(model: EmailEntryModel, copy: AuthenticationCopy, onCodeSent: @escaping (EmailCodeRequest) -> Void) {
        self.model = model
        self.copy = copy
        self.onCodeSent = onCodeSent
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(18)) {
            VStack(alignment: .leading, spacing: metrics.points(6)) {
                Text(copy.emailTitle).typeStyle(Typography.headline).foregroundStyle(Palette.cloud)
                Text(copy.emailBody).typeStyle(Typography.body).foregroundStyle(Palette.cloudAlpha(0.75))
            }
            AuthenticationField(isInvalid: model.showsFormatError) {
                TextField(copy.emailPlaceholder, text: $model.email, prompt: Text(copy.emailPlaceholder).foregroundStyle(Palette.cloudAlpha(0.45)))
                    .typeStyle(Typography.fieldValue)
                    .foregroundStyle(Palette.cloud)
                    .textContentType(.emailAddress)
                    .autocorrectionDisabled()
                    .submitLabel(.continue)
                    .focused($isFocused)
                    .onSubmit { Task { await submit() } }
                    .onChange(of: isFocused) { _, focused in if !focused { model.finishEditing() } }
                    .accessibilityLabel(copy.emailPlaceholder)
                    #if os(iOS)
                    .keyboardType(.emailAddress)
                    .textInputAutocapitalization(.never)
                    #endif
            }
            if model.showsFormatError {
                Text(copy.invalidEmail).typeStyle(Typography.caption).foregroundStyle(Palette.gold)
                    .accessibilityLabel(copy.invalidEmail)
            }
            if let failure = model.rejection {
                AuthenticationFailureCard(failure: failure, copy: copy, retryLabel: copy.tryAgain) { Task { await submit() } }
            }
            Spacer()
            CallToActionButton(copy.continueLabel, isLoading: model.isSubmitting) { Task { await submit() } }
                .disabled(!model.canSubmit)
                .opacity(model.canSubmit || model.isSubmitting ? 1 : 0.55)
        }
        .padding(.horizontal, metrics.points(24))
        .padding(.top, metrics.points(20))
        .padding(.bottom, metrics.points(24))
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        .background(AuthenticationBackdrop())
        .onAppear { isFocused = true }
    }

    private func submit() async {
        if let request = await model.submit() { onCodeSent(request) }
    }
}
