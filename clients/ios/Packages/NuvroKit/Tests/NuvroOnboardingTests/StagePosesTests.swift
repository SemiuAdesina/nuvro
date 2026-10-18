import Testing

@testable import NuvroOnboarding

@Suite("StagePoses")
struct StagePosesTests {
    @Test("the primary card rises from below into frame and grows on the plan screen")
    func primaryCard() {
        #expect(StagePoses.primaryCard(.splash) == ContentPose(101, 300, 0.6, 0))
        #expect(StagePoses.primaryCard(.discover) == ContentPose(101, 118, 1, 1))
        #expect(StagePoses.primaryCard(.plan) == ContentPose(71, 112, 1.3, 1))
        #expect(!StagePoses.primaryCard(.travel).isVisible)
    }

    @Test("the companion cards only show on the first screen")
    func companionCards() {
        for stage in OnboardingStage.allCases {
            let expected = stage == .discover
            #expect(StagePoses.secondaryCard(stage).isVisible == expected)
            #expect(StagePoses.tertiaryCard(stage).isVisible == expected)
        }
        #expect(StagePoses.secondaryCard(.discover) == ContentPose(14, 292, 0.88, 0.94))
        #expect(StagePoses.tertiaryCard(.discover) == ContentPose(248, 318, 0.88, 0.94))
    }

    @Test("the world dims behind the live trip and nowhere else")
    func worldOpacity() {
        #expect(StagePoses.worldOpacity(.travel) == 0.18)
        #expect(StagePoses.worldOpacity(.splash) == 1)
        #expect(StagePoses.worldOpacity(.plan) == 1)
    }

    @Test("content drifts toward its screen from the direction of travel")
    func drift() {
        #expect(StagePoses.verticalDrift(for: .discover, shownAt: .plan, distance: 30) == 30)
        #expect(StagePoses.verticalDrift(for: .plan, shownAt: .plan, distance: 30) == 0)
        #expect(StagePoses.verticalDrift(for: .travel, shownAt: .plan, distance: 30) == -30)
        #expect(StagePoses.horizontalSlide(for: .discover, shownAt: .travel) == 40)
        #expect(StagePoses.horizontalSlide(for: .travel, shownAt: .discover) == -40)
    }

    @Test("gesture thresholds match the prototype")
    func gestureConstants() {
        #expect(StagePoses.dragLimit == 80)
        #expect(StagePoses.dragDeadZone == 4)
        #expect(StagePoses.swipeThreshold == 60)
        #expect(StagePoses.dragParallaxRatio == 0.35)
    }
}
