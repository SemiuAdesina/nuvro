import SwiftUI

public enum SkyGradient {
    public static let dawn = LinearGradient(
        stops: [
            .init(color: Color(hex: 0x2B3B69), location: 0.00),
            .init(color: Color(hex: 0x4A5D93), location: 0.22),
            .init(color: Color(hex: 0x9B8FB3), location: 0.40),
            .init(color: Color(hex: 0xE4A986), location: 0.56),
            .init(color: Color(hex: 0xF3D39A), location: 0.68),
            .init(color: Color(hex: 0xF6E2B8), location: 0.76),
            .init(color: Color(hex: 0x1D2A4D), location: 1.00)
        ],
        startPoint: .top,
        endPoint: .bottom
    )

    public static let heroDawn = LinearGradient(
        stops: [
            .init(color: Color(hex: 0x2B3B69), location: 0.00),
            .init(color: Color(hex: 0x4A5D93), location: 0.30),
            .init(color: Color(hex: 0x9B8FB3), location: 0.52),
            .init(color: Color(hex: 0xE4A986), location: 0.70),
            .init(color: Color(hex: 0xF3D39A), location: 1.00)
        ],
        startPoint: .top,
        endPoint: .bottom
    )

    public static let photoFooterScrim = LinearGradient(
        stops: [
            .init(color: Color(hex: 0x14213D, opacity: 0), location: 0.45),
            .init(color: Color(hex: 0x14213D, opacity: 0.55), location: 1)
        ],
        startPoint: .top,
        endPoint: .bottom
    )

    public static let groundScrim = LinearGradient(
        stops: [
            .init(color: Color(hex: 0x1D2A4D, opacity: 0.00), location: 0.00),
            .init(color: Color(hex: 0x1D2A4D, opacity: 0.72), location: 0.34),
            .init(color: Color(hex: 0x14213D, opacity: 0.96), location: 0.70)
        ],
        startPoint: .top,
        endPoint: .bottom
    )

    public static let globe = RadialGradient(
        stops: [
            .init(color: Color(hex: 0x7FD3CD), location: 0.00),
            .init(color: Color(hex: 0x3F9FB4), location: 0.18),
            .init(color: Color(hex: 0x2E6FA0), location: 0.34),
            .init(color: Color(hex: 0x233F73), location: 0.55),
            .init(color: Color(hex: 0x14213D), location: 0.80)
        ],
        center: UnitPoint(x: 0.5, y: 0.12),
        startRadius: 0,
        endRadius: 470
    )

    public static let sunGlow = RadialGradient(
        stops: [
            .init(color: Color(hex: 0xFFECBE, opacity: 0.95), location: 0.00),
            .init(color: Color(hex: 0xF3D39A, opacity: 0.55), location: 0.22),
            .init(color: Color(hex: 0xF3D39A, opacity: 0.00), location: 0.62)
        ],
        center: .center,
        startRadius: 0,
        endRadius: 160
    )
}
