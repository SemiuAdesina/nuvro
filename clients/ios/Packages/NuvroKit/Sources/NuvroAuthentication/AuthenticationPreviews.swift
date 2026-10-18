import NuvroDesignSystem
import SwiftUI

private struct FailingPreviewClient: AuthenticationClient {
    let failure: AuthenticationFailure

    func signInWithApple(_ credential: AppleCredential) async throws(AuthenticationFailure) -> AuthSession { throw failure }
    func requestEmailCode(email: String) async throws(AuthenticationFailure) -> EmailCodeRequest { throw failure }
    func verifyEmailCode(email: String, code: String) async throws(AuthenticationFailure) -> AuthSession { throw failure }
    func refresh(_ session: AuthSession) async throws(AuthenticationFailure) -> AuthSession { throw failure }
    func signOut(_ session: AuthSession) async throws(AuthenticationFailure) {}
    func deleteAccount(_ session: AuthSession) async throws(AuthenticationFailure) {}
}

#Preview("Landing") {
    DesignCanvas { _ in
        AuthenticationView(coordinator: AuthenticationCoordinator(client: PendingAuthenticationClient(), store: InMemorySessionStore()))
    }
    .preferredColorScheme(.dark)
}

#Preview("Landing with failure") {
    let coordinator = AuthenticationCoordinator(client: FailingPreviewClient(failure: .unavailable), store: InMemorySessionStore())
    return DesignCanvas { _ in AuthenticationView(coordinator: coordinator) }
        .task { await coordinator.complete(.authorized(AppleCredential(userIdentifier: "preview", identityToken: "t", authorizationCode: "c"))) }
        .preferredColorScheme(.dark)
}

#Preview("Email entry, offline") {
    DesignCanvas { _ in
        EmailEntryView(model: EmailEntryModel(client: FailingPreviewClient(failure: .network)), copy: .english) { _ in }
    }
    .preferredColorScheme(.dark)
}

#Preview("Verification, accessibility text size") {
    DesignCanvas { _ in
        EmailVerificationView(
            model: EmailVerificationModel(request: EmailCodeRequest(email: "alex@example.com", maskedEmail: "a•••@example.com", codeLength: 6, resendAfterSeconds: 30), client: FailingPreviewClient(failure: .invalidCode)),
            copy: .english, onVerified: { _ in }, onChangeEmail: {}
        )
    }
    .dynamicTypeSize(.accessibility2)
    .preferredColorScheme(.dark)
}

#Preview("Bootstrap") {
    DesignCanvas { _ in AuthenticationBootstrapView() }.preferredColorScheme(.dark)
}
