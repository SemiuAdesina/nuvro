import SwiftUI

public struct GlassPillButtonStyle: ButtonStyle {
    let style: GlassStyle

    public init(style: GlassStyle = .light) {
        self.style = style
    }

    public func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .glassCapsule(style)
            .scaleEffect(configuration.isPressed ? 0.96 : 1)
            .animation(.easeOut(duration: 0.15), value: configuration.isPressed)
    }
}

public struct GlassCallToActionStyle: ButtonStyle {
    public init() {}

    public func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .background(Palette.sky.opacity(configuration.isPressed ? 0.85 : 1), in: Capsule())
            .scaleEffect(configuration.isPressed ? 0.98 : 1)
            .animation(.easeOut(duration: 0.15), value: configuration.isPressed)
    }
}

public struct PressScaleButtonStyle: ButtonStyle {
    let scale: CGFloat

    public init(scale: CGFloat = 0.985) {
        self.scale = scale
    }

    public func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .scaleEffect(configuration.isPressed ? scale : 1)
            .animation(.easeOut(duration: 0.15), value: configuration.isPressed)
    }
}

public struct CallToActionButton: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let isLoading: Bool
    let action: () -> Void

    public init(_ title: String, isLoading: Bool = false, action: @escaping () -> Void) {
        self.title = title
        self.isLoading = isLoading
        self.action = action
    }

    public var body: some View {
        Button(action: action) {
            ZStack {
                Text(title)
                    .typeStyle(Typography.callToAction)
                    .foregroundStyle(.white)
                    .opacity(isLoading ? 0 : 1)
                if isLoading {
                    ProgressView().tint(.white)
                }
            }
            .frame(maxWidth: .infinity)
            .frame(height: metrics.points(50))
        }
        .buttonStyle(GlassCallToActionStyle())
        .disabled(isLoading)
        .accessibilityValue(isLoading ? "In progress" : "")
    }
}

public struct LinkButton: View {
    let title: String
    let action: () -> Void

    public init(_ title: String, action: @escaping () -> Void) {
        self.title = title
        self.action = action
    }

    public var body: some View {
        Button(action: action) {
            Text(title).typeStyle(Typography.link).foregroundStyle(Palette.sky)
        }
        .buttonStyle(.plain)
    }
}

public struct GlassIconButton: View {
    @Environment(\.designMetrics) private var metrics
    let systemImage: String
    let style: GlassStyle
    let accessibilityLabel: String
    let action: () -> Void

    public init(systemImage: String, style: GlassStyle = .light, accessibilityLabel: String, action: @escaping () -> Void) {
        self.systemImage = systemImage
        self.style = style
        self.accessibilityLabel = accessibilityLabel
        self.action = action
    }

    public var body: some View {
        Button(action: action) {
            Image(systemName: systemImage)
                .font(.system(size: metrics.points(15), weight: .semibold))
                .foregroundStyle(style == .light ? Palette.label : Palette.cloud)
                .frame(width: metrics.points(38), height: metrics.points(38))
        }
        .buttonStyle(GlassPillButtonStyle(style: style))
        .accessibilityLabel(accessibilityLabel)
    }
}

public struct SaveButton: View {
    let isSaved: Bool
    let style: GlassStyle
    let action: () -> Void

    public init(isSaved: Bool, style: GlassStyle = .control, action: @escaping () -> Void) {
        self.isSaved = isSaved
        self.style = style
        self.action = action
    }

    public var body: some View {
        GlassIconButton(
            systemImage: isSaved ? "bookmark.fill" : "bookmark",
            style: style,
            accessibilityLabel: isSaved ? "Saved" : "Save",
            action: action
        )
        .contentTransition(.symbolEffect(.replace))
    }
}

public struct InlineActions<Content: View>: View {
    @Environment(\.designMetrics) private var metrics
    private let content: Content

    public init(@ViewBuilder content: () -> Content) {
        self.content = content()
    }

    public var body: some View {
        VStack(spacing: metrics.points(10)) {
            content
        }
        .padding(.top, metrics.points(8))
        .frame(maxWidth: .infinity)
    }
}
