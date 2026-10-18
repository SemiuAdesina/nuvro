import NuvroModels
import Testing

@testable import NuvroTravel

@Suite("Alerts preferences and nearby states")
@MainActor
struct AlertsAndNearbyStateTests {
    @Test("alert preferences toggle and a refresh replaces alerts")
    func preferencesAndRefresh() async {
        let group = NotificationGroup(name: "Trip alerts", settings: [NotificationSetting(name: "Weather", isOn: true)])
        let alert = TripAlert(id: "w", priority: .now, kind: .weather, severity: .actionNeeded, title: "Rain", body: "", reason: "", actionLabel: "Adjust", destination: .replanning)
        let model = TripAlertsModel(alerts: [], preferences: [group], source: { [alert] })
        #expect(model.isOn(group.settings[0]))
        model.toggle(group.settings[0])
        #expect(!model.isOn(group.settings[0]))
        #expect(model.isAllClear)
        await model.refresh()
        #expect(model.alerts.count == 1)
        #expect(!model.isAllClear)
    }

    @Test("nearby filters can produce no results and adding reports once")
    func nearby() {
        let place = NearbyPlace(id: "m", name: "Mercado", category: .food, categoryLabel: "Market", walkMinutes: 8, travelLabel: "8 min walk", isOpen: true, hoursLabel: "Open now", reason: "", media: DestinationMedia(placeholder: "gold", altText: "Mercado"), coordinate: GeoCoordinate(latitude: 0, longitude: 0), markerLetter: "F")
        let model = ExploreNearbyModel(places: [place], freeTime: FreeTimeWindow(minutes: 120, untilLabel: "dinner"), locationName: "Funchal", destinationName: "Madeira")
        #expect(model.hasResults)
        model.toggleFilter(.views)
        #expect(!model.hasResults)
        model.toggleFilter(.views)
        model.addToToday(place)
        #expect(model.lastEvent == "Added to today")
        model.addToToday(place)
        #expect(model.addedIDs.count == 1)
        #expect(model.directionsURL(for: place)?.host() == "maps.apple.com")
        #expect(model.directionsURL(for: place)?.query()?.contains("daddr=") == true)
    }
}
