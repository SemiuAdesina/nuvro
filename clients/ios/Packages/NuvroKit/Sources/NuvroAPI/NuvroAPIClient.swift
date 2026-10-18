import Foundation
import OpenAPIRuntime
import OpenAPIURLSession

public enum NuvroAPIClient {
    public static func make(
        environment: APIEnvironment, credentials: any SessionCredentials,
        appVersion: String = Bundle.main.object(forInfoDictionaryKey: "CFBundleShortVersionString") as? String ?? "0",
        build: String = Bundle.main.object(forInfoDictionaryKey: "CFBundleVersion") as? String ?? "0",
        transport: (any ClientTransport)? = nil
    ) -> Client {
        Client(
            serverURL: environment.baseURL,
            configuration: Configuration(dateTranscoder: LenientDateTranscoder()),
            transport: transport ?? URLSessionTransport(),
            middlewares: [ClientIdentityMiddleware(appVersion: appVersion, build: build), BearerMiddleware(credentials: credentials)]
        )
    }
}
