import SwiftUI
#if canImport(UIKit)
import UIKit
#endif

public enum Palette {
    public static let navy = Color(hex: 0x14213D)
    public static let deepNavy = Color(hex: 0x1D2A4D)
    public static let twilight = Color(hex: 0x3B4E7A)
    public static let twilightText = Color.adaptive(light: 0x3B4E7A, dark: 0x8A9AC6)
    public static let twilightDeep = Color(hex: 0x2B3B69)
    public static let cloud = Color(hex: 0xF6F4EF)
    public static let lavender = Color(hex: 0x9B8FB3)
    public static let coral = Color(hex: 0xE58F6A)
    public static let goldDeep = Color(hex: 0xC99A5B)
    public static let sunGlow = Color(hex: 0xF3D39A)
    public static let dawnCream = Color(hex: 0xF6E2B8)

    public static let sky = Color.adaptive(light: 0x2B8FE8, dark: 0x4DA3F0)
    public static let aqua = Color.adaptive(light: 0x5FC9C4, dark: 0x74D6D1)
    public static let gold = Color.adaptive(light: 0xE8C27A, dark: 0xEFCB86)
    public static let matchTeal = Color.adaptive(light: 0x1F7F7A, dark: 0x7FD9D3)
    public static let sunlight = Color.adaptive(light: 0xE0A83A, dark: 0xF0BE5A)
    public static let rainBlue = Color.adaptive(light: 0x9BD4FF, dark: 0x7FC0F5)
    public static let amber = Color.adaptive(light: 0x8A6420, dark: 0xE8C27A)
    public static let destructive = Color.adaptive(light: 0xB4553F, dark: 0xE8765C)
    public static let mapLand = Color.adaptive(light: 0xEEF0EA, dark: 0x1E2740)
    public static let dawnWashTop = Color.adaptive(light: 0xF3D39A, dark: 0x2B3B69)
    public static let dawnWashMid = Color.adaptive(light: 0xF6E2B8, dark: 0x1D2A4D)

    public static let background = Color.adaptive(light: 0xF6F4EF, dark: 0x0F1626)
    public static let surface = Color.adaptive(light: 0xFFFDF8, dark: 0x1A2238)
    public static let surfaceElevated = Color.adaptive(light: 0xFFFFFF, dark: 0x232C45)
    public static let segmentTrack = Color.adaptive(light: 0xE9E6DF, dark: 0x1C2540)
    public static let segmentSelection = Color.adaptive(light: 0xFFFDF8, dark: 0x3A4666)
    public static let label = Color.adaptive(light: 0x14213D, dark: 0xF3F1EC)
    public static let labelLight = Color(hex: 0x14213D)
    public static let labelDark = Color(hex: 0xF3F1EC)

    public static func labelAlpha(_ opacity: Double) -> Color {
        Color.adaptive(light: 0x14213D, dark: 0xF3F1EC, opacity: opacity)
    }

    public static func cloudAlpha(_ opacity: Double) -> Color {
        Color(hex: 0xF6F4EF, opacity: opacity)
    }

    public static func navyAlpha(_ opacity: Double) -> Color {
        Color(hex: 0x14213D, opacity: opacity)
    }

    public static var separator: Color { labelAlpha(0.08) }
    public static var fill: Color { labelAlpha(0.06) }
}

public extension Color {
    static func adaptive(light: UInt32, dark: UInt32, opacity: Double = 1) -> Color {
        #if canImport(UIKit)
        Color(uiColor: UIColor { traits in
            UIColor(Color(hex: traits.userInterfaceStyle == .dark ? dark : light, opacity: opacity))
        })
        #else
        Color(hex: light, opacity: opacity)
        #endif
    }
}

public extension Color {
    init(hex: UInt32) {
        self.init(
            .sRGB,
            red: Double((hex >> 16) & 0xFF) / 255,
            green: Double((hex >> 8) & 0xFF) / 255,
            blue: Double(hex & 0xFF) / 255,
            opacity: 1
        )
    }

    init(hex: UInt32, opacity: Double) {
        self.init(
            .sRGB,
            red: Double((hex >> 16) & 0xFF) / 255,
            green: Double((hex >> 8) & 0xFF) / 255,
            blue: Double(hex & 0xFF) / 255,
            opacity: opacity
        )
    }
}
