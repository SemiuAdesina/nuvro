import Foundation
import Testing

@testable import NuvroModels
@testable import NuvroProfile

@Suite("Profile models")
@MainActor
struct ProfileModelsTests {
    private var calendar: Calendar {
        var calendar = Calendar(identifier: .gregorian)
        calendar.timeZone = TimeZone(identifier: "UTC")!
        return calendar
    }

    private func region(_ units: UnitPreferences) -> RegionModel {
        var components = DateComponents()
        components.year = 2026; components.month = 9; components.day = 8
        let date = calendar.date(from: components)!
        return RegionModel(settings: RegionSettings(appLanguage: "English", spokenLanguages: "English, French", region: "United States", regionNote: "", currency: "USD", homeAirport: "JFK"), units: units, sampleDate: date, calendar: calendar)
    }

    @Test("the preview follows every unit choice")
    func preview() {
        let imperial = region(UnitPreferences(distance: .miles, temperature: .fahrenheit, clock: .twelveHour, date: .monthFirst))
        #expect(imperial.preview == ["72°F", "3.2 mi", "$2,780", "Sep 8, 2026", "7:30 PM"])

        let metric = region(UnitPreferences(distance: .kilometres, temperature: .celsius, clock: .twentyFourHour, date: .dayFirst))
        #expect(metric.preview == ["22°C", "5.1 km", "$2,780", "8 Sep 2026", "19:30"])
    }

    @Test("notification toggles flip one setting only")
    func notifications() {
        let group = NotificationGroup(name: "Trip alerts", settings: [NotificationSetting(name: "Weather", isOn: true), NotificationSetting(name: "Reservations", isOn: true)])
        let model = SettingsModel(groups: [], notifications: [group], permissions: [])

        model.toggle(group.settings[0], in: group)

        #expect(!model.isOn(group.settings[0]))
        #expect(model.isOn(group.settings[1]))
    }

    @Test("trip history state label follows the toggle")
    func history() {
        let profile = TravelerProfile(name: "Alex", initial: "A", homeLabel: "", styleChips: [], preferences: [], learned: [], usesTripHistory: true)
        let model = ProfileModel(profile: profile)
        #expect(model.historyStateLabel == "On — turn off any time")
        model.usesTripHistory = false
        #expect(model.historyStateLabel == "Off — recommendations use this trip only")
    }
}
