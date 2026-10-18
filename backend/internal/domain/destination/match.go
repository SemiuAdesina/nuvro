package destination

import "errors"

type Assessment string

const (
	Strong  Assessment = "strong"
	Good    Assessment = "good"
	Neutral Assessment = "neutral"
	Caution Assessment = "caution"
)

type MatchComponent struct {
	Key            string
	Label          string
	AwardedPoints  int
	PossiblePoints int
	Assessment     Assessment
	Explanation    string
}

type MatchScore struct {
	Score           int
	Components      []MatchComponent
	Summary         string
	WeightsVersion  string
	EvidenceVersion string
}

func (m MatchScore) Validate() error {
	if m.Score < 0 || m.Score > 100 || m.WeightsVersion == "" || m.EvidenceVersion == "" {
		return errors.New("match: invalid score or provenance")
	}
	for _, component := range m.Components {
		if component.AwardedPoints < 0 || component.PossiblePoints < component.AwardedPoints || component.PossiblePoints > 100 {
			return errors.New("match: invalid component points")
		}
		switch component.Assessment {
		case Strong, Good, Neutral, Caution:
		default:
			return errors.New("match: invalid assessment")
		}
	}
	return nil
}
