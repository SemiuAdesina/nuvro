import Foundation

public struct APIEnvironment: Sendable, Hashable {
    public static let bundleKey = "NuvroAPIBaseURL"
    public static let argument = "-apiBaseURL"

    public let baseURL: URL

    public init(baseURL: URL) {
        self.baseURL = baseURL
    }

    public init?(string: String) {
        guard let url = URL(string: string), let scheme = url.scheme, url.host() != nil else { return nil }
        guard scheme == "https" || (scheme == "http" && Self.isLoopback(url)) else { return nil }
        self.init(baseURL: url)
    }

    public static func resolve(arguments: [String] = CommandLine.arguments, bundle: Bundle = .main) -> APIEnvironment? {
        if let index = arguments.firstIndex(of: argument), arguments.indices.contains(index + 1),
           let environment = APIEnvironment(string: arguments[index + 1]) {
            return environment
        }
        guard let configured = bundle.object(forInfoDictionaryKey: bundleKey) as? String, !configured.isEmpty else { return nil }
        return APIEnvironment(string: configured)
    }

    private static func isLoopback(_ url: URL) -> Bool {
        guard let host = url.host() else { return false }
        return host == "localhost" || host == "127.0.0.1" || host == "::1"
    }
}
