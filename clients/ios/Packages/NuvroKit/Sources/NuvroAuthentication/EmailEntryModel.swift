import Foundation
import Observation

public enum EmailFormat {
    public static func isPlausible(_ email: String) -> Bool {
        let trimmed = email.trimmingCharacters(in: .whitespacesAndNewlines)
        let parts = trimmed.split(separator: "@", omittingEmptySubsequences: false)
        guard parts.count == 2, !parts[0].isEmpty else { return false }
        let domain = parts[1]
        guard let dot = domain.lastIndex(of: "."), dot != domain.startIndex, domain.distance(from: dot, to: domain.endIndex) > 2 else { return false }
        return !trimmed.contains(" ")
    }
}

@Observable
@MainActor
public final class EmailEntryModel {
    public enum Phase: Equatable, Sendable {
        case pristine
        case editing
        case submitting
        case rejected(AuthenticationFailure)
    }

    public var email = "" {
        didSet { if phase != .submitting { phase = .editing } }
    }
    public private(set) var phase: Phase = .pristine
    public private(set) var hasFinishedEditing = false
    private let client: any AuthenticationClient

    public init(client: any AuthenticationClient) {
        self.client = client
    }

    public var normalisedEmail: String {
        email.trimmingCharacters(in: .whitespacesAndNewlines).lowercased()
    }

    public var isFormatValid: Bool {
        EmailFormat.isPlausible(normalisedEmail)
    }

    public var showsFormatError: Bool {
        hasFinishedEditing && !normalisedEmail.isEmpty && !isFormatValid
    }

    public var rejection: AuthenticationFailure? {
        if case .rejected(let failure) = phase { return failure }
        return nil
    }

    public var isSubmitting: Bool { phase == .submitting }

    public var canSubmit: Bool { isFormatValid && !isSubmitting }

    public func finishEditing() {
        hasFinishedEditing = true
    }

    public func submit() async -> EmailCodeRequest? {
        hasFinishedEditing = true
        guard canSubmit else { return nil }
        phase = .submitting
        do {
            let request = try await client.requestEmailCode(email: normalisedEmail)
            phase = .editing
            return request
        } catch {
            phase = .rejected(error)
            return nil
        }
    }
}
