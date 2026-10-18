import CoreGraphics
import Testing

@testable import NuvroOnboarding

@Suite("Route geometry")
struct RouteGeometryTests {
    @Test("the globe route starts and ends on the design's coordinates")
    func globeRouteEndpoints() {
        #expect(GlobeRoute.point(at: 0) == CGPoint(x: 92, y: 352))
        #expect(GlobeRoute.point(at: 1) == CGPoint(x: 330, y: 350))
    }

    @Test("the travelling light peaks near the arc's apex halfway along")
    func globeRouteApex() {
        let midpoint = GlobeRoute.point(at: 0.5)

        #expect(abs(midpoint.x - 205.5) < 0.01)
        #expect(abs(midpoint.y - 273.5) < 0.01)
        #expect(midpoint.y < GlobeRoute.point(at: 0).y)
    }

    @Test("progress is clamped so the light never leaves the route")
    func globeRouteClamps() {
        #expect(GlobeRoute.point(at: -1) == GlobeRoute.point(at: 0))
        #expect(GlobeRoute.point(at: 2) == GlobeRoute.point(at: 1))
    }

    @Test("the globe route path scales with its frame")
    func globeRouteScales() {
        let bounds = GlobeRoute().path(in: CGRect(x: 0, y: 0, width: 804, height: 1120)).boundingRect

        #expect(abs(bounds.minX - 184) < 0.01)
        #expect(abs(bounds.maxX - 660) < 0.01)
    }

    @Test("the live trip route runs from the traveller to the museum")
    func liveTripRoute() {
        let bounds = LiveTripRoute().path(in: CGRect(x: 0, y: 0, width: 342, height: 330)).boundingRect

        #expect(abs(bounds.minX - 62) < 0.5)
        #expect(abs(bounds.maxX - 212) < 0.5)
        #expect(abs(bounds.maxY - 282) < 0.5)
        #expect(abs(bounds.minY - 100) < 0.5)
    }

    @Test("the travelling light is a circle at the current progress")
    func travellingLight() {
        let bounds = TravellingLight(progress: 0).path(in: CGRect(x: 0, y: 0, width: 402, height: 560)).boundingRect

        #expect(abs(bounds.midX - 92) < 0.01)
        #expect(abs(bounds.midY - 352) < 0.01)
        #expect(abs(bounds.width - 6.4) < 0.01)
    }

    @Test("the planning connectors join the estimate to both bookings")
    func planningConnectors() {
        let bounds = PlanningConnectors().path(in: CGRect(x: 0, y: 0, width: 402, height: 560)).boundingRect

        #expect(abs(bounds.minX - 82) < 0.01)
        #expect(abs(bounds.maxX - 322) < 0.01)
        #expect(abs(bounds.minY - 162) < 0.01)
        #expect(abs(bounds.maxY - 412) < 0.01)
    }

    @Test("mountain peaks follow the clip path percentages")
    func mountains() {
        let path = Mountains().path(in: CGRect(x: 0, y: 0, width: 100, height: 100))
        let bounds = path.boundingRect

        #expect(bounds.minY == 0)
        #expect(bounds.maxY == 100)
        #expect(bounds.width == 100)
    }
}
