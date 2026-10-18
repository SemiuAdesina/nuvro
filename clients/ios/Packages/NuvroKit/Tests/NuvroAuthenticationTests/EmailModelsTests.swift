import Foundation
import Testing

@testable import NuvroAuthentication

@Suite("EmailEntryModel")
@MainActor
struct EmailEntryModelTests {
    @Test("format validation is lenient but rejects obvious mistakes")
    func format() {
        #expect(EmailFormat.isPlausible("alex@example.com"))
        #expect(EmailFormat.isPlausible(" Alex.R@sub.example.co "))
        #expect(!EmailFormat.isPlausible("alex"))
        #expect(!EmailFormat.isPlausible("alex@example"))
        #expect(!EmailFormat.isPlausible("@example.com"))
        #expect(!EmailFormat.isPlausible("alex @example.com"))
    }

    @Test("errors only show after the traveller has finished editing")
    func errorTiming() {
        let model = EmailEntryModel(client: ScriptedAuthenticationClient())
        #expect(model.phase == .pristine)
        model.email = "alex"
        #expect(model.phase == .editing)
        #expect(!model.showsFormatError)
        model.finishEditing()
        #expect(model.showsFormatError)
        model.email = "alex@example.com"
        #expect(!model.showsFormatError)
        #expect(model.canSubmit)
    }

    @Test("submitting sends one normalised request and returns the code request")
    func submit() async {
        let client = ScriptedAuthenticationClient()
        let model = EmailEntryModel(client: client)
        model.email = " Alex@Example.com "
        let request = await model.submit()
        #expect(request == .fixture)
        #expect(client.codeRequests == ["alex@example.com"])
        #expect(model.phase == .editing)
    }

    @Test("a rejected request is exposed and can be retried")
    func rejection() async {
        let client = ScriptedAuthenticationClient()
        client.codeRequestResult = .failure(.rateLimited(retryAfterSeconds: 30))
        let model = EmailEntryModel(client: client)
        model.email = "alex@example.com"
        #expect(await model.submit() == nil)
        #expect(model.rejection == .rateLimited(retryAfterSeconds: 30))
        #expect(model.canSubmit)
    }

    @Test("an invalid address never reaches the backend")
    func invalidNeverSubmits() async {
        let client = ScriptedAuthenticationClient()
        let model = EmailEntryModel(client: client)
        model.email = "nope"
        #expect(await model.submit() == nil)
        #expect(client.codeRequests.isEmpty)
        #expect(model.showsFormatError)
    }
}

@Suite("EmailVerificationModel")
@MainActor
struct EmailVerificationModelTests {
    private let start = Date(timeIntervalSince1970: 1_800_000_000)

    private func model(client: ScriptedAuthenticationClient = ScriptedAuthenticationClient()) -> EmailVerificationModel {
        let start = start
        return EmailVerificationModel(request: .fixture, client: client, now: { start })
    }

    @Test("entry keeps digits only and truncates to the contract length")
    func entry() {
        let model = model()
        model.enter("12a3 45-6789")
        #expect(model.code == "123456")
        #expect(model.isComplete)
    }

    @Test("verification sends the email and code and yields a session")
    func verify() async {
        let client = ScriptedAuthenticationClient()
        let model = model(client: client)
        model.enter("123456")
        await model.verify()
        #expect(model.session == .fixture)
        #expect(client.verifications.first?.code == "123456")
        #expect(client.verifications.first?.email == "alex@example.com")
    }

    @Test("an incomplete code is never sent")
    func incomplete() async {
        let client = ScriptedAuthenticationClient()
        let model = model(client: client)
        model.enter("123")
        await model.verify()
        #expect(client.verifications.isEmpty)
        #expect(model.phase == .entering)
    }

    @Test("a wrong code clears the entry and reports the failure")
    func wrongCode() async {
        let client = ScriptedAuthenticationClient()
        client.verifyResult = .failure(.invalidCode)
        let model = model(client: client)
        model.enter("000000")
        await model.verify()
        #expect(model.failure == .invalidCode)
        #expect(model.code.isEmpty)
        model.enter("1")
        #expect(model.phase == .entering)
    }

    @Test("resend honours the backend cooldown and refreshes it after a resend")
    func resend() async {
        let client = ScriptedAuthenticationClient()
        let model = model(client: client)
        #expect(!model.canResend(at: start))
        #expect(model.resendCountdown(at: start.addingTimeInterval(10)) == 20)
        await model.resend()
        #expect(client.codeRequests.isEmpty)
        #expect(model.canResend(at: start.addingTimeInterval(30)))
    }

    @Test("a rate limited resend adopts the server's retry window")
    func rateLimitedResend() async {
        let client = ScriptedAuthenticationClient()
        client.codeRequestResult = .failure(.rateLimited(retryAfterSeconds: 120))
        let later = start.addingTimeInterval(60)
        let immediate = EmailCodeRequest(email: "alex@example.com", maskedEmail: "a•••@example.com", codeLength: 6, resendAfterSeconds: 0)
        let model = EmailVerificationModel(request: immediate, client: client, now: { later })
        #expect(model.canResend(at: later))
        await model.resend()
        #expect(model.failure == .rateLimited(retryAfterSeconds: 120))
        #expect(model.resendCountdown(at: later) == 120)
        #expect(!model.canResend(at: later))
    }
}
