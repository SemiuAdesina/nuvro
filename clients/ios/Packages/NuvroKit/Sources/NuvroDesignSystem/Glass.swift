import SwiftUI

public enum GlassStyle: Equatable, Sendable {
    case control
    case light
    case tinted(Color)
    case notice
    case dark
    case sheet

    var fallbackFill: Color {
        switch self {
        case .control: Color.white.opacity(0.16)
        case .light: Palette.surface.opacity(0.62)
        case .tinted(let tint): tint.opacity(0.82)
        case .notice: Color.white.opacity(0.16)
        case .dark: Palette.navy.opacity(0.62)
        case .sheet: Palette.background.opacity(0.86)
        }
    }

    var fallbackStroke: Color {
        switch self {
        case .control: Color.white.opacity(0.35)
        case .light: Palette.labelAlpha(0.08)
        case .tinted: Color.white.opacity(0.45)
        case .notice: Color.white.opacity(0.35)
        case .dark: Color.white.opacity(0.30)
        case .sheet: Palette.surfaceElevated.opacity(0.70)
        }
    }

    var isInteractive: Bool {
        if case .tinted = self { return true }
        return false
    }
}

public struct GlassSurface<S: InsettableShape>: ViewModifier {
    let style: GlassStyle
    let shape: S

    public func body(content: Content) -> some View {
        if style == .light {
            content
                .background(Palette.surface, in: shape)
                .overlay(shape.strokeBorder(Palette.labelAlpha(0.08), lineWidth: 0.5))
        } else if #available(iOS 26.0, macOS 26.0, *) {
            content.glassEffect(nativeGlass, in: shape)
        } else {
            content
                .background(.ultraThinMaterial, in: shape)
                .background(style.fallbackFill, in: shape)
                .overlay(shape.strokeBorder(style.fallbackStroke, lineWidth: 0.5))
        }
    }

    @available(iOS 26.0, macOS 26.0, *)
    private var nativeGlass: Glass {
        switch style {
        case .control, .notice, .light:
            return .regular
        case .sheet:
            return .regular.tint(Palette.background.opacity(0.86))
        case .dark:
            return .regular.tint(Palette.navy.opacity(0.62))
        case .tinted(let tint):
            return .regular.tint(tint).interactive()
        }
    }
}

public extension View {
    func glass<S: InsettableShape>(_ style: GlassStyle, in shape: S) -> some View {
        modifier(GlassSurface(style: style, shape: shape))
    }

    func glassCapsule(_ style: GlassStyle = .control) -> some View {
        modifier(GlassSurface(style: style, shape: Capsule()))
    }
}

public struct GlassGroup<Content: View>: View {
    private let spacing: CGFloat
    private let content: Content

    public init(spacing: CGFloat = 12, @ViewBuilder content: () -> Content) {
        self.spacing = spacing
        self.content = content()
    }

    public var body: some View {
        if #available(iOS 26.0, macOS 26.0, *) {
            GlassEffectContainer(spacing: spacing) { content }
        } else {
            content
        }
    }
}
