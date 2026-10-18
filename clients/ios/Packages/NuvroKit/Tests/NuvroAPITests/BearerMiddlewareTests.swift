import Foundation
import NuvroModels
import Testing

@testable import NuvroAPI

final class RenewingCredentials: SessionCredentials, @unchecked Sendable {
    private let lock = NSLock()
    private var token: String?
    private let renewed: String?
    private(set) var rejections: [String] = []

    init(token: String?, renewed: String?) {
        self.token = token
        self.renewed = renewed
    }

    func accessToken() async -> String? { lock.withLock { token } }

    func accessTokenAfterRejection(of rejected: String) async -> String? {
        lock.withLock {
            rejections.append(rejected)
            if let renewed { token = renewed }
            return renewed
        }
    }
}

@Suite("BearerMiddleware")
struct BearerMiddlewareTests {
    private let profile = #"{"accessibilityPreferences":[],"departureAirport":null,"homeCountry":null,"homeTimezone":null,"hourCycle":"24","locale":"en","preferredCurrency":"USD","preferredLanguage":"en","units":"metric","userId":"3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"}"#

    @Test("a rejected access token is renewed once and the request replayed")
    func retriesOnceAfterRenewal() async throws {
        let transport = StubTransport(sequence: [StubTransport.problem(401, "authentication_required"), (.ok, profile)])
        let credentials = RenewingCredentials(token: "stale", renewed: "fresh")
        let loaded = try await NuvroAccountGateway(client: testClient(transport, credentials: credentials)).profile()
        #expect(loaded.preferredCurrency == "USD")
        #expect(credentials.rejections == ["stale"])
        #expect(transport.exchanges.map { $0.headers[.authorization] } == ["Bearer stale", "Bearer fresh"])
    }

    @Test("a rejection that cannot be renewed surfaces as the authentication failure")
    func givesUpWhenRenewalFails() async {
        let transport = StubTransport(sequence: [StubTransport.problem(401, "authentication_required"), (.ok, profile)])
        let credentials = RenewingCredentials(token: "stale", renewed: nil)
        do {
            _ = try await NuvroAccountGateway(client: testClient(transport, credentials: credentials)).profile()
            Issue.record("a dead session produced a profile")
        } catch let failure as APIFailure {
            #expect(failure.code == .authenticationRequired)
        } catch {
            Issue.record("unexpected error \(error)")
        }
        #expect(transport.exchanges.count == 1)
    }

    @Test("a request sent without a token is never retried")
    func anonymousRequestsAreNotRetried() async {
        let transport = StubTransport(sequence: [StubTransport.problem(401, "authentication_required"), (.ok, profile)])
        let credentials = RenewingCredentials(token: nil, renewed: "fresh")
        _ = try? await NuvroAccountGateway(client: testClient(transport, credentials: credentials)).profile()
        #expect(credentials.rejections.isEmpty)
        #expect(transport.exchanges.count == 1)
    }

    @Test("a non-authentication failure passes straight through")
    func otherFailuresAreNotRetried() async {
        let transport = StubTransport(sequence: [StubTransport.problem(503, "upstream_unavailable"), (.ok, profile)])
        let credentials = RenewingCredentials(token: "current", renewed: "fresh")
        _ = try? await NuvroAccountGateway(client: testClient(transport, credentials: credentials)).profile()
        #expect(credentials.rejections.isEmpty)
        #expect(transport.exchanges.count == 1)
    }

    @Test("session endpoints never ask for a token, so an expired session can still refresh")
    func sessionEndpointsAreAnonymous() async throws {
        let issued = #"{"accessToken":"a.b.c","refreshToken":"r-1","expiresAt":"2026-09-09T12:15:00Z","accountId":"3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"}"#
        let transport = StubTransport((.created, issued))
        let credentials = BlockingCredentials()
        _ = try await NuvroSessionGateway(client: testClient(transport, credentials: credentials)).refreshSession(refreshToken: "r-0")
        _ = try await NuvroSessionGateway(client: testClient(transport, credentials: credentials)).createSession(identityToken: "t", authorizationCode: nil)
        #expect(transport.exchanges.allSatisfy { $0.headers[.authorization] == nil })
        #expect(credentials.requests == 0)
    }
}

final class BlockingCredentials: SessionCredentials, @unchecked Sendable {
    private let lock = NSLock()
    private var count = 0
    var requests: Int { lock.withLock { count } }

    func accessToken() async -> String? {
        lock.withLock { count += 1 }
        try? await Task.sleep(for: .seconds(30))
        return nil
    }

    func accessTokenAfterRejection(of rejected: String) async -> String? { nil }
}
