import SwiftUI

struct GlobeRoute: Shape {
    static let start = CGPoint(x: 92, y: 352)
    static let control = CGPoint(x: 200, y: 196)
    static let end = CGPoint(x: 330, y: 350)
    static let designSize = CGSize(width: 402, height: 560)

    func path(in rect: CGRect) -> Path {
        let sx = rect.width / Self.designSize.width
        let sy = rect.height / Self.designSize.height
        var path = Path()
        path.move(to: CGPoint(x: Self.start.x * sx, y: Self.start.y * sy))
        path.addQuadCurve(
            to: CGPoint(x: Self.end.x * sx, y: Self.end.y * sy),
            control: CGPoint(x: Self.control.x * sx, y: Self.control.y * sy)
        )
        return path
    }

    static func point(at t: CGFloat) -> CGPoint {
        let clamped = max(0, min(1, t))
        let inverse = 1 - clamped
        let x = inverse * inverse * start.x + 2 * inverse * clamped * control.x + clamped * clamped * end.x
        let y = inverse * inverse * start.y + 2 * inverse * clamped * control.y + clamped * clamped * end.y
        return CGPoint(x: x, y: y)
    }
}

struct TravellingLight: Shape {
    var progress: CGFloat

    var animatableData: CGFloat {
        get { progress }
        set { progress = newValue }
    }

    func path(in rect: CGRect) -> Path {
        let scale = rect.width / GlobeRoute.designSize.width
        let centre = GlobeRoute.point(at: progress)
        let radius = 3.2 * scale
        return Path(ellipseIn: CGRect(
            x: rect.minX + centre.x * scale - radius,
            y: rect.minY + centre.y * scale - radius,
            width: radius * 2,
            height: radius * 2
        ))
    }
}
