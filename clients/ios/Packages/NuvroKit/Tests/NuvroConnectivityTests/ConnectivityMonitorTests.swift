import NuvroModels
import Testing

@testable import NuvroConnectivity

@Suite("ConnectivityMonitor")
@MainActor
struct ConnectivityMonitorTests {
    @Test("a lost path reports offline and a restored path passes through reconnecting")
    func transitions() async throws {
        let monitor = ConnectivityMonitor()
        #expect(monitor.status == .online)
        monitor.apply(isSatisfied: false)
        #expect(monitor.status == .offline)
        #expect(monitor.isOffline)
        monitor.apply(isSatisfied: true)
        #expect(monitor.status == .reconnecting)
        try await Task.sleep(for: .milliseconds(900))
        #expect(monitor.status == .online)
    }

    @Test("a satisfied path that was never offline stays online")
    func steady() {
        let monitor = ConnectivityMonitor()
        monitor.apply(isSatisfied: true)
        #expect(monitor.status == .online)
    }
}
