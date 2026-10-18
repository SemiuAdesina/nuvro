import SwiftUI

public struct RouteArc: Shape {
    public init() {}

    public func path(in rect: CGRect) -> Path {
        let sx = rect.width / 46
        let sy = rect.height / 30
        var path = Path()
        path.move(to: CGPoint(x: rect.minX + 2 * sx, y: rect.minY + 26 * sy))
        path.addQuadCurve(
            to: CGPoint(x: rect.minX + 44 * sx, y: rect.minY + 10 * sy),
            control: CGPoint(x: rect.minX + 20 * sx, y: rect.minY - 4 * sy)
        )
        return path
    }

    public static func pinCenter(in rect: CGRect) -> CGPoint {
        CGPoint(x: rect.minX + rect.width * 44 / 46, y: rect.minY + rect.height * 10 / 30)
    }
}

public struct RouteArcMark: View {
    private let width: CGFloat

    public init(width: CGFloat) {
        self.width = width
    }

    private var height: CGFloat { width * 30 / 46 }
    private var strokeWidth: CGFloat { width * 2 / 46 }
    private var pinRadius: CGFloat { width * 2.6 / 46 }

    public var body: some View {
        ZStack {
            RouteArc()
                .stroke(Palette.gold, style: StrokeStyle(lineWidth: strokeWidth, lineCap: .round))
            Circle()
                .fill(Palette.gold)
                .frame(width: pinRadius * 2, height: pinRadius * 2)
                .position(RouteArc.pinCenter(in: CGRect(x: 0, y: 0, width: width, height: height)))
        }
        .frame(width: width, height: height)
    }
}
