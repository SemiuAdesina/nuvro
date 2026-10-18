import Testing

@testable import NuvroPermissions

@Suite("PermissionCenter")
@MainActor
struct PermissionCenterTests {
    @Test("granted permissions proceed without any presentation")
    func granted() async {
        let center = PermissionCenter(authority: ScriptedPermissionAuthority(states: [.location: .granted]))
        #expect(await center.prepare(.location))
        #expect(center.presentation == nil)
        #expect(center.state(of: .location) == .granted)
    }

    @Test("an undetermined permission shows the explainer before the system prompt")
    func explainer() async {
        let center = PermissionCenter(authority: ScriptedPermissionAuthority(states: [.camera: .notDetermined], requestResults: [.camera: .granted]))
        #expect(await center.prepare(.camera) == false)
        #expect(center.presentation == .explainer(.camera))
        #expect(await center.proceedFromExplainer() == .granted)
        #expect(center.presentation == nil)
        #expect(center.state(of: .camera) == .granted)
    }

    @Test("a denial after the prompt leads to settings guidance")
    func deniedAfterPrompt() async {
        let center = PermissionCenter(authority: ScriptedPermissionAuthority(states: [.microphone: .notDetermined], requestResults: [.microphone: .denied]))
        _ = await center.prepare(.microphone)
        #expect(await center.proceedFromExplainer() == .denied)
        #expect(center.presentation == .deniedGuidance(.microphone))
        center.dismissPresentation()
        #expect(center.presentation == nil)
    }

    @Test("previously denied permissions go straight to guidance")
    func alreadyDenied() async {
        let center = PermissionCenter(authority: ScriptedPermissionAuthority(states: [.notifications: .denied]))
        #expect(await center.prepare(.notifications) == false)
        #expect(center.presentation == .deniedGuidance(.notifications))
        #expect(center.state(of: .notifications).needsSettings)
    }
}
