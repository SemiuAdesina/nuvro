import Foundation

@testable import NuvroAuthentication

final class ScriptedAuthenticationClient: AuthenticationClient, @unchecked Sendable {
    var appleResult: Result<AuthSession, AuthenticationFailure> = .success(.fixture)
    var codeRequestResult: Result<EmailCodeRequest, AuthenticationFailure> = .success(.fixture)
    var verifyResult: Result<AuthSession, AuthenticationFailure> = .success(.fixture)
    var refreshResult: Result<AuthSession, AuthenticationFailure> = .success(.renewed)
    var signOutResult: Result<Void, AuthenticationFailure> = .success(())
    var deleteResult: Result<Void, AuthenticationFailure> = .success(())
    private(set) var codeRequests: [String] = []
    private(set) var verifications: [(email: String, code: String)] = []
    private(set) var signOutCount = 0
    private(set) var deleteCount = 0
    private(set) var refreshCount = 0

    func signInWithApple(_ credential: AppleCredential) async throws(AuthenticationFailure) -> AuthSession { try appleResult.get() }

    func requestEmailCode(email: String) async throws(AuthenticationFailure) -> EmailCodeRequest {
        codeRequests.append(email)
        return try codeRequestResult.get()
    }

    func verifyEmailCode(email: String, code: String) async throws(AuthenticationFailure) -> AuthSession {
        verifications.append((email, code))
        return try verifyResult.get()
    }

    func refresh(_ session: AuthSession) async throws(AuthenticationFailure) -> AuthSession {
        refreshCount += 1
        return try refreshResult.get()
    }

    func signOut(_ session: AuthSession) async throws(AuthenticationFailure) {
        signOutCount += 1
        try signOutResult.get()
    }

    func deleteAccount(_ session: AuthSession) async throws(AuthenticationFailure) {
        deleteCount += 1
        try deleteResult.get()
    }
}

extension AuthSession {
    static let fixture = AuthSession(accessToken: "access-1", refreshToken: "refresh-1", expiresAt: Date(timeIntervalSince1970: 2_000_000_000), accountID: "acct-1")
    static let renewed = AuthSession(accessToken: "access-2", refreshToken: "refresh-2", expiresAt: Date(timeIntervalSince1970: 2_100_000_000), accountID: "acct-1")
    static let stale = AuthSession(accessToken: "access-0", refreshToken: "refresh-0", expiresAt: Date(timeIntervalSince1970: 1_000), accountID: "acct-1")
}

extension EmailCodeRequest {
    static let fixture = EmailCodeRequest(email: "alex@example.com", maskedEmail: "a•••@example.com", codeLength: 6, resendAfterSeconds: 30)
}

extension Result where Failure == AuthenticationFailure {
    func get() throws(AuthenticationFailure) -> Success {
        switch self {
        case .success(let value): return value
        case .failure(let failure): throw failure
        }
    }
}
