import Foundation

public struct StyleChip: Sendable, Hashable, Codable, Identifiable {
    public enum Tone: String, Sendable, Codable {
        case aqua, gold, twilight, sky
    }

    public let name: String
    public let tone: Tone

    public init(name: String, tone: Tone) {
        self.name = name
        self.tone = tone
    }

    public var id: String { name }
}

public struct PreferenceRow: Sendable, Hashable, Codable, Identifiable {
    public let name: String
    public let value: String
    public let placeholder: String

    public init(name: String, value: String, placeholder: String) {
        self.name = name
        self.value = value
        self.placeholder = placeholder
    }

    public var id: String { name }
}

public struct LearnedInsight: Sendable, Hashable, Codable, Identifiable {
    public let text: String
    public let source: String

    public init(text: String, source: String) {
        self.text = text
        self.source = source
    }

    public var id: String { text }
}

public struct TravelerProfile: Sendable, Hashable, Codable {
    public let name: String
    public let initial: String
    public let homeLabel: String
    public let styleChips: [StyleChip]
    public let preferences: [PreferenceRow]
    public let learned: [LearnedInsight]
    public let usesTripHistory: Bool

    public init(name: String, initial: String, homeLabel: String, styleChips: [StyleChip], preferences: [PreferenceRow], learned: [LearnedInsight], usesTripHistory: Bool) {
        self.name = name
        self.initial = initial
        self.homeLabel = homeLabel
        self.styleChips = styleChips
        self.preferences = preferences
        self.learned = learned
        self.usesTripHistory = usesTripHistory
    }

    private enum CodingKeys: String, CodingKey {
        case name, initial, preferences, learned
        case homeLabel = "home_label"
        case styleChips = "style_chips"
        case usesTripHistory = "uses_trip_history"
    }
}

public enum DistanceUnit: String, Sendable, Codable, CaseIterable { case miles, kilometres }
public enum TemperatureUnit: String, Sendable, Codable, CaseIterable { case fahrenheit, celsius }
public enum ClockFormat: String, Sendable, Codable, CaseIterable { case twelveHour = "12h", twentyFourHour = "24h" }
public enum DateStyle: String, Sendable, Codable, CaseIterable { case monthFirst = "month_first", dayFirst = "day_first" }

public struct UnitPreferences: Sendable, Hashable, Codable {
    public var distance: DistanceUnit
    public var temperature: TemperatureUnit
    public var clock: ClockFormat
    public var date: DateStyle

    public init(distance: DistanceUnit, temperature: TemperatureUnit, clock: ClockFormat, date: DateStyle) {
        self.distance = distance
        self.temperature = temperature
        self.clock = clock
        self.date = date
    }
}

public struct RegionSettings: Sendable, Hashable, Codable {
    public let appLanguage: String
    public let spokenLanguages: String
    public let region: String
    public let regionNote: String
    public let currency: String
    public let homeAirport: String

    public init(appLanguage: String, spokenLanguages: String, region: String, regionNote: String, currency: String, homeAirport: String) {
        self.appLanguage = appLanguage
        self.spokenLanguages = spokenLanguages
        self.region = region
        self.regionNote = regionNote
        self.currency = currency
        self.homeAirport = homeAirport
    }

    private enum CodingKeys: String, CodingKey {
        case region, currency
        case appLanguage = "app_language"
        case spokenLanguages = "spoken_languages"
        case regionNote = "region_note"
        case homeAirport = "home_airport"
    }
}

public struct SettingsRow: Sendable, Hashable, Codable, Identifiable {
    public let name: String
    public let value: String
    public let isDestructive: Bool

    public init(name: String, value: String = "", isDestructive: Bool = false) {
        self.name = name
        self.value = value
        self.isDestructive = isDestructive
    }

    public var id: String { name }

    private enum CodingKeys: String, CodingKey {
        case name, value
        case isDestructive = "is_destructive"
    }
}

public struct SettingsGroup: Sendable, Hashable, Codable, Identifiable {
    public let name: String
    public let rows: [SettingsRow]

    public init(name: String, rows: [SettingsRow]) {
        self.name = name
        self.rows = rows
    }

    public var id: String { name }
}

public struct NotificationSetting: Sendable, Hashable, Codable, Identifiable {
    public let name: String
    public let isOn: Bool

    public init(name: String, isOn: Bool) {
        self.name = name
        self.isOn = isOn
    }

    public var id: String { name }

    private enum CodingKeys: String, CodingKey {
        case name
        case isOn = "is_on"
    }
}

public struct NotificationGroup: Sendable, Hashable, Codable, Identifiable {
    public let name: String
    public let settings: [NotificationSetting]

    public init(name: String, settings: [NotificationSetting]) {
        self.name = name
        self.settings = settings
    }

    public var id: String { name }
}

public struct PermissionStatus: Sendable, Hashable, Codable, Identifiable {
    public enum Tone: String, Sendable, Codable {
        case granted, ask, none
    }

    public let name: String
    public let value: String
    public let tone: Tone

    public init(name: String, value: String, tone: Tone) {
        self.name = name
        self.value = value
        self.tone = tone
    }

    public var id: String { name }
}
