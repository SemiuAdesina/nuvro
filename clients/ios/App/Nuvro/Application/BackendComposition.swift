import Foundation
import NuvroAPI
import NuvroAuthentication

@MainActor
struct BackendComposition {
    let authentication: AuthenticationCoordinator
    let account: (any AccountGateway)?
    let discovery: (any DiscoverGateway)?
    let developmentCredential: String?

    static func resolve(arguments: [String] = CommandLine.arguments) -> BackendComposition {
        let store = KeychainSessionStore()
        guard let environment = APIEnvironment.resolve(arguments: arguments) else {
            return BackendComposition(authentication: AuthenticationCoordinator(client: PrototypeAuthenticationClient(), store: store), account: nil, discovery: nil, developmentCredential: nil)
        }
        let tokens = SessionTokenBox()
        let client = NuvroAPIClient.make(environment: environment, credentials: tokens)
        let sessions = NuvroSessionGateway(client: client)
        let account = NuvroAccountGateway(client: client)
        let coordinator = AuthenticationCoordinator(client: LiveAuthenticationClient(sessions: sessions, account: account), store: store)
        tokens.install(coordinator)
        return BackendComposition(authentication: coordinator, account: account, discovery: NuvroDiscoverGateway(client: client), developmentCredential: developmentCredential(from: arguments))
    }

    private static func developmentCredential(from arguments: [String]) -> String? {
        #if DEBUG
        guard let index = arguments.firstIndex(of: "-devCredential"), arguments.indices.contains(index + 1) else { return nil }
        return arguments[index + 1]
        #else
        return nil
        #endif
    }
}
