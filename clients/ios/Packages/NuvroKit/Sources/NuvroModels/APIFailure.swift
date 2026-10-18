import Foundation

public struct APIFailure: Sendable, Hashable, Codable, Error {
    public enum Code: String, Sendable, Codable, CaseIterable {
        case validationError = "validation_error"
        case authenticationRequired = "authentication_required"
        case forbidden
        case notFound = "not_found"
        case conflict
        case rateLimited = "rate_limited"
        case upstreamUnavailable = "upstream_unavailable"
        case providerTimeout = "provider_timeout"
        case invalidState = "invalid_state"
        case idempotencyConflict = "idempotency_conflict"
        case internalError = "internal_error"
        case methodNotAllowed = "method_not_allowed"
        case payloadTooLarge = "payload_too_large"
        case unsupportedMediaType = "unsupported_media_type"
    }

    public let status: Int
    public let code: Code
    public let title: String
    public let detail: String
    public let requestID: String?

    private enum CodingKeys: String, CodingKey {
        case type
        case title
        case status
        case detail
        case code
        case requestID = "requestId"
    }

    public init(status: Int, code: Code, title: String = "", detail: String, requestID: String? = nil) {
        self.status = status
        self.code = code
        self.title = title
        self.detail = detail
        self.requestID = requestID
    }

    public init(from decoder: any Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        status = try container.decode(Int.self, forKey: .status)
        code = Code(rawValue: try container.decode(String.self, forKey: .code)) ?? .internalError
        title = try container.decodeIfPresent(String.self, forKey: .title) ?? ""
        detail = try container.decode(String.self, forKey: .detail)
        requestID = try container.decodeIfPresent(String.self, forKey: .requestID)
    }

    public func encode(to encoder: any Encoder) throws {
        var container = encoder.container(keyedBy: CodingKeys.self)
        try container.encode("about:blank", forKey: .type)
        try container.encode(title, forKey: .title)
        try container.encode(status, forKey: .status)
        try container.encode(detail, forKey: .detail)
        try container.encode(code.rawValue, forKey: .code)
        try container.encodeIfPresent(requestID, forKey: .requestID)
    }

    public var isRetryable: Bool {
        switch code {
        case .rateLimited, .upstreamUnavailable, .providerTimeout: true
        default: false
        }
    }

    public var isAuthenticationFailure: Bool {
        code == .authenticationRequired || code == .forbidden
    }

    public static func status(_ status: Int) -> APIFailure {
        APIFailure(status: status, code: Code(status: status), detail: "")
    }
}

extension APIFailure.Code {
    init(status: Int) {
        switch status {
        case 400: self = .validationError
        case 401: self = .authenticationRequired
        case 403: self = .forbidden
        case 404: self = .notFound
        case 405: self = .methodNotAllowed
        case 409: self = .conflict
        case 413: self = .payloadTooLarge
        case 415: self = .unsupportedMediaType
        case 429: self = .rateLimited
        case 503: self = .upstreamUnavailable
        case 504: self = .providerTimeout
        default: self = .internalError
        }
    }
}
