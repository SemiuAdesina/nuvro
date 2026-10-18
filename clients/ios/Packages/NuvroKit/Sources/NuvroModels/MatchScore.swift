import Foundation

public struct MatchScore: Sendable, Hashable, Codable {
    public enum Failure: Error, Equatable, Sendable {
        case outOfRange(Int)
    }

    public let percent: Int

    public init(percent: Int) throws {
        guard (0...100).contains(percent) else { throw Failure.outOfRange(percent) }
        self.percent = percent
    }

    public init(from decoder: any Decoder) throws {
        let container = try decoder.singleValueContainer()
        let value = try container.decode(Int.self)
        guard (0...100).contains(value) else {
            throw DecodingError.dataCorruptedError(in: container, debugDescription: "Match score must be between 0 and 100.")
        }
        self.percent = value
    }

    public func encode(to encoder: any Encoder) throws {
        var container = encoder.singleValueContainer()
        try container.encode(percent)
    }

    public var fraction: Double { Double(percent) / 100 }
    public var label: String { "\(percent)%" }
}

public struct MatchReason: Sendable, Hashable, Codable, Identifiable {
    public enum Verdict: String, Sendable, Codable {
        case fits
        case tradeoff
    }

    public let text: String
    public let verdict: Verdict

    public init(text: String, verdict: Verdict) {
        self.text = text
        self.verdict = verdict
    }

    public var id: String { text }
    public var isPositive: Bool { verdict == .fits }
}
