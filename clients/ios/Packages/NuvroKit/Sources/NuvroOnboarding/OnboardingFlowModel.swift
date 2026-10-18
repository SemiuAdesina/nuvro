import Foundation
import NuvroLocalization
import Observation

@Observable
@MainActor
public final class OnboardingFlowModel {
    public private(set) var stage: OnboardingStage = .splash
    public private(set) var hasBooted = false
    public private(set) var hasReplanned = false
    public private(set) var dragOffset: CGFloat = 0
    public private(set) var exit: OnboardingExit?
    public var language: Language
    public var isLanguageSheetPresented = false
    public var languageQuery = ""

    public let reduceMotion: Bool
    public let headlineVariant: HeadlineVariant
    private let onExit: @MainActor (OnboardingExit) -> Void

    public init(
        language: Language = .english,
        reduceMotion: Bool = false,
        headlineVariant: HeadlineVariant = .main,
        startingStage: OnboardingStage = .splash,
        onExit: @escaping @MainActor (OnboardingExit) -> Void = { _ in }
    ) {
        self.language = language
        self.reduceMotion = reduceMotion
        self.headlineVariant = headlineVariant
        self.stage = startingStage
        self.onExit = onExit
    }

    public var copy: OnboardingCopy {
        .resolve(language)
    }

    public var isShowingScreenControls: Bool {
        stage.isScreen && exit == nil
    }

    public var isShowingLanguagePill: Bool {
        exit == nil
    }

    public var callToActionLabel: String {
        switch stage {
        case .splash, .discover: copy.callToAction1
        case .plan: copy.callToAction2
        case .travel: copy.callToAction3
        }
    }

    public var headline3: String {
        headlineVariant == .alternate ? copy.supporting3 : copy.headline3
    }

    public var supporting3: String {
        headlineVariant == .alternate ? copy.headline3 : copy.supporting3
    }

    public var filteredLanguages: [Language] {
        Language.matching(languageQuery)
    }

    public var cameraPose: CameraPose {
        CameraPose.pose(for: stage, reduceMotion: reduceMotion)
    }

    public func boot() {
        hasBooted = true
    }

    public func splashDidElapse() {
        guard stage == .splash, exit == nil else { return }
        move(to: .discover)
    }

    public func replanDidElapse() {
        guard stage == .travel, exit == nil else { return }
        hasReplanned = true
    }

    public func advance() {
        guard exit == nil else { return }
        if let next = stage.next {
            move(to: next)
        } else {
            finish(.findVacation)
        }
    }

    public func retreat() {
        guard exit == nil, let previous = stage.previousScreen else { return }
        move(to: previous)
    }

    public func skip() {
        guard stage.isScreen else { return }
        finish(.skipped)
    }

    public func chooseExistingTrip() {
        guard stage == .travel else { return }
        finish(.existingTrip)
    }

    public func presentLanguageSheet() {
        guard exit == nil else { return }
        isLanguageSheetPresented = true
    }

    public func dismissLanguageSheet() {
        isLanguageSheetPresented = false
        languageQuery = ""
    }

    public func select(_ language: Language) {
        self.language = language
        dismissLanguageSheet()
    }

    public func dragChanged(_ translation: CGFloat) {
        guard canDrag, abs(translation) > StagePoses.dragDeadZone else { return }
        dragOffset = max(-StagePoses.dragLimit, min(StagePoses.dragLimit, translation))
    }

    public func dragEnded(_ translation: CGFloat) {
        let wasDragging = canDrag
        dragOffset = 0
        guard wasDragging else { return }
        if translation < -StagePoses.swipeThreshold, stage < .travel {
            advance()
        } else if translation > StagePoses.swipeThreshold, stage > .discover {
            retreat()
        }
    }

    private var canDrag: Bool {
        stage.isScreen && !isLanguageSheetPresented && exit == nil
    }

    private func move(to next: OnboardingStage) {
        stage = next
        hasReplanned = false
    }

    private func finish(_ reason: OnboardingExit) {
        guard exit == nil else { return }
        exit = reason
        onExit(reason)
    }
}
