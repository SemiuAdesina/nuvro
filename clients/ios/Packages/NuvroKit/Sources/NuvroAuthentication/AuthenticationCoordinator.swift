import Foundation
import NuvroAPI
import Observation

public enum SignedOutReason: Equatable, Sendable {
    case none
    case signedOut
    case expired
    case deleted
}

public enum AuthenticationState: Equatable, Sendable {
    case unknown
    case signedOut(SignedOutReason)
    case authenticating
    case signedIn(AuthSession)
    case refreshing(AuthSession)
    case expired

    public var requiresAuthentication: Bool {
        switch self {
        case .signedOut, .expired, .authenticating: true
        case .unknown, .signedIn, .refreshing: false
        }
    }

    public var session: AuthSession? {
        switch self {
        case .signedIn(let session), .refreshing(let session): session
        case .unknown, .signedOut, .authenticating, .expired: nil
        }
    }
}

@Observable
@MainActor
public final class AuthenticationCoordinator: SessionCredentials {
    public private(set) var state: AuthenticationState = .unknown
    public private(set) var lastFailure: AuthenticationFailure?
    public let client: any AuthenticationClient
    private let store: any SessionStore
    private let now: @Sendable () -> Date
    private var renewal: Task<String?, Never>?

    public init(client: any AuthenticationClient, store: any SessionStore, now: @escaping @Sendable () -> Date = { Date() }) {
        self.client = client
        self.store = store
        self.now = now
    }

    public func restoreSession() async {
        guard let stored = try? store.load() else {
            state = .signedOut(.none)
            return
        }
        if stored.isValid(at: now()) {
            state = .signedIn(stored)
            return
        }
        state = .refreshing(stored)
        do {
            let renewed = try await client.refresh(stored)
            try? store.save(renewed)
            state = .signedIn(renewed)
        } catch .unauthenticated {
            try? store.clear()
            state = .expired
        } catch {
            state = .signedIn(stored)
        }
    }

    public func complete(_ outcome: AppleAuthorizationOutcome) async {
        switch outcome {
        case .cancelled:
            lastFailure = nil
        case .failed:
            lastFailure = .unexpected
        case .authorized(let credential):
            await authenticate { client throws(AuthenticationFailure) in try await client.signInWithApple(credential) }
        }
    }

    public func signIn(credential: String) async {
        let development = AppleCredential(userIdentifier: "development", identityToken: credential, authorizationCode: "")
        await authenticate { client throws(AuthenticationFailure) in try await client.signInWithApple(development) }
    }

    public func accessToken() async -> String? {
        await freshAccessToken()
    }

    public func freshAccessToken() async -> String? {
        guard let session = state.session else { return nil }
        if session.isValid(at: now()) { return session.accessToken }
        return await renewCoalesced(session)
    }

    public func accessTokenAfterRejection(of rejected: String) async -> String? {
        guard let session = state.session else { return nil }
        if session.accessToken != rejected { return session.accessToken }
        return await renewCoalesced(session)
    }

    private func renewCoalesced(_ session: AuthSession) async -> String? {
        if let renewal { return await renewal.value }
        let task = Task<String?, Never> { await self.renew(session) }
        renewal = task
        let token = await task.value
        renewal = nil
        return token
    }

    private func renew(_ session: AuthSession) async -> String? {
        do {
            let renewed = try await client.refresh(session)
            try? store.save(renewed)
            state = .signedIn(renewed)
            return renewed.accessToken
        } catch .unauthenticated {
            try? store.clear()
            state = .expired
            return nil
        } catch {
            return session.accessToken
        }
    }

    public func establish(_ session: AuthSession) {
        try? store.save(session)
        lastFailure = nil
        state = .signedIn(session)
    }

    public func signOut() async {
        if let session = state.session {
            try? await client.signOut(session)
        }
        try? store.clear()
        lastFailure = nil
        state = .signedOut(.signedOut)
    }

    public func deleteAccount() async throws(AuthenticationFailure) {
        guard let session = state.session else { throw .unauthenticated }
        try await client.deleteAccount(session)
        try? store.clear()
        state = .signedOut(.deleted)
    }

    public func clearFailure() {
        lastFailure = nil
    }

    private func authenticate(_ operation: (any AuthenticationClient) async throws(AuthenticationFailure) -> AuthSession) async {
        guard state != .authenticating else { return }
        state = .authenticating
        lastFailure = nil
        do {
            let session = try await operation(client)
            try? store.save(session)
            state = .signedIn(session)
        } catch {
            lastFailure = error
            state = .signedOut(.none)
        }
    }
}
