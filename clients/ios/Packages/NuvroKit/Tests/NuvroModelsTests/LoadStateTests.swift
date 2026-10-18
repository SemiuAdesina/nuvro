import Foundation
import Testing

@testable import NuvroModels

@Suite("LoadState")
struct LoadStateTests {
    @Test("beginning a load keeps existing content as refreshing")
    func refreshKeepsContent() {
        let loaded = LoadState<[String]>.loaded(["Madeira"])
        #expect(loaded.beginning() == .refreshing(["Madeira"]))
        #expect(LoadState<[String]>.idle.beginning() == .loading)
        #expect(loaded.beginning().isLoading)
        #expect(loaded.beginning().value == ["Madeira"])
    }

    @Test("a failure keeps cached content and exposes the failure")
    func failureKeepsCache() {
        let failed = LoadState<[String]>.loaded(["Madeira"]).failing(.offline)
        #expect(failed == .failed(.offline, cached: ["Madeira"]))
        #expect(failed.value == ["Madeira"])
        #expect(failed.failure == .offline)
        #expect(LoadState<[String]>.loading.failing(.unexpected).value == nil)
    }

    @Test("API and transport errors map to load failures")
    func mapping() {
        #expect(LoadFailure(apiFailure: APIFailure(status: 503, code: .upstreamUnavailable, detail: "")) == .unavailable)
        #expect(LoadFailure(apiFailure: APIFailure(status: 401, code: .authenticationRequired, detail: "")) == .unauthorized)
        #expect(LoadFailure(error: URLError(.notConnectedToInternet)) == .offline)
        #expect(LoadFailure(error: URLError(.badURL)) == .unexpected)
        #expect(!LoadFailure.unauthorized.isRetryable)
        #expect(LoadFailure.offline.isRetryable)
    }
}
