import NuvroDesignSystem
import NuvroLocalization
import SwiftUI

public enum LegalDocument: Sendable {
    case terms
    case privacy
}

struct AuthenticationBackdrop: View {
    var body: some View {
        ZStack {
            SkyGradient.dawn
            SkyGradient.groundScrim
        }
        .ignoresSafeArea()
    }
}

struct LegalLine: View {
    @Environment(\.designMetrics) private var metrics
    let copy: AuthenticationCopy
    let onOpen: (LegalDocument) -> Void

    var body: some View {
        VStack(spacing: metrics.points(2)) {
            Text(copy.legalPrefix).typeStyle(Typography.small).foregroundStyle(Palette.cloudAlpha(0.6))
            HStack(spacing: metrics.points(6)) {
                Button(copy.terms) { onOpen(.terms) }
                Text("·").foregroundStyle(Palette.cloudAlpha(0.6))
                Button(copy.privacy) { onOpen(.privacy) }
            }
            .typeStyle(Typography.smallEmphasis)
            .foregroundStyle(Palette.cloudAlpha(0.85))
            .buttonStyle(.plain)
        }
        .multilineTextAlignment(.center)
    }
}

public struct AuthenticationFailureCard: View {
    @Environment(\.designMetrics) private var metrics
    let failure: AuthenticationFailure
    let copy: AuthenticationCopy
    let retryLabel: String
    var isAppleFailure = false
    let onRetry: () -> Void

    public init(failure: AuthenticationFailure, copy: AuthenticationCopy, retryLabel: String, isAppleFailure: Bool = false, onRetry: @escaping () -> Void) {
        self.failure = failure
        self.copy = copy
        self.retryLabel = retryLabel
        self.isAppleFailure = isAppleFailure
        self.onRetry = onRetry
    }

    public var body: some View {
        HStack(alignment: .top, spacing: metrics.points(10)) {
            Image(systemName: failure == .network ? "wifi.slash" : "exclamationmark.circle.fill")
                .font(.system(size: metrics.points(15), weight: .semibold))
                .foregroundStyle(Palette.gold)
                .accessibilityHidden(true)
            VStack(alignment: .leading, spacing: metrics.points(6)) {
                Text(message).typeStyle(Typography.bodyCompact).foregroundStyle(Palette.cloud)
                if failure.isRetryable {
                    Button(retryLabel, action: onRetry)
                        .typeStyle(Typography.link)
                        .foregroundStyle(Palette.gold)
                        .buttonStyle(.plain)
                }
            }
        }
        .padding(metrics.points(14))
        .frame(maxWidth: .infinity, alignment: .leading)
        .glass(.notice, in: RoundedRectangle(cornerRadius: metrics.points(Radius.notice), style: .continuous))
        .accessibilityElement(children: .combine)
    }

    private var message: String {
        switch failure {
        case .invalidEmail: copy.invalidEmail
        case .invalidCode: copy.failureInvalidCode
        case .expiredCode: copy.failureExpiredCode
        case .rateLimited: copy.failureRateLimited
        case .network: copy.failureNetwork
        case .unavailable: copy.failureUnavailable
        case .unauthenticated: copy.sessionExpired
        case .unexpected: isAppleFailure ? copy.failureApple : copy.failureUnexpected
        }
    }
}

public struct AuthenticationBootstrapView: View {
    @Environment(\.designMetrics) private var metrics
    private let copy: AuthenticationCopy

    public init(language: Language = .english) {
        self.copy = .resolve(language)
    }

    public var body: some View {
        VStack(spacing: metrics.points(18)) {
            Wordmark(size: metrics.points(46)).accessibilityLabel("NUVRO")
            ProgressView()
                .tint(Palette.cloud)
                .accessibilityLabel(copy.restoring)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
        .background(AuthenticationBackdrop())
    }
}

struct AuthenticationField<Content: View>: View {
    @Environment(\.designMetrics) private var metrics
    let isInvalid: Bool
    let content: Content

    init(isInvalid: Bool = false, @ViewBuilder content: () -> Content) {
        self.isInvalid = isInvalid
        self.content = content()
    }

    var body: some View {
        content
            .padding(.horizontal, metrics.points(16))
            .frame(height: metrics.points(54))
            .glass(.control, in: RoundedRectangle(cornerRadius: metrics.points(Radius.notice), style: .continuous))
            .overlay {
                if isInvalid {
                    RoundedRectangle(cornerRadius: metrics.points(Radius.notice), style: .continuous).strokeBorder(Palette.gold, lineWidth: 1)
                }
            }
    }
}

extension View {
    func hiddenNavigationBar() -> some View {
        #if os(iOS)
        toolbar(.hidden, for: .navigationBar)
        #else
        self
        #endif
    }
}
