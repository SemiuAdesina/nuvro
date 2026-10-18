import Foundation
import Testing

@testable import NuvroDesignSystem

@Suite("TransientFeedback")
@MainActor
struct TransientFeedbackTests {
    @Test("showing replaces the current message instead of stacking")
    func replaces() {
        let feedback = TransientFeedback(holdDuration: .seconds(60))
        feedback.show("Saved")
        let first = feedback.current
        feedback.show("Added to trip", symbol: "plus")
        #expect(feedback.current?.text == "Added to trip")
        #expect(feedback.current?.symbol == "plus")
        #expect(feedback.current?.id != first?.id)
    }

    @Test("messages dismiss after the hold duration")
    func autoDismiss() async throws {
        let feedback = TransientFeedback(holdDuration: .milliseconds(30))
        feedback.show("Saved")
        #expect(feedback.current != nil)
        try await Task.sleep(for: .milliseconds(200))
        #expect(feedback.current == nil)
    }

    @Test("manual dismissal clears immediately")
    func dismiss() {
        let feedback = TransientFeedback(holdDuration: .seconds(60))
        feedback.show("Saved")
        feedback.dismiss()
        #expect(feedback.current == nil)
    }
}
