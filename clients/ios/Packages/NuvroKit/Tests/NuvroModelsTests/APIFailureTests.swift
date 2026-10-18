import Foundation
import Testing

@testable import NuvroModels

@Suite("APIFailure")
struct APIFailureTests {
    private let decoder = JSONDecoder()
    private let encoder = JSONEncoder()

    @Test("decodes RFC 9457 problem details the backend emits")
    func decodesProblemDetails() throws {
        let payload = Data(#"""
        {"type":"about:blank","title":"Not Found","status":404,"detail":"The requested resource does not exist.","code":"not_found","requestId":"4f0d5c1a"}
        """#.utf8)

        let failure = try decoder.decode(APIFailure.self, from: payload)

        #expect(failure.status == 404)
        #expect(failure.code == .notFound)
        #expect(failure.title == "Not Found")
        #expect(failure.detail == "The requested resource does not exist.")
        #expect(failure.requestID == "4f0d5c1a")
    }

    @Test("tolerates a missing request identifier and title")
    func toleratesMissingOptionalFields() throws {
        let payload = Data(#"{"type":"about:blank","status":409,"detail":"Trip already exists.","code":"conflict"}"#.utf8)

        let failure = try decoder.decode(APIFailure.self, from: payload)

        #expect(failure.code == .conflict)
        #expect(failure.requestID == nil)
        #expect(failure.title.isEmpty)
    }

    @Test("treats an unrecognised code as an internal error rather than crashing")
    func toleratesUnknownCode() throws {
        let payload = Data(#"{"type":"about:blank","status":500,"detail":"Unsupported.","code":"teapot"}"#.utf8)

        let failure = try decoder.decode(APIFailure.self, from: payload)

        #expect(failure.code == .internalError)
    }

    @Test("rejects a problem missing the detail the contract requires")
    func rejectsMissingDetail() {
        let payload = Data(#"{"type":"about:blank","status":404,"code":"not_found"}"#.utf8)

        #expect(throws: DecodingError.self) {
            _ = try decoder.decode(APIFailure.self, from: payload)
        }
    }

    @Test("round trips through the problem details shape")
    func roundTrips() throws {
        let original = APIFailure(status: 429, code: .rateLimited, title: "Too Many Requests", detail: "Request limit reached.", requestID: "abc")

        let encoded = try encoder.encode(original)
        let rendered = String(decoding: encoded, as: UTF8.self)

        #expect(rendered.contains("rate_limited"))
        #expect(rendered.contains("requestId"))
        #expect(rendered.contains("about:blank"))
        #expect(try decoder.decode(APIFailure.self, from: encoded) == original)
    }

    @Test("only transient codes are retryable")
    func marksTransientCodesRetryable() {
        for code in APIFailure.Code.allCases {
            let failure = APIFailure(status: 500, code: code, detail: "")
            let expected = code == .rateLimited || code == .upstreamUnavailable || code == .providerTimeout
            #expect(failure.isRetryable == expected)
        }
    }

    @Test("a bodiless status maps onto the matching code")
    func mapsBareStatuses() {
        #expect(APIFailure.status(401).code == .authenticationRequired)
        #expect(APIFailure.status(429).code == .rateLimited)
        #expect(APIFailure.status(503).code == .upstreamUnavailable)
        #expect(APIFailure.status(418).code == .internalError)
        #expect(APIFailure.status(403).isAuthenticationFailure)
    }

    @Test("every code in the API contract has a client case")
    func coversTheContractEnum() {
        let contractCodes = [
            "validation_error", "authentication_required", "forbidden", "not_found", "conflict", "rate_limited",
            "upstream_unavailable", "provider_timeout", "invalid_state", "idempotency_conflict", "internal_error",
            "method_not_allowed", "payload_too_large", "unsupported_media_type"
        ]

        #expect(Set(APIFailure.Code.allCases.map(\.rawValue)) == Set(contractCodes))
    }
}
