import Testing

@testable import NuvroOnboarding

@Suite("CameraPose")
struct CameraPoseTests {
    @Test("one camera pushes in across the four stages")
    func poses() {
        #expect(CameraPose.byStage[.splash] == CameraPose(zoom: 1, rise: 0))
        #expect(CameraPose.byStage[.discover] == CameraPose(zoom: 1.22, rise: 70))
        #expect(CameraPose.byStage[.plan] == CameraPose(zoom: 1.45, rise: 110))
        #expect(CameraPose.byStage[.travel] == CameraPose(zoom: 2.1, rise: 210))
    }

    @Test("layers scale and rise in proportion to their depth")
    func depthMath() {
        let pose = CameraPose(zoom: 2.1, rise: 210)

        #expect(abs(pose.layerScale(depth: LayerDepth.far) - 1.385) < 0.0001)
        #expect(abs(pose.layerScale(depth: LayerDepth.mid) - 2.1) < 0.0001)
        #expect(abs(pose.layerScale(depth: LayerDepth.near) - 2.375) < 0.0001)
        #expect(abs(pose.layerOffset(depth: LayerDepth.far) - 73.5) < 0.0001)
        #expect(pose.layerOffset(depth: LayerDepth.mid) == 210)
        #expect(pose.layerOffset(depth: LayerDepth.near) == 262.5)
    }

    @Test("the identity pose leaves every layer untouched")
    func identity() {
        let pose = CameraPose(zoom: 1, rise: 0)

        #expect(pose.layerScale(depth: LayerDepth.near) == 1)
        #expect(pose.layerOffset(depth: LayerDepth.near) == 0)
    }

    @Test("reduce motion freezes the camera after the first screen")
    func reduceMotion() {
        #expect(CameraPose.pose(for: .splash, reduceMotion: true) == CameraPose.byStage[.splash])
        #expect(CameraPose.pose(for: .plan, reduceMotion: true) == CameraPose.byStage[.discover])
        #expect(CameraPose.pose(for: .travel, reduceMotion: true) == CameraPose.byStage[.discover])
        #expect(CameraPose.pose(for: .travel, reduceMotion: false) == CameraPose.byStage[.travel])
    }

    @Test("the camera anchors at the design origin of fifty by forty two percent")
    func anchor() {
        #expect(CameraPose.anchor.x == 0.5)
        #expect(CameraPose.anchor.y == 0.42)
    }
}
