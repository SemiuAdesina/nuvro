import Foundation
import NuvroModels
import Testing

@testable import NuvroAPI

@Suite("NuvroAccountGateway")
struct AccountGatewayTests {
    private let profile = #"{"accessibilityPreferences":[],"departureAirport":"LOS","homeCountry":"NG","homeTimezone":null,"hourCycle":"24","locale":"en","preferredCurrency":"NGN","preferredLanguage":"en","units":"metric","userId":"3f1c9c8e-0e3a-4f0a-9b1f-1a2b3c4d5e6f"}"#
    private let preferences = #"{"travelStyle":["beach"],"weatherPreference":"warm","crowdPreference":"quiet","maxFlightMinutes":420,"tripPace":"relaxed","interests":["food","nature"],"travelingWithChildren":false,"scope":"any","entrySimplicity":"any"}"#
    private let settings = #"{"explicit":[],"generatedAt":"2026-09-09T12:00:00Z","historyLearningEnabled":true,"learned":[],"rulesVersion":"personalization-v1","suppressed":[]}"#
    private let insights = #"{"insights":[{"id":"i-1","key":"interest.coastal","category":"interest","summaryCode":"insight.interest.coastal","confidence":"medium","evidenceCount":3,"status":"active","editable":true}]}"#

    @Test("maps the profile into a plain value")
    func loadsProfile() async throws {
        let transport = StubTransport((.ok, profile))
        let loaded = try await NuvroAccountGateway(client: testClient(transport)).profile()
        #expect(loaded.homeCountry == "NG")
        #expect(loaded.departureAirport == "LOS")
        #expect(loaded.preferredCurrency == "NGN")
        #expect(loaded.units == "metric")
        #expect(transport.exchanges.first?.headers[.authorization] == "Bearer access-1")
    }

    @Test("maps preferences including the nullable flight ceiling")
    func loadsPreferences() async throws {
        let transport = StubTransport((.ok, preferences))
        let loaded = try await NuvroAccountGateway(client: testClient(transport)).preferences()
        #expect(loaded.maxFlightMinutes == 420)
        #expect(loaded.pace == "relaxed")
        #expect(loaded.interests == ["food", "nature"])
        #expect(loaded.crowds == "quiet")
    }

    @Test("personalization combines the learning switch with the insight list")
    func loadsPersonalization() async throws {
        let transport = StubTransport(["/v1/me/personalization": (.ok, settings), "/v1/me/personalization/insights": (.ok, insights)])
        let loaded = try await NuvroAccountGateway(client: testClient(transport)).personalization()
        #expect(loaded.historyLearningEnabled)
        #expect(loaded.insights.count == 1)
        #expect(loaded.insights.first?.summaryCode == "insight.interest.coastal")
        #expect(loaded.insights.first?.evidenceCount == 3)
        #expect(Set(transport.exchanges.map(\.path)) == ["/v1/me/personalization", "/v1/me/personalization/insights"])
    }

    @Test("turning learning off patches the switch and reports no insights")
    func disablesLearning() async throws {
        let disabled = settings.replacingOccurrences(of: "\"historyLearningEnabled\":true", with: "\"historyLearningEnabled\":false")
        let transport = StubTransport((.ok, disabled))
        let updated = try await NuvroAccountGateway(client: testClient(transport)).setHistoryLearning(false)
        #expect(!updated.historyLearningEnabled)
        #expect(updated.insights.isEmpty)
        #expect(transport.exchanges.first?.method == .patch)
        #expect(transport.exchanges.first?.json["historyLearningEnabled"] as? Bool == false)
    }

    @Test("closing the account returns the erasure schedule")
    func deletesAccount() async throws {
        let receipt = #"{"status":"pending","requestedAt":"2026-09-09T12:00:00Z","purgeAfter":"2026-10-09T12:00:00Z"}"#
        let transport = StubTransport((.accepted, receipt))
        let result = try await NuvroAccountGateway(client: testClient(transport)).deleteAccount(reason: "no_longer_travelling")
        #expect(result.status == "pending")
        #expect(result.purgeAfter > result.requestedAt)
        #expect(transport.exchanges.first?.json["reason"] as? String == "no_longer_travelling")
    }

    @Test("a closed account is refused with the contract's failure code")
    func refusedWhenClosed() async {
        let transport = StubTransport(StubTransport.problem(403, "forbidden"))
        do {
            _ = try await NuvroAccountGateway(client: testClient(transport)).profile()
            Issue.record("a forbidden profile read succeeded")
        } catch let failure as APIFailure {
            #expect(failure.code == .forbidden)
            #expect(failure.isAuthenticationFailure)
        } catch {
            Issue.record("unexpected error \(error)")
        }
    }
}
