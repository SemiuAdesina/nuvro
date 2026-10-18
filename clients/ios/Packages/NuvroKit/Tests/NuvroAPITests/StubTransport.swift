import Foundation
import HTTPTypes
import OpenAPIRuntime

@testable import NuvroAPI

final class StubTransport: ClientTransport, @unchecked Sendable {
    struct Exchange: Sendable {
        let method: HTTPRequest.Method
        let path: String
        let headers: HTTPFields
        let body: String?

        var json: [String: Any] {
            guard let body, let data = body.data(using: .utf8), let object = try? JSONSerialization.jsonObject(with: data) as? [String: Any] else { return [:] }
            return object
        }
    }

    typealias Canned = (HTTPResponse.Status, String)

    private let lock = NSLock()
    private var routes: [String: [Canned]]
    private(set) var exchanges: [Exchange] = []

    init(sequences: [String: [Canned]]) {
        self.routes = sequences
    }

    convenience init(_ routes: [String: Canned]) {
        self.init(sequences: routes.mapValues { [$0] })
    }

    convenience init(_ single: Canned) {
        self.init(["*": single])
    }

    convenience init(sequence: [Canned]) {
        self.init(sequences: ["*": sequence])
    }

    func send(_ request: HTTPRequest, body: HTTPBody?, baseURL: URL, operationID: String) async throws -> (HTTPResponse, HTTPBody?) {
        var sent: String?
        if let body { sent = try await String(collecting: body, upTo: 1 << 16) }
        let path = request.path ?? ""
        let (status, payload): Canned = lock.withLock {
            exchanges.append(Exchange(method: request.method, path: path, headers: request.headerFields, body: sent))
            let key = [request.method.rawValue + " " + path, path, "*"].first { routes[$0]?.isEmpty == false }
            guard let key, var queue = routes[key] else { return (.internalServerError, "") }
            let canned = queue.count > 1 ? queue.removeFirst() : queue[0]
            routes[key] = queue
            return canned
        }
        var response = HTTPResponse(status: status)
        response.headerFields[.contentType] = status.code >= 400 ? "application/problem+json" : "application/json"
        return (response, payload.isEmpty ? nil : HTTPBody(payload))
    }

    func exchange(_ path: String) -> Exchange? {
        lock.withLock { exchanges.first { $0.path == path } }
    }
}

extension StubTransport {
    static func problem(_ status: Int, _ code: String, requestID: String = "req-1") -> Canned {
        (HTTPResponse.Status(code: status), #"{"type":"about:blank","title":"Problem","status":\#(status),"detail":"Refused.","code":"\#(code)","requestId":"\#(requestID)"}"#)
    }
}

func testClient(_ transport: StubTransport, token: String? = "access-1") -> Client {
    testClient(transport, credentials: StaticCredentials(token: token))
}

func testClient(_ transport: StubTransport, credentials: any SessionCredentials) -> Client {
    NuvroAPIClient.make(environment: APIEnvironment(baseURL: URL(string: "https://api.example.test")!), credentials: credentials, appVersion: "1.0", build: "7", transport: transport)
}
