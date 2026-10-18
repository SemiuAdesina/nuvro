import Foundation

public struct DestinationMedia: Sendable, Hashable, Codable {
    public let url: URL?
    public let placeholder: String
    public let altText: String
    public let attribution: String?
    public let provider: String?
    public let retrievedAt: Date?

    public init(url: URL? = nil, placeholder: String, altText: String, attribution: String? = nil, provider: String? = nil, retrievedAt: Date? = nil) {
        self.url = url
        self.placeholder = placeholder
        self.altText = altText
        self.attribution = attribution
        self.provider = provider
        self.retrievedAt = retrievedAt
    }

    public var hasPhoto: Bool { url != nil }

    private enum CodingKeys: String, CodingKey {
        case url, placeholder, attribution, provider
        case altText = "alt_text"
        case retrievedAt = "retrieved_at"
    }
}
