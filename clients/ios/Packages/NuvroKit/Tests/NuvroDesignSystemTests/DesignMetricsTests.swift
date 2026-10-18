import CoreGraphics
import Testing

@testable import NuvroDesignSystem

@Suite("DesignMetrics")
struct DesignMetricsTests {
    @Test("the reference device maps design points one to one")
    func referenceIsIdentity() {
        let metrics = DesignMetrics.reference

        #expect(metrics.scale == 1)
        #expect(metrics.points(31) == 31)
        #expect(metrics.fromTop(168) == 168)
        #expect(metrics.fromBottom(92) == 782.0)
        #expect(metrics.verticalSlack == 0)
    }

    @Test("a wider device scales every dimension by width")
    func scalesByWidth() {
        let metrics = DesignMetrics(size: CGSize(width: 430, height: 932))

        #expect(abs(metrics.scale - 430.0 / 402.0) < 0.0001)
        #expect(abs(metrics.points(200) - 200 * 430 / 402) < 0.001)
        #expect(metrics.horizontalCenter == 215)
    }

    @Test("bottom anchored offsets are measured from the actual bottom edge")
    func bottomAnchoring() {
        let metrics = DesignMetrics(size: CGSize(width: 402, height: 1000))

        #expect(metrics.fromBottom(92) == 908)
        #expect(metrics.verticalSlack == 126)
    }

    @Test("safe area aware positions follow the device insets, not the canvas")
    func safeAreaAnchoring() {
        let metrics = DesignMetrics(size: CGSize(width: 402, height: 874), safeTop: 50, safeBottom: 20)

        #expect(metrics.belowSafeTop(66) == 54)
        #expect(metrics.aboveSafeBottom(92) == 796.0)
    }

    @Test("safe area helpers match the canvas on the reference device")
    func safeAreaIdentityOnReference() {
        let metrics = DesignMetrics.reference

        #expect(metrics.belowSafeTop(66) == 66)
        #expect(metrics.aboveSafeBottom(92) == metrics.fromBottom(92))
    }

    @Test("zero width never divides by zero")
    func zeroWidthIsSafe() {
        #expect(DesignMetrics(size: .zero).scale == 1)
    }

    @Test("top leading frames convert to SwiftUI centre positions")
    func topLeadingCentre() {
        let centre = DesignMetrics.reference.topLeading(30, 96, 342, 330)

        #expect(centre.x == 201.0)
        #expect(centre.y == 261.0)
    }
}
