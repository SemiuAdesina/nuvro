import Testing

@testable import NuvroLocalization
@testable import NuvroOnboarding

@Suite("OnboardingFlowModel")
@MainActor
struct OnboardingFlowModelTests {
    private final class ExitRecorder {
        var exits: [OnboardingExit] = []
    }

    private func makeModel(reduceMotion: Bool = false, variant: HeadlineVariant = .main) -> (OnboardingFlowModel, ExitRecorder) {
        let recorder = ExitRecorder()
        let model = OnboardingFlowModel(reduceMotion: reduceMotion, headlineVariant: variant) { recorder.exits.append($0) }
        return (model, recorder)
    }

    @Test("starts on the splash with nothing booted")
    func initialState() {
        let (model, _) = makeModel()

        #expect(model.stage == .splash)
        #expect(!model.hasBooted)
        #expect(!model.isShowingScreenControls)
        #expect(model.isShowingLanguagePill)
        #expect(model.exit == nil)
    }

    @Test("the splash hands off to the first screen when its hold elapses")
    func splashElapses() {
        let (model, _) = makeModel()

        model.splashDidElapse()

        #expect(model.stage == .discover)
        #expect(model.isShowingScreenControls)
    }

    @Test("a late splash timer is ignored once the traveller has moved on")
    func staleSplashTimer() {
        let (model, _) = makeModel()
        model.advance()
        model.advance()

        model.splashDidElapse()

        #expect(model.stage == .plan)
    }

    @Test("the call to action walks the three screens then finishes with Find my vacation")
    func callToActionWalk() {
        let (model, recorder) = makeModel()
        model.splashDidElapse()

        #expect(model.callToActionLabel == "Show me where to go")
        model.advance()
        #expect(model.stage == .plan)
        #expect(model.callToActionLabel == "Continue")
        model.advance()
        #expect(model.stage == .travel)
        #expect(model.callToActionLabel == "Find my vacation")
        model.advance()

        #expect(model.exit == .findVacation)
        #expect(recorder.exits == [.findVacation])
        #expect(!model.isShowingScreenControls)
        #expect(!model.isShowingLanguagePill)
    }

    @Test("finishing is idempotent")
    func finishOnce() {
        let (model, recorder) = makeModel()
        model.splashDidElapse()
        model.skip()
        model.skip()
        model.advance()

        #expect(recorder.exits == [.skipped])
    }

    @Test("skip is unavailable on the splash and exits from any screen")
    func skip() {
        let (model, recorder) = makeModel()
        model.skip()
        #expect(model.exit == nil)

        model.splashDidElapse()
        model.skip()

        #expect(model.exit == .skipped)
        #expect(recorder.exits == [.skipped])
    }

    @Test("I already have a trip only exists on the final screen")
    func existingTrip() {
        let (model, _) = makeModel()
        model.splashDidElapse()
        model.chooseExistingTrip()
        #expect(model.exit == nil)

        model.advance()
        model.advance()
        model.chooseExistingTrip()

        #expect(model.exit == .existingTrip)
    }

    @Test("retreat never returns to the splash")
    func retreat() {
        let (model, _) = makeModel()
        model.splashDidElapse()
        model.retreat()
        #expect(model.stage == .discover)

        model.advance()
        model.retreat()
        #expect(model.stage == .discover)
    }

    @Test("the replan only fires on the travel screen and resets when leaving")
    func replan() {
        let (model, _) = makeModel()
        model.splashDidElapse()
        model.replanDidElapse()
        #expect(!model.hasReplanned)

        model.advance()
        model.advance()
        model.replanDidElapse()
        #expect(model.hasReplanned)

        model.retreat()
        #expect(!model.hasReplanned)
    }

    @Test("a left swipe past the threshold advances and a right swipe retreats")
    func swipes() {
        let (model, _) = makeModel()
        model.splashDidElapse()

        model.dragEnded(-61)
        #expect(model.stage == .plan)
        model.dragEnded(-59)
        #expect(model.stage == .plan)
        model.dragEnded(61)
        #expect(model.stage == .discover)
        model.dragEnded(61)
        #expect(model.stage == .discover)
    }

    @Test("a left swipe on the final screen does not finish the flow")
    func swipeOnFinalScreen() {
        let (model, _) = makeModel()
        model.splashDidElapse()
        model.advance()
        model.advance()

        model.dragEnded(-200)

        #expect(model.stage == .travel)
        #expect(model.exit == nil)
    }

    @Test("drag offset is clamped, ignores the dead zone, and resets on release")
    func dragOffset() {
        let (model, _) = makeModel()
        model.splashDidElapse()

        model.dragChanged(3)
        #expect(model.dragOffset == 0)
        model.dragChanged(-140)
        #expect(model.dragOffset == -80)
        model.dragChanged(30)
        #expect(model.dragOffset == 30)
        model.dragEnded(30)
        #expect(model.dragOffset == 0)
    }

    @Test("dragging is disabled on the splash and while the language sheet is open")
    func dragGuards() {
        let (model, _) = makeModel()
        model.dragChanged(-70)
        #expect(model.dragOffset == 0)

        model.splashDidElapse()
        model.presentLanguageSheet()
        model.dragChanged(-70)
        model.dragEnded(-70)

        #expect(model.dragOffset == 0)
        #expect(model.stage == .discover)
    }

    @Test("choosing a language closes the sheet, clears the query, and changes the copy")
    func languageSelection() {
        let (model, _) = makeModel()
        model.presentLanguageSheet()
        model.languageQuery = "esp"
        #expect(model.filteredLanguages == [.spanish])

        model.select(.spanish)

        #expect(model.language == .spanish)
        #expect(!model.isLanguageSheetPresented)
        #expect(model.languageQuery.isEmpty)
        #expect(model.copy.skip == "Omitir")
        #expect(model.callToActionLabel == "Muéstrame adónde ir")
    }

    @Test("the alternate headline swaps the final screen's headline and supporting line")
    func headlineVariant() {
        let (main, _) = makeModel(variant: .main)
        let (alternate, _) = makeModel(variant: .alternate)

        #expect(main.headline3 == "Your trip doesn’t stop when you arrive.")
        #expect(main.supporting3 == "And when plans change, so does your trip.")
        #expect(alternate.headline3 == "And when plans change, so does your trip.")
        #expect(alternate.supporting3 == "Your trip doesn’t stop when you arrive.")
    }

    @Test("reduce motion pins the camera to the first screen pose")
    func reduceMotionCamera() {
        let (model, _) = makeModel(reduceMotion: true)
        model.splashDidElapse()
        model.advance()
        model.advance()

        #expect(model.stage == .travel)
        #expect(model.cameraPose == CameraPose.byStage[.discover])
    }
}
