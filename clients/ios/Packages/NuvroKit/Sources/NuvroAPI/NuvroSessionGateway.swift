import Foundation
import NuvroModels

public struct NuvroSessionGateway: SessionGateway {
    private let client: any APIProtocol

    public init(client: any APIProtocol) {
        self.client = client
    }

    public func createSession(identityToken: String, authorizationCode: String?) async throws -> IssuedSession {
        let body = Components.Schemas.ComponentsSchemasAuthSessionRequest(authorizationCode: authorizationCode, identityToken: identityToken)
        switch try await client.createSession(.init(body: .json(body))) {
        case .created(let created): return try IssuedSession(created.body.json)
        case .default(let statusCode, let problem): throw APIFailure.from(problem, statusCode: statusCode)
        }
    }

    public func refreshSession(refreshToken: String) async throws -> IssuedSession {
        let body = Components.Schemas.ComponentsSchemasAuthRefreshRequest(refreshToken: refreshToken)
        switch try await client.refreshSession(.init(body: .json(body))) {
        case .created(let created): return try IssuedSession(created.body.json)
        case .default(let statusCode, let problem): throw APIFailure.from(problem, statusCode: statusCode)
        }
    }

    public func deleteSession() async throws {
        switch try await client.deleteSession(.init()) {
        case .noContent: return
        case .default(let statusCode, let problem): throw APIFailure.from(problem, statusCode: statusCode)
        }
    }
}

extension IssuedSession {
    init(_ session: Components.Schemas.ComponentsSchemasAuthSession) {
        self.init(accessToken: session.accessToken, refreshToken: session.refreshToken, expiresAt: session.expiresAt, accountID: session.accountId)
    }
}
