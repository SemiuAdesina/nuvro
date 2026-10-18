import Foundation
import NuvroLocalization
import NuvroModels
import Observation

public typealias SuggestionSource = @Sendable () async throws -> [DestinationMatch]

@Observable
@MainActor
public final class DiscoverHomeModel {
    public private(set) var suggestionsState: LoadState<[DestinationMatch]>
    public private(set) var selectedThemes: Set<InspirationTheme> = []
    public private(set) var savedDestinationIDs: Set<String> = []
    public private(set) var lastEvent: String?

    public func clearEvent() {
        lastEvent = nil
    }
    public let copy: DiscoverCopy
    public let locale: Locale
    private let source: SuggestionSource?
    private var replacementSource: SuggestionSource?

    public init(suggestions: [DestinationMatch], source: SuggestionSource? = nil, language: Language = .english, locale: Locale = .current) {
        self.suggestionsState = .loaded(suggestions)
        self.source = source
        self.copy = .resolve(language)
        self.locale = locale
    }

    public init(state: LoadState<[DestinationMatch]>, source: SuggestionSource? = nil, language: Language = .english, locale: Locale = .current) {
        self.suggestionsState = state
        self.source = source
        self.copy = .resolve(language)
        self.locale = locale
    }

    public var suggestions: [DestinationMatch] { suggestionsState.value ?? [] }

    public func refresh() async {
        guard let source = replacementSource ?? source else { return }
        suggestionsState = suggestionsState.beginning()
        do {
            suggestionsState = .loaded(try await source())
        } catch {
            suggestionsState = suggestionsState.failing(LoadFailure(error: error))
        }
    }

    public func greeting(hour: Int) -> String {
        switch hour {
        case ..<12: copy.goodMorning
        case 12..<18: copy.goodAfternoon
        default: copy.goodEvening
        }
    }

    public func toggle(_ theme: InspirationTheme) {
        if selectedThemes.contains(theme) {
            selectedThemes.remove(theme)
        } else {
            selectedThemes.insert(theme)
        }
    }

    public func isSelected(_ theme: InspirationTheme) -> Bool {
        selectedThemes.contains(theme)
    }

    public func toggleSaved(_ destination: DestinationMatch) {
        if savedDestinationIDs.contains(destination.id) {
            savedDestinationIDs.remove(destination.id)
            lastEvent = copy.removedFromSaved
        } else {
            savedDestinationIDs.insert(destination.id)
            lastEvent = copy.saved
        }
    }

    public func isSaved(_ destination: DestinationMatch) -> Bool {
        savedDestinationIDs.contains(destination.id)
    }

    public func meta(for destination: DestinationMatch) -> String {
        var parts: [String] = []
        if let weather = destination.weather { parts.append(weather.condition) }
        if let tag = destination.tags.first { parts.append(tag) }
        if let estimate = destination.estimate { parts.append("\(copy.fromPrice) \(estimate.low.formattedWhole(locale: locale))") }
        return parts.joined(separator: " · ")
    }

    public func reason(for destination: DestinationMatch) -> String? {
        destination.reasons.first(where: \.isPositive)?.text ?? destination.summary
    }

    public func load(from source: @escaping SuggestionSource) async {
        replacementSource = source
        suggestionsState = .loading
        await refresh()
    }
}
