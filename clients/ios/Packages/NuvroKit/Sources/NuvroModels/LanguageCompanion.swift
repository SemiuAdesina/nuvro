import Foundation

public struct LanguagePair: Sendable, Hashable, Codable {
    public let source: String
    public let target: String

    public init(source: String, target: String) {
        self.source = source
        self.target = target
    }

    public var reversed: LanguagePair { LanguagePair(source: target, target: source) }
}

public struct PhraseTranslation: Sendable, Hashable, Codable {
    public let pair: LanguagePair
    public let original: String
    public let translated: String

    public init(pair: LanguagePair, original: String, translated: String) {
        self.pair = pair
        self.original = original
        self.translated = translated
    }
}

public struct TravelPhrase: Sendable, Hashable, Codable, Identifiable {
    public let context: String
    public let systemImage: String
    public let placeholder: String
    public let original: String
    public let translated: String

    public init(context: String, systemImage: String, placeholder: String = "twilight", original: String, translated: String) {
        self.context = context
        self.systemImage = systemImage
        self.placeholder = placeholder
        self.original = original
        self.translated = translated
    }

    public var id: String { original }

    private enum CodingKeys: String, CodingKey {
        case context, original, translated, placeholder
        case systemImage = "system_image"
    }
}
