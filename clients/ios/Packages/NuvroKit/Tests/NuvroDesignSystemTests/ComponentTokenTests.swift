import SwiftUI
import Testing

@testable import NuvroDesignSystem

@Suite("Product tokens")
struct ComponentTokenTests {
    private func opacity(_ color: Color) -> Float { color.resolve(in: EnvironmentValues()).opacity }

    @Test("the in-product light glass mirrors the design's white 62 percent surface")
    func lightGlass() {
        #expect(abs(opacity(GlassStyle.light.fallbackFill) - 0.62) < 0.005)
        #expect(abs(opacity(GlassStyle.light.fallbackStroke) - 0.08) < 0.005)
        #expect(!GlassStyle.light.isInteractive)
    }

    @Test("photo references resolve to a palette and unknown ones fall back")
    func photoReferences() {
        #expect(PhotoPalette.named("madeira") == .lisbon)
        #expect(PhotoPalette.named("kyoto") == .kyoto)
        #expect(PhotoPalette.named("guanacaste") == .rainforest)
        #expect(PhotoPalette.named("curacao") == .aqua)
        #expect(PhotoPalette.named("nowhere") == .lisbon)
    }

    @Test("icon tiles cycle through the four brand palettes")
    func iconTiles() {
        #expect(PhotoPalette.iconTiles == [.twilight, .gold, .aqua, .sky])
    }

    @Test("product typography extends the onboarding hierarchy without new sizes above the large title")
    func productTypography() {
        #expect(Typography.largeTitle.size == 34)
        #expect(Typography.title.size == 20)
        #expect(Typography.budgetFigure.size == 40)
        #expect(Typography.matchLabel.tracking > 1)
        #expect(Typography.link.weight == .semibold)
    }

    @Test("new radii follow the design: hero 24, tile 18, map 22")
    func radii() {
        #expect(Radius.hero == 24)
        #expect(Radius.tile == 18)
        #expect(Radius.mapPreview == 22)
    }
}
