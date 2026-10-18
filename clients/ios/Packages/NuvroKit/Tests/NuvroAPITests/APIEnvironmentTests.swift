import Foundation
import Testing

@testable import NuvroAPI

@Suite("APIEnvironment")
struct APIEnvironmentTests {
    @Test("accepts https origins and loopback http only")
    func schemes() {
        #expect(APIEnvironment(string: "https://api.nuvro.app") != nil)
        #expect(APIEnvironment(string: "http://127.0.0.1:8080") != nil)
        #expect(APIEnvironment(string: "http://localhost:8080") != nil)
        #expect(APIEnvironment(string: "http://api.nuvro.app") == nil)
        #expect(APIEnvironment(string: "ftp://api.nuvro.app") == nil)
        #expect(APIEnvironment(string: "not a url") == nil)
    }

    @Test("a launch argument wins over the bundle")
    func launchArgument() {
        let resolved = APIEnvironment.resolve(arguments: ["Nuvro", "-apiBaseURL", "http://127.0.0.1:8080"], bundle: Bundle(for: Marker.self))
        #expect(resolved?.baseURL.absoluteString == "http://127.0.0.1:8080")
    }

    @Test("nothing configured means no backend")
    func unconfigured() {
        #expect(APIEnvironment.resolve(arguments: ["Nuvro"], bundle: Bundle(for: Marker.self)) == nil)
        #expect(APIEnvironment.resolve(arguments: ["Nuvro", "-apiBaseURL", "http://evil.example"], bundle: Bundle(for: Marker.self)) == nil)
    }

    private final class Marker {}
}
