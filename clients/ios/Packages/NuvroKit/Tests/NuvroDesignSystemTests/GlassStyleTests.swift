import SwiftUI
import Testing

@testable import NuvroDesignSystem

@Suite("GlassStyle")
struct GlassStyleTests {
    private func opacity(_ color: Color) -> Float {
        color.resolve(in: EnvironmentValues()).opacity
    }

    @Test("only the tinted call to action is interactive")
    func interactivity() {
        #expect(GlassStyle.tinted(Palette.sky).isInteractive)
        #expect(!GlassStyle.control.isInteractive)
        #expect(!GlassStyle.notice.isInteractive)
        #expect(!GlassStyle.dark.isInteractive)
        #expect(!GlassStyle.sheet.isInteractive)
    }

    @Test("fallback fills mirror the design's backdrop opacities")
    func fallbackFills() {
        #expect(abs(opacity(GlassStyle.control.fallbackFill) - 0.16) < 0.005)
        #expect(abs(opacity(GlassStyle.tinted(Palette.sky).fallbackFill) - 0.82) < 0.005)
        #expect(abs(opacity(GlassStyle.dark.fallbackFill) - 0.62) < 0.005)
        #expect(abs(opacity(GlassStyle.sheet.fallbackFill) - 0.86) < 0.005)
    }

    @Test("fallback strokes are the hairline highlights from the design")
    func fallbackStrokes() {
        #expect(abs(opacity(GlassStyle.control.fallbackStroke) - 0.35) < 0.005)
        #expect(abs(opacity(GlassStyle.tinted(Palette.sky).fallbackStroke) - 0.45) < 0.005)
        #expect(abs(opacity(GlassStyle.sheet.fallbackStroke) - 0.70) < 0.005)
    }
}
