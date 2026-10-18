import Foundation

public struct AuthSession: Sendable, Hashable, Codable {
    public let accessToken: String
    public let refreshToken: String
    public let expiresAt: Date
    public let accountID: String
    public let displayName: String?

    public init(accessToken: String, refreshToken: String, expiresAt: Date, accountID: String, displayName: String? = nil) {
        self.accessToken = accessToken
        self.refreshToken = refreshToken
        self.expiresAt = expiresAt
        self.accountID = accountID
        self.displayName = displayName
    }

    public func isValid(at now: Date, margin: TimeInterval = 60) -> Bool {
        expiresAt.timeIntervalSince(now) > margin
    }

    private enum CodingKeys: String, CodingKey {
        case accessToken = "access_token"
        case refreshToken = "refresh_token"
        case expiresAt = "expires_at"
        case accountID = "account_id"
        case displayName = "display_name"
    }
}

extension AuthSession: CustomStringConvertible {
    public var description: String {
        "AuthSession(account: \(accountID), expires: \(expiresAt.formatted(.iso8601)))"
    }
}

public enum SessionStoreFailure: Error, Equatable, Sendable {
    case unreadable
    case unwritable(status: Int)
}

public protocol SessionStore: Sendable {
    func load() throws -> AuthSession?
    func save(_ session: AuthSession) throws
    func clear() throws
}

public final class InMemorySessionStore: SessionStore, @unchecked Sendable {
    private let lock = NSLock()
    private var session: AuthSession?

    public init(session: AuthSession? = nil) {
        self.session = session
    }

    public func load() throws -> AuthSession? {
        lock.withLock { session }
    }

    public func save(_ session: AuthSession) throws {
        lock.withLock { self.session = session }
    }

    public func clear() throws {
        lock.withLock { session = nil }
    }
}
