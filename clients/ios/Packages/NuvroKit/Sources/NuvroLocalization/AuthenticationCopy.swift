public struct AuthenticationCopy: Equatable, Sendable {
    public let tagline: String
    public let continueWithEmail: String
    public let or: String
    public let legalPrefix: String
    public let terms: String
    public let privacy: String
    public let emailTitle: String
    public let emailBody: String
    public let emailPlaceholder: String
    public let continueLabel: String
    public let invalidEmail: String
    public let checkEmail: String
    public let enterCodeSentTo: String
    public let codeLabel: String
    public let resend: String
    public let resendIn: String
    public let changeEmail: String
    public let sessionExpired: String
    public let signedOut: String
    public let accountDeleted: String
    public let tryAgain: String
    public let restoring: String
    public let failureInvalidCode: String
    public let failureExpiredCode: String
    public let failureRateLimited: String
    public let failureNetwork: String
    public let failureUnavailable: String
    public let failureUnexpected: String
    public let failureApple: String

    public static let english = AuthenticationCopy(
        tagline: "Your trips, wherever you go.", continueWithEmail: "Continue with Email", or: "or",
        legalPrefix: "By continuing, you agree to NUVRO’s", terms: "Terms", privacy: "Privacy Policy",
        emailTitle: "Continue with email", emailBody: "We’ll send you a verification code.", emailPlaceholder: "Email address", continueLabel: "Continue",
        invalidEmail: "That doesn’t look like an email address.", checkEmail: "Check your email", enterCodeSentTo: "Enter the code sent to", codeLabel: "Verification code",
        resend: "Resend code", resendIn: "Resend in", changeEmail: "Change email",
        sessionExpired: "Your session expired. Sign in again to pick up where you left off.", signedOut: "You’re signed out. Your saved trips are still on this device.",
        accountDeleted: "Your account has been deleted.", tryAgain: "Try again", restoring: "Signing you in…",
        failureInvalidCode: "That code isn’t right. Check it and try again.", failureExpiredCode: "That code has expired. Send a new one.",
        failureRateLimited: "Too many attempts. Wait a moment before trying again.", failureNetwork: "You’re offline. Connect to continue.",
        failureUnavailable: "NUVRO is temporarily unavailable. Try again shortly.", failureUnexpected: "Something went wrong. Try again.",
        failureApple: "Apple couldn’t complete sign-in. Try again or continue with email."
    )

    public static func resolve(_ language: Language) -> AuthenticationCopy { english }
}
