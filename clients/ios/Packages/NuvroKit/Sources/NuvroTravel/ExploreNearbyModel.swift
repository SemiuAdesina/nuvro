import Foundation
import NuvroLocalization
import NuvroModels
import Observation

@Observable
@MainActor
public final class ExploreNearbyModel {
    public let places: [NearbyPlace]
    public let freeTime: FreeTimeWindow
    public let locationName: String
    public let destinationName: String
    public var filter: NearbyPlace.Category?
    public var isListPresentation = false
    public var usesDestinationCentre = false
    public var selectedID: String?
    public private(set) var lastEvent: String?

    public func clearEvent() {
        lastEvent = nil
    }
    public private(set) var addedIDs: Set<String> = []
    public private(set) var savedIDs: Set<String> = []
    public let copy: NearbyCopy

    public init(places: [NearbyPlace], freeTime: FreeTimeWindow, locationName: String, destinationName: String, language: Language = .english) {
        self.places = places
        self.freeTime = freeTime
        self.locationName = locationName
        self.destinationName = destinationName
        self.selectedID = places.first?.id
        self.copy = .resolve(language)
    }

    public var visiblePlaces: [NearbyPlace] {
        guard let filter else { return places }
        if filter == .saved { return places.filter { savedIDs.contains($0.id) } }
        return places.filter { $0.category == filter }
    }

    public var selected: NearbyPlace? {
        visiblePlaces.first { $0.id == selectedID } ?? visiblePlaces.first
    }

    public var freeTimeLabel: String {
        let hours = freeTime.minutes / 60
        let minutes = freeTime.minutes % 60
        return hours > 0 ? "\(hours)h \(minutes)m" : "\(minutes)m"
    }

    public func fits(_ place: NearbyPlace) -> Bool {
        freeTime.fits(place)
    }

    public func select(_ place: NearbyPlace) {
        selectedID = place.id
    }

    public func toggleFilter(_ category: NearbyPlace.Category) {
        filter = filter == category ? nil : category
    }

    public func addToToday(_ place: NearbyPlace) {
        guard !addedIDs.contains(place.id) else { return }
        addedIDs.insert(place.id)
        lastEvent = copy.addedToToday
    }

    public var hasResults: Bool { !visiblePlaces.isEmpty }

    public func toggleSaved(_ place: NearbyPlace) {
        if savedIDs.contains(place.id) { savedIDs.remove(place.id) } else { savedIDs.insert(place.id) }
    }

    public func isAdded(_ place: NearbyPlace) -> Bool { addedIDs.contains(place.id) }
    public func isSaved(_ place: NearbyPlace) -> Bool { savedIDs.contains(place.id) }

    public func label(for category: NearbyPlace.Category) -> String {
        switch category {
        case .food: "Food"
        case .coffee: "Coffee"
        case .thingsToDo: "Things to do"
        case .views: "Views"
        case .shopping: "Shopping"
        case .indoor: "Indoor"
        case .saved: "Saved"
        }
    }

    public func directionsURL(for place: NearbyPlace) -> URL? {
        URL(string: "https://maps.apple.com/?daddr=\(place.coordinate.latitude),\(place.coordinate.longitude)&dirflg=w")
    }
}
