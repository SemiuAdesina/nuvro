import Foundation
import NuvroAPI
import NuvroModels

public struct LiveAuthenticationClient: AuthenticationClient {
    private let sessions: any SessionGateway
    private let account: any AccountGateway

    public init(sessions: any SessionGateway, account: any AccountGateway) {
        self.sessions = sessions
        self.account = account
    }

    public func signInWithApple(_ credential: AppleCredential) async throws(AuthenticationFailure) -> AuthSession {
        let code = credential.authorizationCode.isEmpty ? nil : credential.authorizationCode
        let issued = try await perform { try await sessions.createSession(identityToken: credential.identityToken, authorizationCode: code) }
        return AuthSession(issued, displayName: credential.givenName)
    }

    public func requestEmailCode(email: String) async throws(AuthenticationFailure) -> EmailCodeRequest {
        throw .unavailable
    }

    public func verifyEmailCode(email: String, code: String) async throws(AuthenticationFailure) -> AuthSession {
        throw .unavailable
    }

    public func refresh(_ session: AuthSession) async throws(AuthenticationFailure) -> AuthSession {
        let issued = try await perform { try await sessions.refreshSession(refreshToken: session.refreshToken) }
        return AuthSession(issued, displayName: session.displayName)
    }

    public func signOut(_ session: AuthSession) async throws(AuthenticationFailure) {
        try await perform { try await sessions.deleteSession() }
    }

    public func deleteAccount(_ session: AuthSession) async throws(AuthenticationFailure) {
        _ = try await perform { try await account.deleteAccount(reason: nil) }
    }

    private func perform<Value: Sendable>(_ operation: @Sendable () async throws -> Value) async throws(AuthenticationFailure) -> Value {
        do {
            return try await operation()
        } catch let failure as APIFailure {
            throw AuthenticationFailure(apiFailure: failure)
        } catch is URLError {
            throw .network
        } catch {
            throw .unexpected
        }
    }
}

extension AuthSession {
    init(_ issued: IssuedSession, displayName: String?) {
        self.init(accessToken: issued.accessToken, refreshToken: issued.refreshToken, expiresAt: issued.expiresAt, accountID: issued.accountID, displayName: displayName)
    }
}
