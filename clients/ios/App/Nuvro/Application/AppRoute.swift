import Foundation
import NuvroModels

enum AppRoute: Hashable {
    case matchSetup
    case matchResults
    case destination(DestinationMatch)
    case tripPlan
}

enum TripRoute: Hashable {
    case itinerary
    case budget
    case stay
    case travelMode
    case replan
    case language
    case nearby
    case wallet
    case alerts
    case summary
    case hub
    case flights
    case experiences
    case transport
    case checklist
    case group
    case practical
}

enum SavedRoute: Hashable {
    case compare
}

enum ProfileRoute: Hashable {
    case region
    case settings
}

enum AppTab: Hashable {
    case discover
    case trips
    case saved
    case profile
}
