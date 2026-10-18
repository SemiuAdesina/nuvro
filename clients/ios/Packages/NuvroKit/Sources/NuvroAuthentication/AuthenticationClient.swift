import Foundation
import NuvroModels

public enum AuthenticationFailure: Error, Equatable, Sendable {
    case invalidEmail
    case invalidCode
    case expiredCode
    case rateLimited(retryAfterSeconds: Int?)
    case unauthenticated
    case network
    case unavailable
    case unexpected

    public init(apiFailure: APIFailure) {
        switch apiFailure.code {
        case .validationError: self = .invalidEmail
        case .authenticationRequired, .forbidden: self = .unauthenticated
        case .notFound: self = .invalidCode
        case .conflict, .invalidState, .idempotencyConflict: self = .expiredCode
        case .rateLimited: self = .rateLimited(retryAfterSeconds: nil)
        case .upstreamUnavailable, .providerTimeout: self = .unavailable
        case .methodNotAllowed, .internalError, .payloadTooLarge, .unsupportedMediaType: self = .unexpected
        }
    }

    public var isRetryable: Bool {
        switch self {
        case .network, .unavailable, .rateLimited, .unexpected: true
        case .invalidEmail, .invalidCode, .expiredCode, .unauthenticated: false
        }
    }
}

public struct EmailCodeRequest: Sendable, Hashable {
    public let email: String
    public let maskedEmail: String
    public let codeLength: Int
    public let resendAfterSeconds: Int

    public init(email: String, maskedEmail: String, codeLength: Int, resendAfterSeconds: Int) {
        self.email = email
        self.maskedEmail = maskedEmail
        self.codeLength = codeLength
        self.resendAfterSeconds = resendAfterSeconds
    }
}

public protocol AuthenticationClient: Sendable {
    func signInWithApple(_ credential: AppleCredential) async throws(AuthenticationFailure) -> AuthSession
    func requestEmailCode(email: String) async throws(AuthenticationFailure) -> EmailCodeRequest
    func verifyEmailCode(email: String, code: String) async throws(AuthenticationFailure) -> AuthSession
    func refresh(_ session: AuthSession) async throws(AuthenticationFailure) -> AuthSession
    func signOut(_ session: AuthSession) async throws(AuthenticationFailure)
    func deleteAccount(_ session: AuthSession) async throws(AuthenticationFailure)
}

public struct PendingAuthenticationClient: AuthenticationClient {
    public init() {}

    public func signInWithApple(_ credential: AppleCredential) async throws(AuthenticationFailure) -> AuthSession { throw .unavailable }
    public func requestEmailCode(email: String) async throws(AuthenticationFailure) -> EmailCodeRequest { throw .unavailable }
    public func verifyEmailCode(email: String, code: String) async throws(AuthenticationFailure) -> AuthSession { throw .unavailable }
    public func refresh(_ session: AuthSession) async throws(AuthenticationFailure) -> AuthSession { throw .unavailable }
    public func signOut(_ session: AuthSession) async throws(AuthenticationFailure) { throw .unavailable }
    public func deleteAccount(_ session: AuthSession) async throws(AuthenticationFailure) { throw .unavailable }
}

public enum EmailMasking {
    public static func masked(_ email: String) -> String {
        guard let at = email.firstIndex(of: "@") else { return email }
        let local = email[..<at]
        let domain = email[email.index(after: at)...]
        let visible = local.prefix(1)
        let hidden = String(repeating: "•", count: max(local.count - 1, 1))
        return "\(visible)\(hidden)@\(domain)"
    }
}
