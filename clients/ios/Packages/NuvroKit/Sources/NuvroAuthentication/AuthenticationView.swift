import NuvroDesignSystem
import NuvroLocalization
import SwiftUI
#if canImport(AuthenticationServices)
import AuthenticationServices
#endif

public enum AuthenticationRoute: Hashable, Sendable {
    case emailEntry
    case verification(EmailCodeRequest)
}

public struct AuthenticationView: View {
    @Environment(\.designMetrics) private var metrics
    @State private var path: [AuthenticationRoute] = []
    private let coordinator: AuthenticationCoordinator
    private let copy: AuthenticationCopy
    private let onOpenLegal: (LegalDocument) -> Void

    public init(coordinator: AuthenticationCoordinator, language: Language = .english, onOpenLegal: @escaping (LegalDocument) -> Void = { _ in }) {
        self.coordinator = coordinator
        self.copy = .resolve(language)
        self.onOpenLegal = onOpenLegal
    }

    public var body: some View {
        NavigationStack(path: $path) {
            landing
                .navigationDestination(for: AuthenticationRoute.self) { route in
                    switch route {
                    case .emailEntry:
                        EmailEntryView(model: EmailEntryModel(client: coordinator.client), copy: copy) { path.append(.verification($0)) }
                    case .verification(let request):
                        EmailVerificationView(model: EmailVerificationModel(request: request, client: coordinator.client), copy: copy, onVerified: coordinator.establish, onChangeEmail: { path.removeLast() })
                    }
                }
        }
        .tint(Palette.cloud)
    }

    private var landing: some View {
        VStack(spacing: 0) {
            Spacer()
            Wordmark(size: metrics.points(46))
                .accessibilityLabel("NUVRO")
            Text(copy.tagline)
                .typeStyle(Typography.supporting)
                .foregroundStyle(Palette.cloudAlpha(0.85))
                .padding(.top, metrics.points(14))
            Spacer()
            notice
            VStack(spacing: metrics.points(14)) {
                appleButton
                HStack(spacing: metrics.points(12)) {
                    Rectangle().fill(Palette.cloudAlpha(0.25)).frame(height: 0.5)
                    Text(copy.or).typeStyle(Typography.caption).foregroundStyle(Palette.cloudAlpha(0.6))
                    Rectangle().fill(Palette.cloudAlpha(0.25)).frame(height: 0.5)
                }
                .accessibilityHidden(true)
                Button { path.append(.emailEntry) } label: {
                    Text(copy.continueWithEmail)
                        .typeStyle(Typography.callToAction)
                        .foregroundStyle(Palette.cloud)
                        .frame(maxWidth: .infinity)
                        .frame(height: metrics.points(54))
                }
                .buttonStyle(GlassPillButtonStyle(style: .control))
                .disabled(coordinator.state == .authenticating)
                LegalLine(copy: copy, onOpen: onOpenLegal)
                    .padding(.top, metrics.points(4))
            }
            .padding(.horizontal, metrics.points(24))
            .padding(.bottom, metrics.points(30))
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(AuthenticationBackdrop())
        .hiddenNavigationBar()
    }

    @ViewBuilder
    private var notice: some View {
        if let text = noticeText {
            Text(text)
                .typeStyle(Typography.bodyCompact)
                .foregroundStyle(Palette.cloudAlpha(0.8))
                .multilineTextAlignment(.center)
                .padding(.horizontal, metrics.points(14)).padding(.vertical, metrics.points(11))
                .glass(.notice, in: RoundedRectangle(cornerRadius: metrics.points(Radius.notice), style: .continuous))
                .padding(.horizontal, metrics.points(24))
                .padding(.bottom, metrics.points(18))
                .transition(.opacity)
        }
        if let failure = coordinator.lastFailure {
            AuthenticationFailureCard(failure: failure, copy: copy, retryLabel: copy.tryAgain, isAppleFailure: true) { coordinator.clearFailure() }
                .padding(.horizontal, metrics.points(24))
                .padding(.bottom, metrics.points(18))
        }
    }

    private var noticeText: String? {
        switch coordinator.state {
        case .expired: copy.sessionExpired
        case .signedOut(.signedOut): copy.signedOut
        case .signedOut(.deleted): copy.accountDeleted
        default: nil
        }
    }

    @ViewBuilder
    private var appleButton: some View {
        #if canImport(AuthenticationServices)
        SignInWithAppleButton(.continue) { request in
            request.requestedScopes = [.fullName, .email]
        } onCompletion: { result in
            let outcome = AppleAuthorizationMapper.outcome(from: result)
            Task { await coordinator.complete(outcome) }
        }
        .signInWithAppleButtonStyle(.white)
        .frame(height: metrics.points(54))
        .clipShape(Capsule())
        .disabled(coordinator.state == .authenticating)
        .overlay {
            if coordinator.state == .authenticating {
                ProgressView().tint(Palette.navy)
            }
        }
        #else
        EmptyView()
        #endif
    }
}
