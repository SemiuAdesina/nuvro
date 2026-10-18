import Foundation

public final class SessionTokenBox: SessionCredentials, @unchecked Sendable {
    private let lock = NSLock()
    private var credentials: (any SessionCredentials)?

    public init() {}

    public func install(_ credentials: any SessionCredentials) {
        lock.withLock { self.credentials = credentials }
    }

    private var installed: (any SessionCredentials)? {
        lock.withLock { credentials }
    }

    public func accessToken() async -> String? {
        await installed?.accessToken()
    }

    public func accessTokenAfterRejection(of rejected: String) async -> String? {
        await installed?.accessTokenAfterRejection(of: rejected)
    }
}

public struct StaticCredentials: SessionCredentials {
    private let token: String?

    public init(token: String?) {
        self.token = token
    }

    public func accessToken() async -> String? { token }

    public func accessTokenAfterRejection(of rejected: String) async -> String? { nil }
}
