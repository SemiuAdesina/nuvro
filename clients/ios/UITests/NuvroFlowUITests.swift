import XCTest

final class NuvroFlowUITests: XCTestCase {
    private func launch(finishedOnboarding: Bool) -> XCUIApplication {
        let app = XCUIApplication()
        app.launchArguments = finishedOnboarding
            ? ["-hasFinishedOnboarding", "YES", "-hasCompletedTravelSetup", "YES", "-hasStartedTrip", "NO", "-uiTestResetSession", "YES"]
            : ["-uiTestResetOnboarding", "YES", "-uiTestResetTravelSetup", "YES", "-hasStartedTrip", "NO", "-uiTestResetSession", "YES"]
        app.launch()
        return app
    }

    func testNewTravellerGoesFromAuthenticationToSetupToStory() {
        let app = launch(finishedOnboarding: false)
        XCTAssertTrue(app.buttons["Continue with Email"].waitForExistence(timeout: 10))
        app.buttons["Continue with Email"].tap()
        app.textFields["Email address"].typeText("new@example.com")
        app.buttons["Continue"].firstMatch.tap()
        let code = app.textFields["Verification code"]
        XCTAssertTrue(code.waitForExistence(timeout: 5))
        code.tap()
        code.typeText("123456")
        XCTAssertTrue(app.staticTexts["Let’s make NUVRO yours."].waitForExistence(timeout: 10))
        app.buttons["Continue"].firstMatch.tap()
        XCTAssertTrue(app.staticTexts["Where do you live?"].waitForExistence(timeout: 5))
        app.buttons["Choose a country"].tap()
        let search = app.searchFields.firstMatch
        XCTAssertTrue(search.waitForExistence(timeout: 5))
        search.tap()
        search.typeText("Nigeria")
        app.buttons.matching(NSPredicate(format: "label CONTAINS[c] 'Nigeria'")).firstMatch.tap()
        app.buttons["Continue"].firstMatch.tap()
        XCTAssertTrue(app.staticTexts["Which passport do you travel with?"].waitForExistence(timeout: 5))
        XCTAssertFalse(app.staticTexts["Nigeria"].exists)
        app.buttons["Add passport"].tap()
        let passportSearch = app.searchFields.firstMatch
        XCTAssertTrue(passportSearch.waitForExistence(timeout: 5))
        passportSearch.tap()
        passportSearch.typeText("Nigeria")
        app.buttons.matching(NSPredicate(format: "label CONTAINS[c] 'Nigeria'")).firstMatch.tap()
        app.buttons["Continue"].firstMatch.tap()
        XCTAssertTrue(app.buttons["Skip for now"].waitForExistence(timeout: 5))
        app.buttons["Skip for now"].tap()
        XCTAssertTrue(app.staticTexts["Where do you usually fly from?"].waitForExistence(timeout: 5))
    }

    func testAppleButtonIsVisibleOnAuthentication() {
        let app = launch(finishedOnboarding: true)
        XCTAssertTrue(app.buttons["Continue with Email"].waitForExistence(timeout: 8))
        XCTAssertTrue(app.buttons.matching(NSPredicate(format: "label CONTAINS[c] 'Apple'")).firstMatch.exists)
    }

    func testEmailFlowReachesDiscover() {
        let app = launch(finishedOnboarding: true)
        XCTAssertTrue(app.buttons["Continue with Email"].waitForExistence(timeout: 8))
        app.buttons["Continue with Email"].tap()
        let email = app.textFields["Email address"]
        XCTAssertTrue(email.waitForExistence(timeout: 5))
        email.typeText("alex@example.com")
        app.buttons["Continue"].firstMatch.tap()
        let code = app.textFields["Verification code"]
        XCTAssertTrue(code.waitForExistence(timeout: 5))
        code.tap()
        code.typeText("123456")
        XCTAssertTrue(app.staticTexts["Where should we take you?"].waitForExistence(timeout: 10))
    }

    func testEmptyTripsAndSignOut() {
        let app = launch(finishedOnboarding: true)
        XCTAssertTrue(app.buttons["Continue with Email"].waitForExistence(timeout: 8))
        app.buttons["Continue with Email"].tap()
        app.textFields["Email address"].typeText("alex@example.com")
        app.buttons["Continue"].firstMatch.tap()
        let code = app.textFields["Verification code"]
        XCTAssertTrue(code.waitForExistence(timeout: 5))
        code.tap()
        code.typeText("123456")
        XCTAssertTrue(app.tabBars.buttons["Trips"].waitForExistence(timeout: 10))
        app.tabBars.buttons["Trips"].tap()
        XCTAssertTrue(app.staticTexts["No trips yet"].waitForExistence(timeout: 5))
        app.tabBars.buttons["Profile"].tap()
        app.buttons["Settings"].firstMatch.tap()
        let signOut = app.buttons["Sign out"]
        XCTAssertTrue(signOut.waitForExistence(timeout: 5))
        signOut.tap()
        XCTAssertTrue(app.sheets.buttons["Sign out"].waitForExistence(timeout: 3) || app.buttons["Sign out"].waitForExistence(timeout: 3))
        app.sheets.buttons["Sign out"].firstMatch.tap()
        XCTAssertTrue(app.buttons["Continue with Email"].waitForExistence(timeout: 8))
    }
}

final class NuvroSupportingFlowUITests: XCTestCase {
    private func launchSignedIn(extraArguments: [String] = []) -> XCUIApplication {
        let app = XCUIApplication()
        app.launchArguments = ["-hasFinishedOnboarding", "YES", "-hasCompletedTravelSetup", "YES", "-hasStartedTrip", "YES", "-uiTestResetSession", "YES"] + extraArguments
        app.launch()
        XCTAssertTrue(app.buttons["Continue with Email"].waitForExistence(timeout: 8))
        app.buttons["Continue with Email"].tap()
        app.textFields["Email address"].typeText("alex@example.com")
        app.buttons["Continue"].firstMatch.tap()
        let code = app.textFields["Verification code"]
        XCTAssertTrue(code.waitForExistence(timeout: 5))
        code.tap()
        code.typeText("123456")
        XCTAssertTrue(app.tabBars.buttons["Trips"].waitForExistence(timeout: 10))
        return app
    }

    func testDeleteAccountRequiresExplainerAndConfirmation() {
        let app = launchSignedIn()
        app.tabBars.buttons["Profile"].tap()
        app.buttons["Settings"].firstMatch.tap()
        let deleteRow = app.buttons["Delete account"].firstMatch
        XCTAssertTrue(deleteRow.waitForExistence(timeout: 5))
        deleteRow.tap()
        XCTAssertTrue(app.navigationBars["Delete your account"].waitForExistence(timeout: 5))
        app.buttons["Delete account"].firstMatch.tap()
        let alert = app.alerts["Delete your account?"]
        XCTAssertTrue(alert.waitForExistence(timeout: 5))
        alert.buttons["Cancel"].tap()
        XCTAssertTrue(app.navigationBars["Delete your account"].exists)
        app.buttons["Cancel"].firstMatch.tap()
        XCTAssertTrue(app.buttons["Sign out"].waitForExistence(timeout: 5))
    }

    func testLocationDeniedOffersDestinationCentre() {
        let app = launchSignedIn(extraArguments: ["-uiTestPermissions", "denied"])
        app.tabBars.buttons["Trips"].tap()
        let continueTrip = app.buttons["Continue trip"]
        XCTAssertTrue(continueTrip.waitForExistence(timeout: 5))
        continueTrip.tap()
        let navigate = app.buttons["Navigate"].firstMatch
        XCTAssertTrue(navigate.waitForExistence(timeout: 8))
        navigate.tap()
        XCTAssertTrue(app.staticTexts["Location is off"].waitForExistence(timeout: 8))
        app.buttons["Use destination centre"].tap()
        XCTAssertTrue(app.staticTexts["Location is off"].waitForNonExistence(timeout: 5))
    }

    func testNextCardTicketOpensWalletAndAddToWalletReportsHonestly() {
        let app = launchSignedIn()
        app.tabBars.buttons["Trips"].tap()
        let continueTrip = app.buttons["Continue trip"]
        XCTAssertTrue(continueTrip.waitForExistence(timeout: 5))
        continueTrip.tap()
        let ticket = app.buttons["Ticket"].firstMatch
        XCTAssertTrue(ticket.waitForExistence(timeout: 8))
        ticket.tap()
        let addToWallet = app.buttons["Add to Wallet"].firstMatch
        XCTAssertTrue(addToWallet.waitForExistence(timeout: 8))
        addToWallet.tap()
        let notice = app.staticTexts.matching(NSPredicate(format: "label CONTAINS[c] 'Wallet passes'")).firstMatch
        XCTAssertTrue(notice.waitForExistence(timeout: 3))
        XCTAssertTrue(notice.waitForNonExistence(timeout: 6))
        addToWallet.tap()
        XCTAssertTrue(notice.waitForExistence(timeout: 3))
    }

    func testLearningCardToggleWorksAndIllustrationIsHiddenFromVoiceOver() {
        let app = launchSignedIn()
        app.tabBars.buttons["Profile"].tap()
        let toggle = app.switches.matching(NSPredicate(format: "label CONTAINS[c] 'Use past trips'")).firstMatch
        XCTAssertTrue(toggle.waitForExistence(timeout: 8))
        app.swipeUp()
        let before = toggle.value as? String
        toggle.tap()
        XCTAssertNotEqual(toggle.value as? String, before)
        XCTAssertTrue(app.staticTexts["What NUVRO has learned"].exists)
        XCTAssertEqual(app.images.matching(NSPredicate(format: "identifier CONTAINS[c] 'attendant' OR label CONTAINS[c] 'attendant'")).count, 0)
    }
}
