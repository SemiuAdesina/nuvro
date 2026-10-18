import Foundation
import NuvroModels
import Testing

@testable import NuvroAPI

@Suite("NuvroDiscoverGateway")
struct DiscoverGatewayTests {
    private let response = #"""
    {"algorithmVersion":"match-v1","generatedAt":"2026-09-10T16:44:59Z","missingContext":["passport"],"rankingVersion":"rank-v1","recommendations":[
      {"access":{"confidence":"unknown","entryMode":"unknown","preparation":"unknown","reasons":["entry_provider_unavailable"],"status":"unknown","transitRisk":"unknown"},
       "affordability":{"budget":{"amountMinor":300000,"currency":"EUR"},"confidence":"unknown","status":"insufficient_evidence"},
       "cautions":["entry_evidence_unknown"],
       "destination":{"coordinates":{"latitude":-8.4095,"longitude":115.1889},"countryCode":"ID","id":"bali-id","name":"Bali","primaryAirport":"DPS","region":"Bali","timezone":"Asia/Makassar",
         "media":{"url":"https://media.example.test/bali.jpg","width":1200,"height":800,"aspectRatio":1.5,"attribution":[],"altText":"Bali coast"}},
       "flight":{"confidence":"unknown","status":"unknown"},
       "personalMatch":{"algorithmVersion":"match-v1","score":64,"components":[
         {"assessment":"neutral","awardedPoints":12,"explanation":"interest_overlap","key":"interests","label":"Interests","possiblePoints":25},
         {"assessment":"strong","awardedPoints":15,"explanation":"weather_match","key":"weather","label":"Weather","possiblePoints":15},
         {"assessment":"strong","awardedPoints":5,"explanation":"profile_match","key":"pace","label":"Pace","possiblePoints":5}]},
       "rank":1,"reasons":["entry_evidence_unknown","weather_fit","pace_fit"],"summary":"Bali: weather and crowds fit this trip"},
      {"access":{"confidence":"unknown","entryMode":"unknown","preparation":"unknown","reasons":[],"status":"unknown","transitRisk":"unknown"},
       "affordability":{"budget":{"amountMinor":300000,"currency":"EUR"},"confidence":"high","status":"good_fit",
         "estimatedTrip":{"minimum":{"amountMinor":120000,"currency":"EUR"},"expected":{"amountMinor":180000,"currency":"EUR"},"maximum":{"amountMinor":240000,"currency":"EUR"}}},
       "cautions":[],
       "destination":{"coordinates":{"latitude":32.7607,"longitude":-16.9595},"countryCode":"PT","id":"madeira-pt","name":"Madeira","primaryAirport":"FNC","timezone":"Atlantic/Madeira"},
       "flight":{"confidence":"high","status":"available","stops":1,"durationMinutes":250},
       "personalMatch":{"algorithmVersion":"match-v1","score":49,"components":[]},
       "rank":2,"reasons":["crowds_fit"],"summary":"Madeira: crowds and pace fit this trip"}]}
    """#

    private func brief() -> DiscoverBrief {
        DiscoverBrief(budgetMinor: 300000, currency: "EUR", startDate: Date(timeIntervalSince1970: 1_791_000_000), endDate: Date(timeIntervalSince1970: 1_791_604_800), adults: 2, departureAirport: "LIS", passportCountries: ["NG"], interests: ["food"])
    }

    @Test("ranked recommendations become presentation matches with honest unknowns")
    func mapsRecommendations() async throws {
        let transport = StubTransport((.ok, response))
        let matches = try await NuvroDiscoverGateway(client: testClient(transport)).matches(brief())
        #expect(matches.map(\.id) == ["bali-id", "madeira-pt"])
        let bali = try #require(matches.first)
        #expect(bali.score.percent == 64)
        #expect(bali.country == "Indonesia")
        #expect(bali.photoURL?.absoluteString == "https://media.example.test/bali.jpg")
        #expect(bali.weather == nil)
        #expect(bali.flight == nil)
        #expect(bali.estimate == nil)
        #expect(bali.tags == ["Weather fit", "Pace fit"])
        #expect(bali.durationDays == 7)
        #expect(bali.reasons.filter(\.isPositive).count == 2)
        #expect(bali.reasons.contains { !$0.isPositive })
        #expect(bali.summary == "Bali: weather and crowds fit this trip")
    }

    @Test("evidence that exists is carried across")
    func mapsEvidence() async throws {
        let transport = StubTransport((.ok, response))
        let madeira = try #require(try await NuvroDiscoverGateway(client: testClient(transport)).matches(brief()).last)
        #expect(madeira.photoURL == nil)
        #expect(madeira.region == "Madeira")
        #expect(madeira.estimate?.low.amountMinor == 120000)
        #expect(madeira.estimate?.high.amountMinor == 240000)
        #expect(madeira.flight?.stops == 1)
        #expect(madeira.flight?.approximateHours == 4)
    }

    @Test("the request carries the brief the traveller set")
    func sendsTheBrief() async throws {
        let transport = StubTransport((.ok, response))
        _ = try await NuvroDiscoverGateway(client: testClient(transport)).matches(brief())
        let sent = try #require(transport.exchanges.first)
        #expect(sent.path == "/v1/discover/matches")
        #expect(sent.json["adults"] as? Int == 2)
        #expect(sent.json["startDate"] as? String == "2026-10-03")
        #expect(sent.json["endDate"] as? String == "2026-10-10")
        #expect((sent.json["budget"] as? [String: Any])?["currency"] as? String == "EUR")
        #expect(sent.json["passportCountries"] as? [String] == ["NG"])
        #expect(sent.json["departureAirport"] as? String == "LIS")
    }

    @Test("a refused search surfaces the contract failure")
    func surfacesFailure() async {
        let transport = StubTransport(StubTransport.problem(400, "validation_error"))
        do {
            _ = try await NuvroDiscoverGateway(client: testClient(transport)).matches(brief())
            Issue.record("a refused search returned matches")
        } catch let failure as APIFailure {
            #expect(failure.code == .validationError)
        } catch {
            Issue.record("unexpected error \(error)")
        }
    }
}
