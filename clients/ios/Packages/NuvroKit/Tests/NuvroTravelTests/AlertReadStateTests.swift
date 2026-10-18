import NuvroModels
import Testing

@testable import NuvroTravel

@Suite("Alert read state")
@MainActor
struct AlertReadStateTests {
    @Test("alerts start unread, opening marks them read, dismissing hides them")
    func readState() {
        let rain = TripAlert(id: "rain", priority: .now, kind: .weather, severity: .actionNeeded, title: "Rain", body: "", reason: "", actionLabel: "Adjust", destination: .replanning)
        let flight = TripAlert(id: "flight", priority: .today, kind: .flight, severity: .updated, title: "Flight", body: "", reason: "", actionLabel: "View", destination: .document)
        let model = TripAlertsModel(alerts: [rain, flight])
        #expect(model.unreadCount == 2)
        model.markRead(rain)
        #expect(!model.isUnread(rain))
        #expect(model.unreadCount == 1)
        model.dismiss(flight)
        #expect(!model.isUnread(flight))
        #expect(model.unreadCount == 0)
        #expect(model.activeCount == 1)
    }
}
