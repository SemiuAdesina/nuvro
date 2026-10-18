import Foundation
import Testing

@testable import NuvroAuthentication

@Suite("AuthenticationCoordinator")
@MainActor
struct AuthenticationCoordinatorTests {
    private let now = Date(timeIntervalSince1970: 1_800_000_000)

    private func coordinator(client: ScriptedAuthenticationClient = ScriptedAuthenticationClient(), stored: AuthSession? = nil) -> (AuthenticationCoordinator, InMemorySessionStore) {
        let store = InMemorySessionStore(session: stored)
        let now = now
        return (AuthenticationCoordinator(client: client, store: store, now: { now }), store)
    }

    @Test("restore with no stored session lands on signed out")
    func restoreEmpty() async {
        let (coordinator, _) = coordinator()
        await coordinator.restoreSession()
        #expect(coordinator.state == .signedOut(.none))
        #expect(coordinator.state.requiresAuthentication)
    }

    @Test("restore with a valid session signs in without a network call")
    func restoreValid() async {
        let client = ScriptedAuthenticationClient()
        client.refreshResult = .failure(.network)
        let (coordinator, _) = coordinator(client: client, stored: .fixture)
        await coordinator.restoreSession()
        #expect(coordinator.state == .signedIn(.fixture))
    }

    @Test("restore with a stale session refreshes and persists the renewed session")
    func restoreRefresh() async throws {
        let (coordinator, store) = coordinator(stored: .stale)
        await coordinator.restoreSession()
        #expect(coordinator.state == .signedIn(.renewed))
        #expect(try store.load() == .renewed)
    }

    @Test("a refresh rejected as unauthenticated expires the session and clears the store")
    func restoreExpired() async throws {
        let client = ScriptedAuthenticationClient()
        client.refreshResult = .failure(.unauthenticated)
        let (coordinator, store) = coordinator(client: client, stored: .stale)
        await coordinator.restoreSession()
        #expect(coordinator.state == .expired)
        #expect(coordinator.state.requiresAuthentication)
        #expect(try store.load() == nil)
    }

    @Test("a refresh that fails offline keeps the stored session usable")
    func restoreOffline() async throws {
        let client = ScriptedAuthenticationClient()
        client.refreshResult = .failure(.network)
        let (coordinator, store) = coordinator(client: client, stored: .stale)
        await coordinator.restoreSession()
        #expect(coordinator.state == .signedIn(.stale))
        #expect(try store.load() == .stale)
    }

    @Test("Apple success exchanges the credential and stores the backend session")
    func appleSuccess() async throws {
        let (coordinator, store) = coordinator()
        await coordinator.restoreSession()
        await coordinator.complete(.authorized(AppleCredential(userIdentifier: "u", identityToken: "t", authorizationCode: "c")))
        #expect(coordinator.state == .signedIn(.fixture))
        #expect(try store.load() == .fixture)
        #expect(coordinator.lastFailure == nil)
    }

    @Test("Apple cancellation is not an error and leaves the traveller signed out")
    func appleCancelled() async {
        let (coordinator, _) = coordinator()
        await coordinator.restoreSession()
        await coordinator.complete(.cancelled)
        #expect(coordinator.state == .signedOut(.none))
        #expect(coordinator.lastFailure == nil)
    }

    @Test("Apple failure and backend rejection surface a failure without a session")
    func appleFailure() async throws {
        let client = ScriptedAuthenticationClient()
        client.appleResult = .failure(.unavailable)
        let (coordinator, store) = coordinator(client: client)
        await coordinator.complete(.failed)
        #expect(coordinator.lastFailure == .unexpected)
        await coordinator.complete(.authorized(AppleCredential(userIdentifier: "u", identityToken: "t", authorizationCode: "c")))
        #expect(coordinator.lastFailure == .unavailable)
        #expect(coordinator.state == .signedOut(.none))
        #expect(try store.load() == nil)
    }

    @Test("sign out clears the store even when the backend call fails")
    func signOut() async throws {
        let client = ScriptedAuthenticationClient()
        client.signOutResult = .failure(.network)
        let (coordinator, store) = coordinator(client: client, stored: .fixture)
        await coordinator.restoreSession()
        await coordinator.signOut()
        #expect(client.signOutCount == 1)
        #expect(coordinator.state == .signedOut(.signedOut))
        #expect(try store.load() == nil)
    }

    @Test("delete account only signs out after the backend confirms")
    func deleteAccount() async throws {
        let client = ScriptedAuthenticationClient()
        client.deleteResult = .failure(.unavailable)
        let (coordinator, store) = coordinator(client: client, stored: .fixture)
        await coordinator.restoreSession()
        await #expect(throws: AuthenticationFailure.unavailable) { try await coordinator.deleteAccount() }
        #expect(coordinator.state == .signedIn(.fixture))
        #expect(try store.load() == .fixture)
        client.deleteResult = .success(())
        try await coordinator.deleteAccount()
        #expect(coordinator.state == .signedOut(.deleted))
        #expect(try store.load() == nil)
    }

    @Test("establishing a session from email verification persists it")
    func establish() throws {
        let (coordinator, store) = coordinator()
        coordinator.establish(.fixture)
        #expect(coordinator.state == .signedIn(.fixture))
        #expect(try store.load() == .fixture)
    }

    @Test("a valid session hands out its access token without a network call")
    func freshTokenWithoutRefresh() async {
        let client = ScriptedAuthenticationClient()
        let (coordinator, _) = coordinator(client: client, stored: .fixture)
        await coordinator.restoreSession()
        #expect(await coordinator.accessToken() == AuthSession.fixture.accessToken)
        #expect(client.refreshCount == 0)
    }

    @Test("a rejected access token is refreshed once and the renewed token returned")
    func rejectedTokenIsRefreshed() async throws {
        let client = ScriptedAuthenticationClient()
        let (coordinator, store) = coordinator(client: client, stored: .fixture)
        await coordinator.restoreSession()
        let renewed = await coordinator.accessTokenAfterRejection(of: AuthSession.fixture.accessToken)
        #expect(renewed == AuthSession.renewed.accessToken)
        #expect(client.refreshCount == 1)
        #expect(try store.load() == .renewed)
        #expect(coordinator.state == .signedIn(.renewed))
    }

    @Test("a rejection of a token that was already replaced returns the current one without refreshing")
    func staleRejectionUsesCurrentToken() async {
        let client = ScriptedAuthenticationClient()
        let (coordinator, _) = coordinator(client: client, stored: .renewed)
        await coordinator.restoreSession()
        let token = await coordinator.accessTokenAfterRejection(of: "an-older-token")
        #expect(token == AuthSession.renewed.accessToken)
        #expect(client.refreshCount == 0)
    }

    @Test("concurrent rejections share one refresh")
    func concurrentRejectionsCoalesce() async {
        let client = ScriptedAuthenticationClient()
        let (coordinator, _) = coordinator(client: client, stored: .fixture)
        await coordinator.restoreSession()
        async let first = coordinator.accessTokenAfterRejection(of: AuthSession.fixture.accessToken)
        async let second = coordinator.accessTokenAfterRejection(of: AuthSession.fixture.accessToken)
        let tokens = await [first, second]
        #expect(tokens == [AuthSession.renewed.accessToken, AuthSession.renewed.accessToken])
        #expect(client.refreshCount == 1)
    }

    @Test("a refresh refused after a rejection expires the session")
    func rejectionWithRefusedRefreshExpires() async throws {
        let client = ScriptedAuthenticationClient()
        client.refreshResult = .failure(.unauthenticated)
        let (coordinator, store) = coordinator(client: client, stored: .fixture)
        await coordinator.restoreSession()
        let token = await coordinator.accessTokenAfterRejection(of: AuthSession.fixture.accessToken)
        #expect(token == nil)
        #expect(coordinator.state == .expired)
        #expect(try store.load() == nil)
    }
}
