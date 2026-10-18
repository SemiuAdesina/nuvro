import Foundation
import Observation

@Observable
@MainActor
public final class EmailVerificationModel {
    public enum Phase: Equatable, Sendable {
        case entering
        case verifying
        case failed(AuthenticationFailure)
        case resending
        case verified(AuthSession)
    }

    public private(set) var request: EmailCodeRequest
    public private(set) var code = ""
    public private(set) var phase: Phase = .entering
    public private(set) var resendAvailableAt: Date
    private let client: any AuthenticationClient
    private let now: @Sendable () -> Date

    public init(request: EmailCodeRequest, client: any AuthenticationClient, now: @escaping @Sendable () -> Date = { Date() }) {
        self.request = request
        self.client = client
        self.now = now
        self.resendAvailableAt = now().addingTimeInterval(TimeInterval(request.resendAfterSeconds))
    }

    public var codeLength: Int { request.codeLength }
    public var maskedEmail: String { request.maskedEmail }
    public var isComplete: Bool { code.count == codeLength }
    public var isBusy: Bool { phase == .verifying || phase == .resending }

    public var failure: AuthenticationFailure? {
        if case .failed(let failure) = phase { return failure }
        return nil
    }

    public var session: AuthSession? {
        if case .verified(let session) = phase { return session }
        return nil
    }

    public func enter(_ text: String) {
        guard !isBusy else { return }
        code = String(text.filter(\.isNumber).prefix(codeLength))
        if case .failed = phase { phase = .entering }
    }

    public func canResend(at date: Date) -> Bool {
        date >= resendAvailableAt && !isBusy
    }

    public func resendCountdown(at date: Date) -> Int {
        max(0, Int(resendAvailableAt.timeIntervalSince(date).rounded(.up)))
    }

    public func verify() async {
        guard isComplete, !isBusy else { return }
        phase = .verifying
        do {
            let session = try await client.verifyEmailCode(email: request.email, code: code)
            phase = .verified(session)
        } catch {
            phase = .failed(error)
            if error == .expiredCode || error == .invalidCode { code = "" }
        }
    }

    public func resend() async {
        guard canResend(at: now()) else { return }
        phase = .resending
        do {
            request = try await client.requestEmailCode(email: request.email)
            resendAvailableAt = now().addingTimeInterval(TimeInterval(request.resendAfterSeconds))
            code = ""
            phase = .entering
        } catch {
            if case .rateLimited(let seconds?) = error {
                resendAvailableAt = now().addingTimeInterval(TimeInterval(seconds))
            }
            phase = .failed(error)
        }
    }
}
