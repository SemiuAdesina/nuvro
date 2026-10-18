import SwiftUI
import Testing

@testable import NuvroDesignSystem

@Suite("ThemePreference")
struct ThemePreferenceTests {
    @Test("system defers to the device and the others force a scheme")
    func schemes() {
        #expect(ThemePreference.system.colorScheme == nil)
        #expect(ThemePreference.light.colorScheme == .light)
        #expect(ThemePreference.dark.colorScheme == .dark)
        #expect(ThemePreference(rawValue: "dark") == .dark)
        #expect(ThemePreference.allCases.count == 3)
    }
}
