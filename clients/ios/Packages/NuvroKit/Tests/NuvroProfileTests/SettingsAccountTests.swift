import NuvroModels
import Testing

@testable import NuvroProfile

@Suite("Settings account actions")
@MainActor
struct SettingsAccountTests {
    private final class Recorder: @unchecked Sendable {
        var signOuts = 0
        var deletions = 0
        var deleteFails = false
    }

    private enum SettingsTestFailure: Error { case backend }

    private func model(recorder: Recorder) -> SettingsModel {
        let actions = AccountActions(
            signOut: { recorder.signOuts += 1 },
            deleteAccount: {
                recorder.deletions += 1
                if recorder.deleteFails { throw SettingsTestFailure.backend }
            }
        )
        return SettingsModel(groups: [SettingsGroup(name: "Privacy", rows: [SettingsRow(name: "Delete account", isDestructive: true)])], notifications: [], permissions: [], actions: actions)
    }

    @Test("sign out asks first and only signs out after confirmation")
    func signOut() async {
        let recorder = Recorder()
        let settings = model(recorder: recorder)
        settings.requestSignOut()
        #expect(settings.presentation == .signOutConfirmation)
        #expect(recorder.signOuts == 0)
        await settings.confirmSignOut()
        #expect(recorder.signOuts == 1)
        #expect(settings.presentation == nil)
    }

    @Test("deletion walks explainer then confirmation and only completes when the backend confirms")
    func deletion() async {
        let recorder = Recorder()
        let settings = model(recorder: recorder)
        settings.select(settings.groups[0].rows[0])
        #expect(settings.presentation == .deleteExplainer)
        settings.proceedToDeleteConfirmation()
        #expect(settings.presentation == .deleteConfirmation)
        recorder.deleteFails = true
        await settings.confirmDelete()
        #expect(recorder.deletions == 1)
        #expect(settings.deleteFailed)
        #expect(settings.presentation == .deleteExplainer)
        recorder.deleteFails = false
        await settings.confirmDelete()
        #expect(recorder.deletions == 2)
        #expect(!settings.deleteFailed)
        #expect(settings.presentation == nil)
    }

    @Test("without account actions the destructive rows are inert")
    func withoutActions() {
        let settings = SettingsModel(groups: [SettingsGroup(name: "Privacy", rows: [SettingsRow(name: "Delete account", isDestructive: true)])], notifications: [], permissions: [])
        #expect(!settings.canManageAccount)
        settings.select(settings.groups[0].rows[0])
        settings.requestSignOut()
        #expect(settings.presentation == nil)
    }
}
