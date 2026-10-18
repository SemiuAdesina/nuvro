import Foundation
import NuvroModels
import Testing

@testable import NuvroAPI

@Suite("NuvroSessionGateway")
struct SessionGatewayTests {
    private let issued = #"{"accessToken":"a.b.c","refreshToken":"r-1","expiresAt":"2026-09-09T12:15:00Z","accountId":"3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"}"#

    @Test("exchanges an identity token and authorization code for a session")
    func createsASession() async throws {
        let transport = StubTransport((.created, issued))
        let session = try await NuvroSessionGateway(client: testClient(transport, token: nil)).createSession(identityToken: "apple-token", authorizationCode: "code-1")

        #expect(session.accessToken == "a.b.c")
        #expect(session.refreshToken == "r-1")
        #expect(session.accountID == "3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f")
        #expect(session.expiresAt == Date(timeIntervalSince1970: 1_788_956_100))
        let exchange = try #require(transport.exchanges.first)
        #expect(exchange.method == .post)
        #expect(exchange.path == "/v1/auth/sessions")
        #expect(exchange.json["authorizationCode"] as? String == "code-1")
        #expect(exchange.json["identityToken"] as? String == "apple-token")
        #expect(exchange.headers[.authorization] == nil)
    }

    @Test("a rejected credential surfaces as the contract's authentication failure")
    func rejectedCredential() async {
        let transport = StubTransport(StubTransport.problem(401, "authentication_required"))
        do {
            _ = try await NuvroSessionGateway(client: testClient(transport, token: nil)).createSession(identityToken: "bad", authorizationCode: nil)
        } catch let failure as APIFailure {
            #expect(failure.code == .authenticationRequired)
            #expect(failure.status == 401)
            #expect(failure.requestID == "req-1")
        } catch {
            Issue.record("unexpected error \(error)")
        }
    }

    @Test("refresh posts the refresh token without a bearer header")
    func refreshes() async throws {
        let transport = StubTransport((.created, issued))
        _ = try await NuvroSessionGateway(client: testClient(transport, token: nil)).refreshSession(refreshToken: "r-0")
        let exchange = try #require(transport.exchanges.first)
        #expect(exchange.path == "/v1/auth/sessions/refresh")
        #expect(exchange.json["refreshToken"] as? String == "r-0")
    }

    @Test("sign out sends the bearer credential and accepts no content")
    func signsOut() async throws {
        let transport = StubTransport((.noContent, ""))
        try await NuvroSessionGateway(client: testClient(transport)).deleteSession()
        let exchange = try #require(transport.exchanges.first)
        #expect(exchange.method == .delete)
        #expect(exchange.headers[.authorization] == "Bearer access-1")
        #expect(exchange.headers[.userAgent] == "NUVRO-iOS/1.0 (7)")
        #expect(exchange.headers[.init("X-Request-ID")!]?.count == 32)
    }
}
