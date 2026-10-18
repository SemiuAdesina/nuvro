import Foundation

public enum OnboardingStage: Int, CaseIterable, Sendable, Comparable {
    case splash = 0
    case discover = 1
    case plan = 2
    case travel = 3

    public static func < (lhs: OnboardingStage, rhs: OnboardingStage) -> Bool {
        lhs.rawValue < rhs.rawValue
    }

    public var isScreen: Bool {
        self != .splash
    }

    public var pageIndex: Int? {
        isScreen ? rawValue - 1 : nil
    }

    public var next: OnboardingStage? {
        OnboardingStage(rawValue: rawValue + 1)
    }

    public var previousScreen: OnboardingStage? {
        guard let candidate = OnboardingStage(rawValue: rawValue - 1), candidate.isScreen else { return nil }
        return candidate
    }

    public static let screens: [OnboardingStage] = [.discover, .plan, .travel]
}

public enum OnboardingExit: Equatable, Sendable {
    case findVacation
    case existingTrip
    case skipped
}

public enum HeadlineVariant: String, Sendable, CaseIterable {
    case main
    case alternate
}
