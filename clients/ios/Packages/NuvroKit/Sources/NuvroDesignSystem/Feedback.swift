import Observation
import SwiftUI

public struct TransientMessage: Equatable, Sendable, Identifiable {
    public let id: UUID
    public let text: String
    public let symbol: String

    public init(text: String, symbol: String = "checkmark") {
        self.id = UUID()
        self.text = text
        self.symbol = symbol
    }
}

@Observable
@MainActor
public final class TransientFeedback {
    public private(set) var current: TransientMessage?
    private var dismissal: Task<Void, Never>?
    private let holdDuration: Duration

    public init(holdDuration: Duration = .seconds(2.2)) {
        self.holdDuration = holdDuration
    }

    public func show(_ text: String, symbol: String = "checkmark") {
        dismissal?.cancel()
        current = TransientMessage(text: text, symbol: symbol)
        dismissal = Task { [holdDuration] in
            try? await Task.sleep(for: holdDuration)
            guard !Task.isCancelled else { return }
            current = nil
        }
    }

    public func dismiss() {
        dismissal?.cancel()
        current = nil
    }
}

public struct TransientBanner: View {
    @Environment(\.designMetrics) private var metrics
    let message: TransientMessage

    public init(message: TransientMessage) {
        self.message = message
    }

    public var body: some View {
        HStack(spacing: metrics.points(9)) {
            Image(systemName: message.symbol)
                .font(.system(size: metrics.points(11), weight: .bold))
                .foregroundStyle(.white)
                .frame(width: metrics.points(22), height: metrics.points(22))
                .background(Palette.aqua, in: Circle())
                .accessibilityHidden(true)
            Text(message.text).typeStyle(Typography.bodyCompactEmphasis).foregroundStyle(Palette.label)
        }
        .padding(.leading, metrics.points(10)).padding(.trailing, metrics.points(16)).padding(.vertical, metrics.points(9))
        .glassCapsule(.light)
        .accessibilityElement(children: .combine)
        .accessibilityAddTraits(.updatesFrequently)
    }
}

struct TransientFeedbackOverlay: ViewModifier {
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @Environment(\.designMetrics) private var metrics
    let feedback: TransientFeedback

    func body(content: Content) -> some View {
        content.overlay(alignment: .top) {
            if let message = feedback.current {
                TransientBanner(message: message)
                    .padding(.top, metrics.points(8))
                    .transition(reduceMotion ? .opacity : .move(edge: .top).combined(with: .opacity))
                    .id(message.id)
            }
        }
        .animation(reduceMotion ? Motion.fade(0.25) : Motion.settle(0.45), value: feedback.current)
    }
}

public extension View {
    func transientFeedback(_ feedback: TransientFeedback) -> some View {
        modifier(TransientFeedbackOverlay(feedback: feedback))
    }
}

public struct NuvroOfflineBanner: View {
    @Environment(\.designMetrics) private var metrics
    let title: String
    let detail: String?
    let isReconnecting: Bool

    public init(title: String, detail: String? = nil, isReconnecting: Bool = false) {
        self.title = title
        self.detail = detail
        self.isReconnecting = isReconnecting
    }

    public var body: some View {
        HStack(spacing: metrics.points(10)) {
            if isReconnecting {
                ProgressView().tint(Palette.cloud).controlSize(.small)
            } else {
                Image(systemName: "wifi.slash").font(.system(size: metrics.points(13), weight: .semibold)).foregroundStyle(Palette.gold).accessibilityHidden(true)
            }
            VStack(alignment: .leading, spacing: 1) {
                Text(title).typeStyle(Typography.metaEmphasis).foregroundStyle(Palette.cloud)
                if let detail { Text(detail).typeStyle(Typography.small).foregroundStyle(Palette.cloudAlpha(0.75)) }
            }
            Spacer(minLength: 0)
        }
        .padding(.horizontal, metrics.points(14)).padding(.vertical, metrics.points(9))
        .glass(.dark, in: RoundedRectangle(cornerRadius: metrics.points(Radius.notice), style: .continuous))
        .accessibilityElement(children: .combine)
    }
}
