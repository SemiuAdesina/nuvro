import Foundation
import NuvroLocalization
import NuvroModels
import Observation

public enum CompanionMode: String, CaseIterable, Sendable {
    case speak, type, camera, phrases
}

@Observable
@MainActor
public final class LanguageCompanionModel {
    public private(set) var pair: LanguagePair
    public var mode: CompanionMode = .speak
    public private(set) var isListening = false
    public private(set) var result: PhraseTranslation?
    public var isShowingLarge = false
    public var typedText = ""
    public private(set) var lastEvent: String?

    public func clearEvent() {
        lastEvent = nil
    }
    public let phrases: [TravelPhrase]
    public let copy: LanguageCopy
    private let sample: PhraseTranslation

    public init(pair: LanguagePair, sample: PhraseTranslation, phrases: [TravelPhrase], language: Language = .english) {
        self.pair = pair
        self.sample = sample
        self.phrases = phrases
        self.copy = .resolve(language)
    }

    public var targetLanguage: String { pair.target }

    public func swapLanguages() {
        pair = pair.reversed
        result = result.map { PhraseTranslation(pair: $0.pair.reversed, original: $0.translated, translated: $0.original) }
    }

    public func beginListening() {
        isListening = true
        result = nil
    }

    public func toggleListening() {
        if isListening {
            finishListening()
        } else {
            beginListening()
            Task { [weak self] in
                try? await Task.sleep(for: .seconds(1.4))
                self?.finishListening()
            }
        }
    }

    public func showSample() {
        result = PhraseTranslation(pair: pair, original: sample.original, translated: sample.translated)
    }

    public func finishListening() {
        guard isListening else { return }
        isListening = false
        result = PhraseTranslation(pair: pair, original: sample.original, translated: sample.translated)
    }

    public func show(_ phrase: TravelPhrase) {
        result = PhraseTranslation(pair: pair, original: phrase.original, translated: phrase.translated)
        isShowingLarge = true
    }

    public func copyTranslation() -> String? {
        guard let result else { return nil }
        lastEvent = copy.copied
        return result.translated
    }

    public func translateTyped() {
        let text = typedText.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !text.isEmpty else { return }
        let match = phrases.first { $0.original.caseInsensitiveCompare(text) == .orderedSame }
        result = PhraseTranslation(pair: pair, original: text, translated: match?.translated ?? sample.translated)
    }
}
