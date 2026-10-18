import Foundation
import NuvroModels
import Testing

@testable import NuvroAuthentication

@Suite("Apple credential and session store")
struct CredentialAndStoreTests {
    @Test("the Apple cancellation error code maps to cancellation, anything else to failure")
    func cancellation() {
        let cancelled = NSError(domain: "com.apple.AuthenticationServices.AuthorizationError", code: 1001)
        let failed = NSError(domain: "com.apple.AuthenticationServices.AuthorizationError", code: 1004)
        #expect(AppleAuthorizationMapper.outcome(for: cancelled) == .cancelled)
        #expect(AppleAuthorizationMapper.outcome(for: failed) == .failed)
        #expect(AppleAuthorizationMapper.outcome(for: URLError(.notConnectedToInternet)) == .failed)
    }

    @Test("credential and session descriptions never print tokens")
    func redaction() {
        let credential = AppleCredential(userIdentifier: "001234.abc", identityToken: "secret-token", authorizationCode: "secret-code", email: "x@privaterelay.appleid.com")
        #expect(!credential.description.contains("secret"))
        #expect(!credential.description.contains("privaterelay"))
        #expect(!AuthSession.fixture.description.contains("access-1"))
    }

    @Test("session validity respects the expiry margin")
    func validity() {
        let session = AuthSession(accessToken: "a", refreshToken: "r", expiresAt: Date(timeIntervalSince1970: 1_000), accountID: "x")
        #expect(session.isValid(at: Date(timeIntervalSince1970: 500)))
        #expect(!session.isValid(at: Date(timeIntervalSince1970: 950)))
        #expect(!session.isValid(at: Date(timeIntervalSince1970: 2_000)))
    }

    @Test("session round trips through the contract keys")
    func codable() throws {
        let data = try JSONEncoder().encode(AuthSession.fixture)
        let text = String(decoding: data, as: UTF8.self)
        #expect(text.contains("access_token") && text.contains("account_id"))
        #expect(try JSONDecoder().decode(AuthSession.self, from: data) == .fixture)
    }

    @Test("the in-memory store saves, loads and clears")
    func memoryStore() throws {
        let store = InMemorySessionStore()
        #expect(try store.load() == nil)
        try store.save(.fixture)
        #expect(try store.load() == .fixture)
        try store.clear()
        #expect(try store.load() == nil)
    }

    @Test("API failures map to authentication failures with retryability")
    func failureMapping() {
        #expect(AuthenticationFailure(apiFailure: APIFailure(status: 429, code: .rateLimited, detail: "")) == .rateLimited(retryAfterSeconds: nil))
        #expect(AuthenticationFailure(apiFailure: APIFailure(status: 401, code: .authenticationRequired, detail: "")) == .unauthenticated)
        #expect(AuthenticationFailure(apiFailure: APIFailure(status: 503, code: .upstreamUnavailable, detail: "")).isRetryable)
        #expect(!AuthenticationFailure.invalidCode.isRetryable)
    }

    @Test("email masking keeps the first character and the domain")
    func masking() {
        #expect(EmailMasking.masked("alex@example.com") == "a•••@example.com")
        #expect(EmailMasking.masked("a@b.co") == "a•@b.co")
        #expect(EmailMasking.masked("not-an-email") == "not-an-email")
    }
}
