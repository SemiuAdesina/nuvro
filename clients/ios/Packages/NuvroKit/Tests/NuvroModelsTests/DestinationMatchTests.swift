import Foundation
import Testing

@testable import NuvroModels

@Suite("DestinationMatch")
struct DestinationMatchTests {
    private let decoder = JSONDecoder()

    private let payload = Data(#"""
    {
      "id": "madeira",
      "name": "Madeira",
      "country": "Portugal",
      "region": "Atlantic island",
      "photo_reference": "madeira",
      "score": 94,
      "estimate": { "low": {"amount_minor": 264000, "currency": "USD"}, "high": {"amount_minor": 294000, "currency": "USD"} },
      "weather": { "condition": "Warm", "temperature_celsius": 23 },
      "flight": { "stops": 1, "approximate_hours": 9 },
      "duration_days": 7,
      "pace": "Easygoing",
      "tags": ["Coastal", "Food", "Relaxing"],
      "reasons": [
        { "text": "Within your budget", "verdict": "fits" },
        { "text": "Longer flight than preferred", "verdict": "tradeoff" }
      ]
    }
    """#.utf8)

    @Test("decodes the contract payload")
    func decodes() throws {
        let match = try decoder.decode(DestinationMatch.self, from: payload)

        #expect(match.displayName == "Madeira, Portugal")
        #expect(match.score.percent == 94)
        #expect(match.estimate?.low.amountMinor == 264000)
        #expect(match.weather?.temperatureCelsius == 23)
        #expect(match.flight?.stops == 1)
        #expect(match.reasons.count == 2)
        #expect(match.reasons[1].isPositive == false)
    }

    @Test("a score outside 0 to 100 is rejected")
    func rejectsBadScore() {
        #expect(throws: MatchScore.Failure.outOfRange(101)) { _ = try MatchScore(percent: 101) }
        #expect(throws: MatchScore.Failure.outOfRange(-1)) { _ = try MatchScore(percent: -1) }
        #expect(throws: DecodingError.self) { _ = try decoder.decode(MatchScore.self, from: Data("120".utf8)) }
    }

    @Test("score exposes a fraction for rings and a label for badges")
    func scoreDerivatives() throws {
        let score = try MatchScore(percent: 94)

        #expect(abs(score.fraction - 0.94) < 0.0001)
        #expect(score.label == "94%")
    }

    @Test("round trips through snake case keys")
    func roundTrips() throws {
        let match = try decoder.decode(DestinationMatch.self, from: payload)
        let encoded = try JSONEncoder().encode(match)
        let text = String(decoding: encoded, as: UTF8.self)

        #expect(text.contains("photo_reference"))
        #expect(text.contains("duration_days"))
        #expect(try decoder.decode(DestinationMatch.self, from: encoded) == match)
    }
}
