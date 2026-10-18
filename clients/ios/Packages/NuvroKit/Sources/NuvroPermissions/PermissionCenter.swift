import Foundation
import Observation

@Observable
@MainActor
public final class PermissionCenter {
    public enum Presentation: Equatable, Sendable {
        case explainer(PermissionKind)
        case deniedGuidance(PermissionKind)
    }

    public private(set) var states: [PermissionKind: PermissionState] = [:]
    public private(set) var presentation: Presentation?
    private let authority: any PermissionAuthority

    public init(authority: any PermissionAuthority) {
        self.authority = authority
    }

    public func state(of kind: PermissionKind) -> PermissionState {
        states[kind] ?? .notDetermined
    }

    public func refresh(_ kind: PermissionKind) async {
        states[kind] = await authority.status(for: kind)
    }

    public func refreshAll() async {
        for kind in PermissionKind.allCases { await refresh(kind) }
    }

    public func prepare(_ kind: PermissionKind) async -> Bool {
        await refresh(kind)
        switch state(of: kind) {
        case .granted:
            presentation = nil
            return true
        case .notDetermined:
            presentation = .explainer(kind)
            return false
        case .denied, .restricted:
            presentation = .deniedGuidance(kind)
            return false
        }
    }

    public func proceedFromExplainer() async -> PermissionState {
        guard case .explainer(let kind) = presentation else { return .notDetermined }
        presentation = nil
        let result = await authority.request(kind)
        states[kind] = result
        if result.needsSettings { presentation = .deniedGuidance(kind) }
        return result
    }

    public func dismissPresentation() {
        presentation = nil
    }
}
