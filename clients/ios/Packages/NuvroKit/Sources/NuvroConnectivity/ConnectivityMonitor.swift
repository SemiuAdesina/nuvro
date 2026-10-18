import Foundation
import Network
import NuvroModels
import Observation

@Observable
@MainActor
public final class ConnectivityMonitor {
    public private(set) var status: ConnectivityStatus
    private var monitor: NWPathMonitor?
    private var hasSeenOffline = false

    public init(status: ConnectivityStatus = .online) {
        self.status = status
    }

    public var isOffline: Bool { status != .online }

    public func start() {
        guard monitor == nil else { return }
        let monitor = NWPathMonitor()
        monitor.pathUpdateHandler = { [weak self] path in
            let isSatisfied = path.status == .satisfied
            Task { @MainActor in self?.apply(isSatisfied: isSatisfied) }
        }
        monitor.start(queue: DispatchQueue(label: "com.vectixlogic.nuvro.connectivity"))
        self.monitor = monitor
    }

    public func stop() {
        monitor?.cancel()
        monitor = nil
    }

    public func apply(isSatisfied: Bool) {
        if isSatisfied {
            status = hasSeenOffline && status == .offline ? .reconnecting : .online
            if status == .reconnecting {
                Task { @MainActor in
                    try? await Task.sleep(for: .milliseconds(600))
                    if self.status == .reconnecting { self.status = .online }
                }
            }
        } else {
            hasSeenOffline = true
            status = .offline
        }
    }
}
