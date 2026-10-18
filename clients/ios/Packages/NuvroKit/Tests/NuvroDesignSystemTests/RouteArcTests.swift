import CoreGraphics
import Testing

@testable import NuvroDesignSystem

@Suite("RouteArc")
struct RouteArcTests {
    @Test("the arc starts at the foot of the V and ends at the pin")
    func endpoints() {
        let rect = CGRect(x: 0, y: 0, width: 46, height: 30)
        let path = RouteArc().path(in: rect)
        let bounds = path.boundingRect

        #expect(abs(bounds.minX - 2) < 0.01)
        #expect(abs(bounds.maxX - 44) < 0.01)
        #expect(abs(bounds.maxY - 26) < 0.01)
        #expect(bounds.minY < 10)
    }

    @Test("the pin sits where the arc lands")
    func pinCentre() {
        let pin = RouteArc.pinCenter(in: CGRect(x: 0, y: 0, width: 46, height: 30))

        #expect(abs(pin.x - 44) < 0.01)
        #expect(abs(pin.y - 10) < 0.01)
    }

    @Test("the arc scales proportionally with its frame")
    func scales() {
        let path = RouteArc().path(in: CGRect(x: 0, y: 0, width: 92, height: 60))
        let bounds = path.boundingRect

        #expect(abs(bounds.minX - 4) < 0.01)
        #expect(abs(bounds.maxX - 88) < 0.01)
        #expect(abs(bounds.maxY - 52) < 0.01)
    }

    @Test("the arc crests above its landing pin, as a route leaving the V should")
    func crestsAbovePin() {
        let path = RouteArc().path(in: CGRect(x: 0, y: 0, width: 46, height: 30))
        let apex = path.boundingRect.minY

        #expect(abs(apex - 5.545) < 0.01)
        #expect(apex < 10)
    }
}
