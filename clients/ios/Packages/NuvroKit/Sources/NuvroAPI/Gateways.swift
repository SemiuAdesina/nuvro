import Foundation
import NuvroModels

public struct IssuedSession: Sendable, Hashable {
    public let accessToken: String
    public let refreshToken: String
    public let expiresAt: Date
    public let accountID: String

    public init(accessToken: String, refreshToken: String, expiresAt: Date, accountID: String) {
        self.accessToken = accessToken
        self.refreshToken = refreshToken
        self.expiresAt = expiresAt
        self.accountID = accountID
    }
}

public protocol SessionGateway: Sendable {
    func createSession(identityToken: String, authorizationCode: String?) async throws -> IssuedSession
    func refreshSession(refreshToken: String) async throws -> IssuedSession
    func deleteSession() async throws
}

public struct AccountProfile: Sendable, Hashable {
    public let userID: String
    public let preferredLanguage: String
    public let locale: String
    public let homeCountry: String?
    public let departureAirport: String?
    public let preferredCurrency: String
    public let units: String
    public let hourCycle: String

    public init(userID: String, preferredLanguage: String, locale: String, homeCountry: String?, departureAirport: String?, preferredCurrency: String, units: String, hourCycle: String) {
        self.userID = userID
        self.preferredLanguage = preferredLanguage
        self.locale = locale
        self.homeCountry = homeCountry
        self.departureAirport = departureAirport
        self.preferredCurrency = preferredCurrency
        self.units = units
        self.hourCycle = hourCycle
    }
}

public struct AccountPreferences: Sendable, Hashable {
    public let travelStyle: [String]
    public let weather: String
    public let crowds: String
    public let maxFlightMinutes: Int?
    public let pace: String
    public let interests: [String]
    public let travelingWithChildren: Bool

    public init(travelStyle: [String], weather: String, crowds: String, maxFlightMinutes: Int?, pace: String, interests: [String], travelingWithChildren: Bool) {
        self.travelStyle = travelStyle
        self.weather = weather
        self.crowds = crowds
        self.maxFlightMinutes = maxFlightMinutes
        self.pace = pace
        self.interests = interests
        self.travelingWithChildren = travelingWithChildren
    }
}

public struct LearnedPreference: Sendable, Hashable, Identifiable {
    public let id: String
    public let key: String
    public let summaryCode: String
    public let confidence: String
    public let evidenceCount: Int
    public let status: String
    public let editable: Bool

    public init(id: String, key: String, summaryCode: String, confidence: String, evidenceCount: Int, status: String, editable: Bool) {
        self.id = id
        self.key = key
        self.summaryCode = summaryCode
        self.confidence = confidence
        self.evidenceCount = evidenceCount
        self.status = status
        self.editable = editable
    }
}

public struct AccountPersonalization: Sendable, Hashable {
    public let historyLearningEnabled: Bool
    public let insights: [LearnedPreference]

    public init(historyLearningEnabled: Bool, insights: [LearnedPreference]) {
        self.historyLearningEnabled = historyLearningEnabled
        self.insights = insights
    }
}

public struct AccountDeletionReceipt: Sendable, Hashable {
    public let status: String
    public let requestedAt: Date
    public let purgeAfter: Date

    public init(status: String, requestedAt: Date, purgeAfter: Date) {
        self.status = status
        self.requestedAt = requestedAt
        self.purgeAfter = purgeAfter
    }
}

public protocol AccountGateway: Sendable {
    func profile() async throws -> AccountProfile
    func preferences() async throws -> AccountPreferences
    func personalization() async throws -> AccountPersonalization
    func setHistoryLearning(_ enabled: Bool) async throws -> AccountPersonalization
    func linkAppleIdentity(authorizationCode: String) async throws
    func deleteAccount(reason: String?) async throws -> AccountDeletionReceipt
}
