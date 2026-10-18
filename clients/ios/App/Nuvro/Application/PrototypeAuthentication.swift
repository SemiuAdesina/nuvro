import Foundation
import NuvroAuthentication

struct PrototypeAuthenticationClient: AuthenticationClient {
    private static let codeLength = 6
    private static let resendAfterSeconds = 30
    private static let rejectedCode = "000000"

    private func session(for accountID: String) -> AuthSession {
        AuthSession(accessToken: "prototype-access-\(UUID().uuidString)", refreshToken: "prototype-refresh", expiresAt: Date().addingTimeInterval(3600), accountID: accountID)
    }

    func signInWithApple(_ credential: AppleCredential) async throws(AuthenticationFailure) -> AuthSession {
        session(for: "apple-\(credential.userIdentifier)")
    }

    func requestEmailCode(email: String) async throws(AuthenticationFailure) -> EmailCodeRequest {
        EmailCodeRequest(email: email, maskedEmail: EmailMasking.masked(email), codeLength: Self.codeLength, resendAfterSeconds: Self.resendAfterSeconds)
    }

    func verifyEmailCode(email: String, code: String) async throws(AuthenticationFailure) -> AuthSession {
        guard code != Self.rejectedCode else { throw .invalidCode }
        return session(for: "email-\(email)")
    }

    func refresh(_ session: AuthSession) async throws(AuthenticationFailure) -> AuthSession {
        self.session(for: session.accountID)
    }

    func signOut(_ session: AuthSession) async throws(AuthenticationFailure) {}

    func deleteAccount(_ session: AuthSession) async throws(AuthenticationFailure) {}
}
