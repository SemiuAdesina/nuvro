import SwiftUI

public enum ThemePreference: String, CaseIterable, Sendable, Codable {
    case system
    case light
    case dark

    public static let storageKey = "themePreference"

    public var colorScheme: ColorScheme? {
        switch self {
        case .system: nil
        case .light: .light
        case .dark: .dark
        }
    }
}
