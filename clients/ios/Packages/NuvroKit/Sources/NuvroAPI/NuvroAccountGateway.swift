import Foundation
import NuvroModels

public struct NuvroAccountGateway: AccountGateway {
    private let client: any APIProtocol

    public init(client: any APIProtocol) {
        self.client = client
    }

    public func profile() async throws -> AccountProfile {
        switch try await client.getMe(.init()) {
        case .ok(let ok): return try AccountProfile(ok.body.json)
        case .default(let statusCode, let problem): throw APIFailure.from(problem, statusCode: statusCode)
        }
    }

    public func preferences() async throws -> AccountPreferences {
        switch try await client.getPreferences(.init()) {
        case .ok(let ok): return try AccountPreferences(ok.body.json)
        case .default(let statusCode, let problem): throw APIFailure.from(problem, statusCode: statusCode)
        }
    }

    public func personalization() async throws -> AccountPersonalization {
        async let settings = client.getPersonalization(.init())
        async let insights = client.listPersonalizationInsights(.init())
        let enabled: Bool
        switch try await settings {
        case .ok(let ok): enabled = try ok.body.json.historyLearningEnabled
        case .default(let statusCode, let problem): throw APIFailure.from(problem, statusCode: statusCode)
        }
        switch try await insights {
        case .ok(let ok): return AccountPersonalization(historyLearningEnabled: enabled, insights: try ok.body.json.insights.map { LearnedPreference($0) })
        case .default(let statusCode, let problem): throw APIFailure.from(problem, statusCode: statusCode)
        }
    }

    public func setHistoryLearning(_ enabled: Bool) async throws -> AccountPersonalization {
        let body = Components.Schemas.UpdatePersonalizationRequest(historyLearningEnabled: enabled)
        switch try await client.patchPersonalization(.init(body: .json(body))) {
        case .ok(let ok):
            let settings = try ok.body.json
            return try await AccountPersonalization(historyLearningEnabled: settings.historyLearningEnabled, insights: settings.historyLearningEnabled ? personalization().insights : [])
        case .default(let statusCode, let problem): throw APIFailure.from(problem, statusCode: statusCode)
        }
    }

    public func linkAppleIdentity(authorizationCode: String) async throws {
        let body = Components.Schemas.ComponentsSchemasProfileAppleIdentityLink(authorizationCode: authorizationCode)
        switch try await client.putAppleIdentity(.init(body: .json(body))) {
        case .noContent: return
        case .default(let statusCode, let problem): throw APIFailure.from(problem, statusCode: statusCode)
        }
    }

    public func deleteAccount(reason: String?) async throws -> AccountDeletionReceipt {
        let body = reason.map { Components.Schemas.ComponentsSchemasProfileAccountDeletionRequest(reason: $0) }
        switch try await client.deleteMe(.init(body: body.map { .json($0) })) {
        case .accepted(let accepted):
            let receipt = try accepted.body.json
            return AccountDeletionReceipt(status: receipt.status.rawValue, requestedAt: receipt.requestedAt, purgeAfter: receipt.purgeAfter)
        case .default(let statusCode, let problem): throw APIFailure.from(problem, statusCode: statusCode)
        }
    }
}

extension AccountProfile {
    init(_ profile: Components.Schemas.Profile) {
        self.init(
            userID: profile.userId, preferredLanguage: profile.preferredLanguage, locale: profile.locale,
            homeCountry: profile.homeCountry, departureAirport: profile.departureAirport,
            preferredCurrency: profile.preferredCurrency, units: profile.units.rawValue, hourCycle: profile.hourCycle.rawValue
        )
    }
}

extension AccountPreferences {
    init(_ preferences: Components.Schemas.Preferences) {
        self.init(
            travelStyle: preferences.travelStyle, weather: preferences.weatherPreference.rawValue,
            crowds: preferences.crowdPreference.rawValue, maxFlightMinutes: preferences.maxFlightMinutes,
            pace: preferences.tripPace.rawValue, interests: preferences.interests,
            travelingWithChildren: preferences.travelingWithChildren
        )
    }
}

extension LearnedPreference {
    init(_ insight: Components.Schemas.PersonalizationInsight) {
        self.init(
            id: insight.id, key: insight.key, summaryCode: insight.summaryCode, confidence: insight.confidence.rawValue,
            evidenceCount: insight.evidenceCount, status: insight.status.rawValue, editable: insight.editable
        )
    }
}
