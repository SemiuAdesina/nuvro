import NuvroDesignSystem
import NuvroModels
import NuvroPermissions
import SwiftUI

private enum TravelPreviewFixtures {
    static let alert = TripAlert(id: "rain", priority: .now, kind: .weather, severity: .actionNeeded, title: "Rain expected from 3 PM", body: "Your Botanical Garden visit may be affected.", reason: "Affects your Day 3 itinerary", actionLabel: "Adjust plan", destination: .replanning)
    static let documents = [
        TripDocument(id: "cable", category: .activity, title: "Funchal Cable Car", subtitle: "Today · 10:00", whenLabel: "Today · 10:00", travelers: 2, isAvailableOffline: true, reference: "CBL-4471", canAddToWallet: true, isNeededNext: true),
        TripDocument(id: "museum", category: .activity, title: "Madeira Story Centre", subtitle: "Used", whenLabel: "Oct 13", travelers: 2, isAvailableOffline: true, reference: "MSC-1180", isExpired: true),
        TripDocument(id: "transfer", category: .transfer, title: "Airport transfer", subtitle: "Oct 12 · 17:05", whenLabel: "Oct 12", travelers: 2, isAvailableOffline: false, reference: "TRF-6610", hasBarcode: false)
    ]
}

#Preview("Alerts, notifications denied") {
    DesignCanvas(fullBleed: false) { _ in
        NavigationStack { TripAlertsView(model: TripAlertsModel(alerts: [TravelPreviewFixtures.alert]), onOpen: { _ in }) }
    }
    .environment(PermissionCenter(authority: ScriptedPermissionAuthority(states: [.notifications: .denied])))
    .task { }
}

#Preview("Alerts all clear") {
    DesignCanvas(fullBleed: false) { _ in
        NavigationStack { TripAlertsView(model: TripAlertsModel(alerts: []), onOpen: { _ in }) }
    }
}

#Preview("Wallet with expired ticket and missing barcode") {
    DesignCanvas(fullBleed: false) { _ in
        NavigationStack { TripWalletView(model: TripWalletModel(documents: TravelPreviewFixtures.documents), destinationName: "Madeira", dateRange: "Oct 12–19", onNext: {}) }
    }
}

#Preview("Wallet empty") {
    DesignCanvas(fullBleed: false) { _ in
        NavigationStack { TripWalletView(model: TripWalletModel(documents: []), destinationName: "Madeira", dateRange: "Oct 12–19", onNext: {}) }
    }
}
