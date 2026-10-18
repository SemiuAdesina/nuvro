import SwiftUI
import Testing

@testable import NuvroDesignSystem

@Suite("Palette")
struct PaletteTests {
    private func components(_ color: Color) -> (red: Float, green: Float, blue: Float, opacity: Float) {
        let resolved = color.resolve(in: EnvironmentValues())
        return (resolved.red, resolved.green, resolved.blue, resolved.opacity)
    }

    private func expectHex(_ color: Color, _ hex: UInt32) {
        let value = components(color)
        #expect(abs(value.red - Float((hex >> 16) & 0xFF) / 255) < 0.005)
        #expect(abs(value.green - Float((hex >> 8) & 0xFF) / 255) < 0.005)
        #expect(abs(value.blue - Float(hex & 0xFF) / 255) < 0.005)
        #expect(value.opacity == 1)
    }

    @Test("brand colours resolve to the approved design values")
    func brandColours() {
        expectHex(Palette.sky, 0x2B8FE8)
        expectHex(Palette.aqua, 0x5FC9C4)
        expectHex(Palette.gold, 0xE8C27A)
        expectHex(Palette.twilight, 0x3B4E7A)
        expectHex(Palette.navy, 0x14213D)
        expectHex(Palette.cloud, 0xF6F4EF)
    }

    @Test("hex initialiser carries opacity through")
    func hexWithOpacity() {
        let value = components(Color(hex: 0x14213D, opacity: 0.5))

        #expect(abs(value.opacity - 0.5) < 0.005)
        #expect(abs(value.red - Float(0x14) / 255) < 0.005)
    }

    @Test("alpha helpers keep the base colour")
    func alphaHelpers() {
        let navy = components(Palette.navyAlpha(0.6))
        let cloud = components(Palette.cloudAlpha(0.85))

        #expect(abs(navy.opacity - 0.6) < 0.005)
        #expect(abs(navy.blue - Float(0x3D) / 255) < 0.005)
        #expect(abs(cloud.opacity - 0.85) < 0.005)
    }
}
