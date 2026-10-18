import Foundation

public enum PlanningModule: String, CaseIterable, Sendable, Identifiable {
    case stay = "Stay"
    case flights = "Flights"
    case thingsToDo = "Things to do"
    case gettingAround = "Getting around"
    case packing = "Packing"
    case documents = "Documents"
    case esim = "eSIM"
    case weather = "Weather"

    public var id: String { rawValue }
    public var title: String { rawValue }

    public var systemImage: String {
        switch self {
        case .stay: "bed.double.fill"
        case .flights: "airplane"
        case .thingsToDo: "sparkles"
        case .gettingAround: "car.fill"
        case .packing: "suitcase.fill"
        case .documents: "doc.text.fill"
        case .esim: "simcard.fill"
        case .weather: "cloud.sun.fill"
        }
    }
}
