import XCTest

final class LiveBackendUITests: XCTestCase {
    private var baseURL: String { ProcessInfo.processInfo.environment["NUVRO_UITEST_API"] ?? "" }
    private var credential: String { ProcessInfo.processInfo.environment["NUVRO_UITEST_CREDENTIAL"] ?? "" }

    override func setUpWithError() throws {
        try XCTSkipIf(baseURL.isEmpty || credential.isEmpty, "Set NUVRO_UITEST_API and NUVRO_UITEST_CREDENTIAL to run against a live backend")
        continueAfterFailure = false
    }

    private func launch() -> XCUIApplication {
        let app = XCUIApplication()
        app.launchArguments = [
            "-hasFinishedOnboarding", "YES", "-hasCompletedTravelSetup", "YES", "-hasStartedTrip", "NO",
            "-uiTestResetSession", "YES", "-apiBaseURL", baseURL, "-devCredential", credential
        ]
        app.launch()
        return app
    }

    func testADevelopmentCredentialOpensASessionAndTheProfileComesFromTheBackend() {
        let app = launch()
        XCTAssertTrue(app.tabBars.buttons["Profile"].waitForExistence(timeout: 20), "the session was never established")
        app.tabBars.buttons["Profile"].tap()
        XCTAssertTrue(app.staticTexts["Add where you travel from"].waitForExistence(timeout: 10), "the profile did not load from the backend")
        XCTAssertTrue(app.staticTexts["Currency"].exists)
        XCTAssertTrue(app.staticTexts["USD"].exists)
    }

    func testTurningOffTripHistoryIsPersistedByTheBackend() {
        let app = launch()
        let toggle = learningToggle(in: app)
        let before = toggle.value as? String
        toggle.tap()
        wait(for: toggle, toLeave: before)
        let flipped = toggle.value as? String
        XCTAssertNotEqual(flipped, before)
        let prefix = flipped == "0" ? "Off —" : "On —"
        XCTAssertTrue(app.staticTexts.matching(NSPredicate(format: "label BEGINSWITH %@", prefix)).firstMatch.waitForExistence(timeout: 5))
        app.terminate()

        let relaunched = launch()
        let restored = learningToggle(in: relaunched)
        XCTAssertEqual(restored.value as? String, flipped, "the learning switch did not survive a relaunch, so it was never sent to the backend")
    }

    private func learningToggle(in app: XCUIApplication) -> XCUIElement {
        XCTAssertTrue(app.tabBars.buttons["Profile"].waitForExistence(timeout: 20), "the session was never established")
        app.tabBars.buttons["Profile"].tap()
        XCTAssertTrue(app.staticTexts["Currency"].waitForExistence(timeout: 10), "the profile did not load from the backend")
        app.swipeUp()
        let toggle = app.switches.firstMatch
        XCTAssertTrue(toggle.waitForExistence(timeout: 10))
        return toggle
    }

    private func wait(for element: XCUIElement, toLeave value: String?) {
        let changed = expectation(for: NSPredicate(format: "value != %@", value ?? ""), evaluatedWith: element)
        XCTWaiter().wait(for: [changed], timeout: 10)
    }
}
