import SwiftUI

public struct DiagonalHatch: Shape {
    public init() {}

    public func path(in rect: CGRect) -> Path {
        var path = Path()
        let period: CGFloat = 14
        let band: CGFloat = 6
        var offset = -rect.height
        while offset < rect.width + rect.height {
            path.move(to: CGPoint(x: rect.minX + offset, y: rect.maxY))
            path.addLine(to: CGPoint(x: rect.minX + offset + rect.height, y: rect.minY))
            path.addLine(to: CGPoint(x: rect.minX + offset + rect.height + band, y: rect.minY))
            path.addLine(to: CGPoint(x: rect.minX + offset + band, y: rect.maxY))
            path.closeSubpath()
            offset += period
        }
        return path
    }
}

public struct PhotoStandIn: View {
    @Environment(\.designMetrics) private var metrics
    let palette: PhotoPalette
    let label: String
    let labelSize: CGFloat
    let labelAlignment: Alignment

    public init(palette: PhotoPalette, label: String, labelSize: CGFloat = 9, labelAlignment: Alignment = .bottomLeading) {
        self.palette = palette
        self.label = label
        self.labelSize = labelSize
        self.labelAlignment = labelAlignment
    }

    public var body: some View {
        ZStack(alignment: labelAlignment) {
            LinearGradient(
                stops: [
                    .init(color: palette.start, location: 0),
                    .init(color: palette.middle, location: palette.middleLocation),
                    .init(color: palette.end, location: 1)
                ],
                startPoint: UnitPoint(x: 0.1, y: 0),
                endPoint: UnitPoint(x: 0.9, y: 1)
            )
            DiagonalHatch().fill(Color.white.opacity(0.07))
            Text(label)
                .font(.system(size: metrics.points(labelSize), weight: .medium, design: .monospaced))
                .tracking(metrics.points(labelSize * 0.06))
                .foregroundStyle(Color.white.opacity(0.75))
                .padding(metrics.points(labelSize > 8 ? 10 : 9))
                .opacity(label.isEmpty ? 0 : 1)
        }
        .clipped()
        .accessibilityHidden(label.isEmpty)
    }
}
