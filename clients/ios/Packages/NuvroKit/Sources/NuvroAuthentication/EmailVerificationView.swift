import NuvroDesignSystem
import NuvroLocalization
import SwiftUI

public struct EmailVerificationView: View {
    @Environment(\.designMetrics) private var metrics
    @FocusState private var isFocused: Bool
    @State private var typed = ""
    @State private var clock = Date()
    private let model: EmailVerificationModel
    private let copy: AuthenticationCopy
    private let onVerified: (AuthSession) -> Void
    private let onChangeEmail: () -> Void

    public init(model: EmailVerificationModel, copy: AuthenticationCopy, onVerified: @escaping (AuthSession) -> Void, onChangeEmail: @escaping () -> Void) {
        self.model = model
        self.copy = copy
        self.onVerified = onVerified
        self.onChangeEmail = onChangeEmail
    }

    public var body: some View {
        VStack(alignment: .leading, spacing: metrics.points(18)) {
            VStack(alignment: .leading, spacing: metrics.points(6)) {
                Text(copy.checkEmail).typeStyle(Typography.headline).foregroundStyle(Palette.cloud)
                Text("\(copy.enterCodeSentTo) \(model.maskedEmail)").typeStyle(Typography.body).foregroundStyle(Palette.cloudAlpha(0.75))
            }
            codeEntry
            if let failure = model.failure {
                AuthenticationFailureCard(failure: failure, copy: copy, retryLabel: copy.resend) { Task { await model.resend() } }
            }
            HStack {
                resendControl
                Spacer()
                Button(copy.changeEmail, action: onChangeEmail).typeStyle(Typography.link).foregroundStyle(Palette.cloudAlpha(0.85)).buttonStyle(.plain)
            }
            Spacer()
            CallToActionButton(copy.continueLabel, isLoading: model.phase == .verifying) { Task { await model.verify() } }
                .disabled(!model.isComplete || model.isBusy)
                .opacity(model.isComplete || model.isBusy ? 1 : 0.55)
        }
        .padding(.horizontal, metrics.points(24))
        .padding(.top, metrics.points(20))
        .padding(.bottom, metrics.points(24))
        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
        .background(AuthenticationBackdrop())
        .onAppear { isFocused = true }
        .onChange(of: typed) { _, value in
            model.enter(value)
            if typed != model.code { typed = model.code }
            if model.isComplete { Task { await model.verify() } }
        }
        .onChange(of: model.code) { _, value in if typed != value { typed = value } }
        .onChange(of: model.session) { _, session in if let session { onVerified(session) } }
        .task {
            while !Task.isCancelled {
                clock = Date()
                try? await Task.sleep(for: .seconds(1))
            }
        }
    }

    private var codeEntry: some View {
        ZStack {
            TextField(copy.codeLabel, text: $typed)
                .textContentType(.oneTimeCode)
                .focused($isFocused)
                .opacity(0.02)
                .accessibilityLabel(copy.codeLabel)
                .accessibilityValue(typed.isEmpty ? "" : typed.map(String.init).joined(separator: " "))
                #if os(iOS)
                .keyboardType(.numberPad)
                #endif
            HStack(spacing: metrics.points(8)) {
                ForEach(0..<model.codeLength, id: \.self) { index in
                    let digit = index < model.code.count ? String(model.code[model.code.index(model.code.startIndex, offsetBy: index)]) : ""
                    Text(digit)
                        .typeStyle(TypeStyle(size: 24, weight: .bold))
                        .foregroundStyle(Palette.cloud)
                        .frame(maxWidth: .infinity)
                        .frame(height: metrics.points(58))
                        .glass(.control, in: RoundedRectangle(cornerRadius: metrics.points(Radius.chip), style: .continuous))
                        .overlay {
                            if index == model.code.count, isFocused {
                                RoundedRectangle(cornerRadius: metrics.points(Radius.chip), style: .continuous).strokeBorder(Palette.gold, lineWidth: 1)
                            }
                        }
                }
            }
            .accessibilityHidden(true)
            .allowsHitTesting(false)
        }
        .contentShape(Rectangle())
        .onTapGesture { isFocused = true }
    }

    @ViewBuilder
    private var resendControl: some View {
        if model.canResend(at: clock) {
            Button(copy.resend) { Task { await model.resend() } }.typeStyle(Typography.link).foregroundStyle(Palette.gold).buttonStyle(.plain)
        } else {
            Text("\(copy.resendIn) \(model.resendCountdown(at: clock))s").typeStyle(Typography.link).foregroundStyle(Palette.cloudAlpha(0.5))
        }
    }
}
