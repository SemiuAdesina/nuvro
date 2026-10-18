import Foundation
import HTTPTypes
import OpenAPIRuntime

public protocol SessionCredentials: Sendable {
    func accessToken() async -> String?
    func accessTokenAfterRejection(of rejected: String) async -> String?
}

public struct BearerMiddleware: ClientMiddleware {
    static let anonymousOperations: Set<String> = ["createSession", "refreshSession", "getLiveness", "getReadiness"]
    private let credentials: any SessionCredentials

    public init(credentials: any SessionCredentials) {
        self.credentials = credentials
    }

    public func intercept(
        _ request: HTTPRequest, body: HTTPBody?, baseURL: URL, operationID: String,
        next: @Sendable (HTTPRequest, HTTPBody?, URL) async throws -> (HTTPResponse, HTTPBody?)
    ) async throws -> (HTTPResponse, HTTPBody?) {
        guard !Self.anonymousOperations.contains(operationID) else { return try await next(request, body, baseURL) }
        var request = request
        let token = await credentials.accessToken()
        if let token, !token.isEmpty {
            request.headerFields[.authorization] = "Bearer \(token)"
        }
        let (response, responseBody) = try await next(request, body, baseURL)
        guard response.status == .unauthorized, let token, Self.canReplay(body) else { return (response, responseBody) }
        guard let renewed = await credentials.accessTokenAfterRejection(of: token), renewed != token else { return (response, responseBody) }
        request.headerFields[.authorization] = "Bearer \(renewed)"
        return try await next(request, body, baseURL)
    }

    static func canReplay(_ body: HTTPBody?) -> Bool {
        guard let body else { return true }
        return body.iterationBehavior == .multiple
    }
}

public struct ClientIdentityMiddleware: ClientMiddleware {
    private static let requestID = HTTPField.Name("X-Request-ID")!
    private let userAgent: String

    public init(appVersion: String, build: String) {
        userAgent = "NUVRO-iOS/\(appVersion) (\(build))"
    }

    public func intercept(
        _ request: HTTPRequest, body: HTTPBody?, baseURL: URL, operationID: String,
        next: @Sendable (HTTPRequest, HTTPBody?, URL) async throws -> (HTTPResponse, HTTPBody?)
    ) async throws -> (HTTPResponse, HTTPBody?) {
        var request = request
        request.headerFields[.userAgent] = userAgent
        request.headerFields[Self.requestID] = UUID().uuidString.replacingOccurrences(of: "-", with: "").lowercased()
        return try await next(request, body, baseURL)
    }
}

public struct LenientDateTranscoder: DateTranscoder {
    public init() {}

    public func encode(_ date: Date) throws -> String {
        date.formatted(.iso8601)
    }

    public func decode(_ dateString: String) throws -> Date {
        if let date = try? Date(dateString, strategy: .iso8601.time(includingFractionalSeconds: true)) { return date }
        if let date = try? Date(dateString, strategy: .iso8601) { return date }
        throw DecodingError.dataCorrupted(.init(codingPath: [], debugDescription: "Expected an RFC 3339 date-time, got \(dateString)"))
    }
}
