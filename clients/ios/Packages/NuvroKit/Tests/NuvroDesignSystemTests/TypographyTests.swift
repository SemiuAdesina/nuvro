import Testing

@testable import NuvroDesignSystem

@Suite("Typography")
struct TypographyTests {
    @Test("the hierarchy matches the design: 31 headline, 15 body, 12 labels")
    func hierarchy() {
        #expect(Typography.headline.size == 31)
        #expect(Typography.headline.weight == .bold)
        #expect(Typography.body.size == 15)
        #expect(Typography.caption.size == 12)
        #expect(Typography.callToAction.size == 17)
        #expect(Typography.control.size == 14)
    }

    @Test("tracking is negative on display sizes and positive on the tagline")
    func tracking() {
        #expect(Typography.headline.tracking < 0)
        #expect(Typography.supporting.tracking < 0)
        #expect(Typography.callToAction.tracking < 0)
        #expect(Typography.tagline.tracking > 0)
        #expect(Typography.caption.tracking == 0)
    }

    @Test("wordmark tracking is sixteen percent of the point size")
    func wordmarkTracking() {
        #expect(BrandTypeface.wordmarkTrackingRatio == 0.16)
        #expect(abs(46 * BrandTypeface.wordmarkTrackingRatio - 7.36) < 0.001)
    }
}
