public enum TravelSetupStep: Int, CaseIterable, Sendable, Identifiable {
    case welcome
    case residence
    case passports
    case permits
    case airport
    case currency
    case paperwork
    case interests
    case pace
    case weather
    case crowds
    case flights
    case stays
    case group
    case accessibility
    case review

    public var id: Int { rawValue }

    public var isOptional: Bool {
        switch self {
        case .permits, .accessibility, .flights, .stays: true
        default: false
        }
    }

    public var countsTowardsProgress: Bool {
        self != .welcome && self != .review
    }

    public var next: TravelSetupStep? { TravelSetupStep(rawValue: rawValue + 1) }
    public var previous: TravelSetupStep? { TravelSetupStep(rawValue: rawValue - 1) }
}
