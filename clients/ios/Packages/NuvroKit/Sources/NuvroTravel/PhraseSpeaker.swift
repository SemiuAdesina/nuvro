import AVFoundation
import Foundation

@MainActor
public final class PhraseSpeaker {
    private let synthesizer = AVSpeechSynthesizer()

    public init() {}

    public func speak(_ text: String, languageName: String) {
        let utterance = AVSpeechUtterance(string: text)
        utterance.voice = Self.voice(for: languageName)
        synthesizer.stopSpeaking(at: .immediate)
        synthesizer.speak(utterance)
    }

    static func voice(for languageName: String) -> AVSpeechSynthesisVoice? {
        let voices = AVSpeechSynthesisVoice.speechVoices()
        guard let language = language(named: languageName, among: voices.map(\.language)) else { return nil }
        return voices.first { $0.language == language }
    }

    public static func language(named name: String, among identifiers: [String]) -> String? {
        identifiers.first { identifier in
            let code = Locale(identifier: identifier).language.languageCode?.identifier ?? ""
            return Locale(identifier: "en").localizedString(forLanguageCode: code)?.caseInsensitiveCompare(name) == .orderedSame
        }
    }
}
