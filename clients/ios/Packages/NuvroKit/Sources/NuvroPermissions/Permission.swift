import Foundation

public enum PermissionKind: String, CaseIterable, Sendable, Codable {
    case location
    case notifications
    case camera
    case microphone
}

public enum PermissionState: String, Sendable, Codable {
    case notDetermined = "not_determined"
    case granted
    case denied
    case restricted

    public var isUsable: Bool { self == .granted }
    public var needsSettings: Bool { self == .denied || self == .restricted }
}

public protocol PermissionAuthority: Sendable {
    func status(for kind: PermissionKind) async -> PermissionState
    func request(_ kind: PermissionKind) async -> PermissionState
}

public struct ScriptedPermissionAuthority: PermissionAuthority {
    private let states: [PermissionKind: PermissionState]
    private let requestResults: [PermissionKind: PermissionState]

    public init(states: [PermissionKind: PermissionState], requestResults: [PermissionKind: PermissionState] = [:]) {
        self.states = states
        self.requestResults = requestResults
    }

    public func status(for kind: PermissionKind) async -> PermissionState {
        states[kind] ?? .notDetermined
    }

    public func request(_ kind: PermissionKind) async -> PermissionState {
        requestResults[kind] ?? states[kind] ?? .denied
    }
}
