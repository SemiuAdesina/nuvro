import SwiftUI

public struct PhotoPalette: Equatable, Sendable {
    public let start: Color
    public let middle: Color
    public let end: Color
    public let middleLocation: CGFloat

    public init(start: Color, middle: Color, end: Color, middleLocation: CGFloat) {
        self.start = start
        self.middle = middle
        self.end = end
        self.middleLocation = middleLocation
    }

    public static let lisbon = PhotoPalette(start: Color(hex: 0xF6D9A6), middle: Color(hex: 0xE58F6A), end: Color(hex: 0x2F6FA8), middleLocation: 0.48)
    public static let kyoto = PhotoPalette(start: Color(hex: 0xF8DEE0), middle: Color(hex: 0xC97B8E), end: Color(hex: 0x3B4E7A), middleLocation: 0.50)
    public static let rainforest = PhotoPalette(start: Color(hex: 0xE4F4DC), middle: Color(hex: 0x5FC9C4), end: Color(hex: 0x1F6B5E), middleLocation: 0.50)
    public static let aqua = PhotoPalette(start: Color(hex: 0x9BEAE4), middle: Color(hex: 0x5FC9C4), end: Color(hex: 0x2E6FA0), middleLocation: 0.45)
    public static let twilight = PhotoPalette(start: Color(hex: 0x8A9AC6), middle: Color(hex: 0x3B4E7A), end: Color(hex: 0x14213D), middleLocation: 0.55)
    public static let gold = PhotoPalette(start: Color(hex: 0xF6E2B8), middle: Color(hex: 0xE8C27A), end: Color(hex: 0xC99A5B), middleLocation: 0.50)
    public static let sky = PhotoPalette(start: Color(hex: 0x9BD4FF), middle: Color(hex: 0x2B8FE8), end: Color(hex: 0x233F73), middleLocation: 0.60)
    public static let adventure = PhotoPalette(start: Color(hex: 0xE4A986), middle: Color(hex: 0xC97B8E), end: Color(hex: 0x3B4E7A), middleLocation: 0.50)
    public static let family = PhotoPalette(start: Color(hex: 0xF6E2B8), middle: Color(hex: 0x9BD4FF), end: Color(hex: 0x2B8FE8), middleLocation: 0.60)

    public static let iconTiles: [PhotoPalette] = [.twilight, .gold, .aqua, .sky]

    public static func named(_ reference: String) -> PhotoPalette {
        switch reference {
        case "madeira", "lisbon", "portugal": .lisbon
        case "kyoto", "japan": .kyoto
        case "guanacaste", "costa-rica", "rainforest": .rainforest
        case "curacao", "aqua": .aqua
        case "twilight": .twilight
        case "gold": .gold
        case "sky": .sky
        default: .lisbon
        }
    }

    public var gradient: LinearGradient {
        LinearGradient(
            stops: [
                .init(color: start, location: 0),
                .init(color: middle, location: middleLocation),
                .init(color: end, location: 1)
            ],
            startPoint: UnitPoint(x: 0.1, y: 0),
            endPoint: UnitPoint(x: 0.9, y: 1)
        )
    }
}
