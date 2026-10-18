import Foundation

public enum GuideCategory: String, CaseIterable, Sendable, Identifiable {
    case stay = "Stay"
    case thingsToDo = "Things to do"
    case food = "Food"
    case areas = "Areas"
    case gettingAround = "Getting around"
    case weather = "Weather"
    case budget = "Budget"
    case entry = "Entry"
    case safety = "Safety"
    case money = "Money"
    case internet = "Internet"
    case language = "Language"
    case localCustoms = "Local customs"
    case events = "Events"

    public var id: String { rawValue }
    public var title: String { rawValue }

    public var systemImage: String {
        switch self {
        case .stay: "bed.double.fill"
        case .thingsToDo: "sparkles"
        case .food: "fork.knife"
        case .areas: "map.fill"
        case .gettingAround: "bus.fill"
        case .weather: "sun.max.fill"
        case .budget: "banknote.fill"
        case .entry: "person.text.rectangle.fill"
        case .safety: "shield.fill"
        case .money: "creditcard.fill"
        case .internet: "wifi"
        case .language: "character.bubble.fill"
        case .localCustoms: "hand.wave.fill"
        case .events: "calendar"
        }
    }
}
