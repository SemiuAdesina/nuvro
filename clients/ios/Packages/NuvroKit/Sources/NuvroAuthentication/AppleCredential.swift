import Foundation
#if canImport(AuthenticationServices)
import AuthenticationServices
#endif

public struct AppleCredential: Sendable, Hashable {
    public let userIdentifier: String
    public let identityToken: String
    public let authorizationCode: String
    public let email: String?
    public let givenName: String?
    public let familyName: String?

    public init(userIdentifier: String, identityToken: String, authorizationCode: String, email: String? = nil, givenName: String? = nil, familyName: String? = nil) {
        self.userIdentifier = userIdentifier
        self.identityToken = identityToken
        self.authorizationCode = authorizationCode
        self.email = email
        self.givenName = givenName
        self.familyName = familyName
    }
}

extension AppleCredential: CustomStringConvertible {
    public var description: String {
        "AppleCredential(user: \(userIdentifier.prefix(6))…, email: \(email == nil ? "absent" : "present"))"
    }
}

public enum AppleAuthorizationOutcome: Equatable, Sendable {
    case authorized(AppleCredential)
    case cancelled
    case failed
}

public enum AppleAuthorizationMapper {
    #if canImport(AuthenticationServices)
    public static func outcome(from result: Result<ASAuthorization, any Error>) -> AppleAuthorizationOutcome {
        switch result {
        case .success(let authorization):
            guard let appleID = authorization.credential as? ASAuthorizationAppleIDCredential,
                  let token = appleID.identityToken.flatMap({ String(data: $0, encoding: .utf8) }),
                  let code = appleID.authorizationCode.flatMap({ String(data: $0, encoding: .utf8) }) else { return .failed }
            return .authorized(AppleCredential(
                userIdentifier: appleID.user, identityToken: token, authorizationCode: code,
                email: appleID.email, givenName: appleID.fullName?.givenName, familyName: appleID.fullName?.familyName
            ))
        case .failure(let error):
            return outcome(for: error)
        }
    }
    #endif

    public static func outcome(for error: any Error) -> AppleAuthorizationOutcome {
        let nsError = error as NSError
        #if canImport(AuthenticationServices)
        if nsError.domain == ASAuthorizationError.errorDomain, nsError.code == ASAuthorizationError.canceled.rawValue { return .cancelled }
        #endif
        return isCancellation(domain: nsError.domain, code: nsError.code) ? .cancelled : .failed
    }

    public static func isCancellation(domain: String, code: Int) -> Bool {
        domain == "com.apple.AuthenticationServices.AuthorizationError" && code == 1001
    }
}
