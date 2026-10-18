import Foundation
#if canImport(CoreLocation)
import CoreLocation
#endif
#if canImport(UserNotifications)
import UserNotifications
#endif
#if canImport(AVFoundation)
import AVFoundation
#endif

public struct SystemPermissionAuthority: PermissionAuthority {
    public init() {}

    public func status(for kind: PermissionKind) async -> PermissionState {
        switch kind {
        case .location: locationState()
        case .notifications: await notificationState()
        case .camera: mediaState(isCamera: true)
        case .microphone: mediaState(isCamera: false)
        }
    }

    public func request(_ kind: PermissionKind) async -> PermissionState {
        switch kind {
        case .location:
            return await LocationRequest().requestWhenInUse()
        case .notifications:
            #if canImport(UserNotifications)
            let granted = (try? await UNUserNotificationCenter.current().requestAuthorization(options: [.alert, .sound, .badge])) ?? false
            return granted ? .granted : .denied
            #else
            return .denied
            #endif
        case .camera, .microphone:
            #if canImport(AVFoundation) && os(iOS)
            let granted = await AVCaptureDevice.requestAccess(for: kind == .camera ? .video : .audio)
            return granted ? .granted : .denied
            #else
            return .denied
            #endif
        }
    }

    private func locationState() -> PermissionState {
        #if canImport(CoreLocation)
        switch CLLocationManager().authorizationStatus {
        case .authorizedAlways, .authorizedWhenInUse: .granted
        case .denied: .denied
        case .restricted: .restricted
        case .notDetermined: .notDetermined
        @unknown default: .notDetermined
        }
        #else
        .notDetermined
        #endif
    }

    private func notificationState() async -> PermissionState {
        #if canImport(UserNotifications)
        switch await UNUserNotificationCenter.current().notificationSettings().authorizationStatus {
        case .authorized, .provisional, .ephemeral: .granted
        case .denied: .denied
        case .notDetermined: .notDetermined
        @unknown default: .notDetermined
        }
        #else
        .notDetermined
        #endif
    }

    private func mediaState(isCamera: Bool) -> PermissionState {
        #if canImport(AVFoundation) && os(iOS)
        switch AVCaptureDevice.authorizationStatus(for: isCamera ? .video : .audio) {
        case .authorized: .granted
        case .denied: .denied
        case .restricted: .restricted
        case .notDetermined: .notDetermined
        @unknown default: .notDetermined
        }
        #else
        .notDetermined
        #endif
    }
}

#if canImport(CoreLocation)
private final class LocationRequest: NSObject, CLLocationManagerDelegate, @unchecked Sendable {
    private let manager = CLLocationManager()
    private var continuation: CheckedContinuation<PermissionState, Never>?

    func requestWhenInUse() async -> PermissionState {
        await withCheckedContinuation { continuation in
            self.continuation = continuation
            manager.delegate = self
            manager.requestWhenInUseAuthorization()
        }
    }

    func locationManagerDidChangeAuthorization(_ manager: CLLocationManager) {
        let state: PermissionState = switch manager.authorizationStatus {
        case .authorizedAlways, .authorizedWhenInUse: .granted
        case .denied: .denied
        case .restricted: .restricted
        case .notDetermined: .notDetermined
        @unknown default: .notDetermined
        }
        guard state != .notDetermined, let continuation else { return }
        self.continuation = nil
        continuation.resume(returning: state)
    }
}
#else
private struct LocationRequest {
    func requestWhenInUse() async -> PermissionState { .denied }
}
#endif
