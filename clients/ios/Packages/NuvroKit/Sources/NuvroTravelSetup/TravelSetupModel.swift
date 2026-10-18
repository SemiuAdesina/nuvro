import Foundation
import NuvroDesignSystem
import NuvroLocalization
import NuvroModels
import Observation

public enum SetupPicker: String, Identifiable, Sendable {
    case residence, passport, permit, airport, currency

    public var id: String { rawValue }
}

@Observable
@MainActor
public final class TravelSetupModel {
    public private(set) var step: TravelSetupStep
    public private(set) var draft: TravelSetupDraft
    public private(set) var isMovingForward = true
    public private(set) var completedProfile: TravelProfile?
    public var picker: SetupPicker? { didSet { if picker == nil { replacing = nil } } }
    public private(set) var replacing: String?
    public let regions: [RegionOption]
    public let currencies: [CurrencyOption]
    public let airports: [Airport]
    public let editingStep: TravelSetupStep?
    public let copy: TravelSetupCopy
    public let options: TravelOptionCopy
    private let store: any TravelSetupStore
    private let locale: Locale

    public init(store: any TravelSetupStore, airports: [Airport], editing: TravelSetupStep? = nil, language: Language = .english, locale: Locale = .current) {
        self.store = store
        self.airports = airports
        self.editingStep = editing
        self.copy = .resolve(language)
        self.options = .resolve(language)
        self.locale = locale
        self.regions = TravelCatalog.regions(locale: locale)
        self.currencies = TravelCatalog.currencies(locale: locale)
        if let editing, let profile = store.loadProfile() {
            self.draft = TravelSetupDraft(profile: profile)
            self.step = editing
        } else {
            let restored = store.loadDraft() ?? TravelSetupDraft()
            self.draft = restored
            self.step = TravelSetupStep(rawValue: restored.stepIndex) ?? .welcome
        }
    }

    public var isEditing: Bool { editingStep != nil }
    public var progressSteps: [TravelSetupStep] { TravelSetupStep.allCases.filter(\.countsTowardsProgress) }
    public var progressIndex: Int { progressSteps.firstIndex(of: step).map { $0 + 1 } ?? 0 }
    public var progressFraction: Double { Double(progressIndex) / Double(progressSteps.count) }
    public var showsProgress: Bool { step.countsTowardsProgress && !isEditing }
    public var canGoBack: Bool { step.previous != nil && !isEditing }
    public var canSkip: Bool { step.isOptional && !isEditing }

    public var canContinue: Bool {
        switch step {
        case .welcome, .review, .permits, .flights, .stays, .accessibility: true
        case .residence: draft.countryOfResidence != nil
        case .passports: !draft.passportCountries.isEmpty || draft.hasDeclinedPassport
        case .airport: draft.departureAirport != nil
        case .currency: draft.currency != nil
        case .paperwork: draft.paperwork != nil
        case .interests: !draft.interests.isEmpty
        case .pace: draft.pace != nil
        case .weather: !draft.weather.isEmpty
        case .crowds: draft.crowds != nil
        case .group: draft.group != nil
        }
    }

    public var suggestedCurrency: CurrencyOption? {
        TravelCatalog.suggestedCurrency(forRegion: draft.countryOfResidence).flatMap { code in currencies.first { $0.code == code } }
    }

    public func region(_ code: String?) -> RegionOption? { code.flatMap { code in regions.first { $0.code == code } } }
    public func currency(_ code: String?) -> CurrencyOption? { code.flatMap { code in currencies.first { $0.code == code } } }
    public func airport(_ code: String?) -> Airport? { code.flatMap { code in airports.first { $0.code == code } } }

    public func update(_ change: (inout TravelSetupDraft) -> Void) {
        change(&draft)
        persist()
    }

    public func replacePassport(_ code: String) {
        replacing = code
        picker = .passport
    }

    public func replacePermit(_ code: String) {
        replacing = code
        picker = .permit
    }

    public func addPassport(_ code: String) {
        let replaced = replacing
        update { draft in
            if let replaced { draft.passportCountries.removeAll { $0 == replaced } }
            if !draft.passportCountries.contains(code) { draft.passportCountries.append(code) }
            draft.hasDeclinedPassport = false
        }
    }

    public func removePassport(_ code: String) {
        update { $0.passportCountries.removeAll { $0 == code } }
    }

    public func declinePassport() {
        update { $0.hasDeclinedPassport = true; $0.passportCountries = [] }
    }

    public func addPermit(_ code: String) {
        let replaced = replacing
        update { draft in
            if let replaced { draft.residencePermits.removeAll { $0 == replaced } }
            draft.permitAnswer = .yes
            if !draft.residencePermits.contains(code) { draft.residencePermits.append(code) }
        }
    }

    public func removePermit(_ code: String) {
        update { $0.residencePermits.removeAll { $0 == code } }
    }

    public func toggle<Value: Hashable>(_ value: Value, in keyPath: WritableKeyPath<TravelSetupDraft, Set<Value>>) {
        update { draft in
            if draft[keyPath: keyPath].contains(value) { draft[keyPath: keyPath].remove(value) } else { draft[keyPath: keyPath].insert(value) }
        }
    }

    public func selectExclusive<Value: Hashable>(_ value: Value, exclusive: Value, in keyPath: WritableKeyPath<TravelSetupDraft, Set<Value>>) {
        update { draft in
            if value == exclusive {
                draft[keyPath: keyPath] = draft[keyPath: keyPath].contains(exclusive) ? [] : [exclusive]
            } else {
                draft[keyPath: keyPath].remove(exclusive)
                if draft[keyPath: keyPath].contains(value) { draft[keyPath: keyPath].remove(value) } else { draft[keyPath: keyPath].insert(value) }
            }
        }
    }

    public func advance() {
        guard canContinue else { return }
        if isEditing || step == .review {
            finish()
            return
        }
        guard let next = step.next else { return }
        isMovingForward = true
        move(to: next)
    }

    public func skip() {
        guard canSkip, let next = step.next else { return }
        if step == .permits { update { $0.permitAnswer = .skipped } }
        isMovingForward = true
        move(to: next)
    }

    public func back() {
        guard canGoBack, let previous = step.previous else { return }
        isMovingForward = false
        move(to: previous)
    }

    public func edit(_ target: TravelSetupStep) {
        isMovingForward = false
        move(to: target)
    }

    public func returnToReview() {
        isMovingForward = true
        move(to: .review)
    }

    private func move(to target: TravelSetupStep) {
        step = target
        persist()
    }

    private func finish() {
        guard let profile = draft.profile() else { return }
        store.saveProfile(profile)
        store.clearDraft()
        completedProfile = profile
    }

    private func persist() {
        guard !isEditing else { return }
        draft.stepIndex = step.rawValue
        store.saveDraft(draft)
    }
}
