import SwiftUI

#Preview("Loading, empty, error") {
    ScrollView {
        VStack(spacing: 20) {
            NuvroLoadingState("Loading…")
            NuvroEmptyState(symbol: "suitcase", title: "No trips yet", message: "Your next trip starts here.", action: ("Find a vacation", {}))
            NuvroErrorState(title: "Couldn’t load this right now", message: "NUVRO is temporarily unavailable.", retry: ("Try again", {}))
            NuvroErrorState(symbol: "wifi.slash", title: "You’re offline", message: "Connect to load this.", retry: ("Try again", {}))
            NuvroInlineError("That doesn’t look like an email address.")
            HStack(spacing: 16) {
                FreshnessLabel(tone: .fresh, text: "Live")
                FreshnessLabel(tone: .cached, text: "Saved copy")
                FreshnessLabel(tone: .stale, text: "May be out of date")
                FreshnessLabel(tone: .unavailable, text: "Unavailable")
            }
        }
        .padding(20)
    }
    .background(Palette.background)
}

#Preview("Banners and permissions") {
    VStack(spacing: 20) {
        NuvroOfflineBanner(title: "You’re offline", detail: "Some saved trip details are still available.")
        NuvroOfflineBanner(title: "Reconnecting…", isReconnecting: true)
        TransientBanner(message: TransientMessage(text: "Saved"))
        NuvroPermissionExplainer(symbol: "location.fill", title: "Find what’s around you", message: "Allow location to show nearby places and travel times.", primary: ("Continue", {}), secondary: ("Not now", {}))
        NuvroPermissionExplainer(symbol: "mic.fill", palette: .gold, title: "Microphone access is off", message: "Turn it on in Settings, or type what you want to say.", primary: ("Open Settings", {}), secondary: ("Type instead", {}))
    }
    .padding(20)
    .background(Palette.background)
}

#Preview("States at accessibility text size") {
    VStack(spacing: 20) {
        NuvroEmptyState(symbol: "bookmark", title: "Nothing saved yet", message: "Save destinations you like and compare them here.", action: ("Explore destinations", {}))
        NuvroPermissionExplainer(symbol: "bell.badge.fill", title: "Stay ahead of changes", message: "Get flight changes, weather shifts and leave reminders while you travel.", primary: ("Continue", {}), secondary: ("Not now", {}))
    }
    .padding(20)
    .background(Palette.background)
    .dynamicTypeSize(.accessibility3)
}
