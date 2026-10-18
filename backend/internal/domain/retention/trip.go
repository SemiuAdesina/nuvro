package retention

import "time"

type Membership struct {
	UserID   string
	JoinedAt time.Time
}

type Succession struct {
	TripID    string
	Successor string
}

func Succeed(tripID string, remaining []Membership) (Succession, bool) {
	var chosen Membership
	for _, member := range remaining {
		if member.UserID == "" {
			continue
		}
		if chosen.UserID == "" || member.JoinedAt.Before(chosen.JoinedAt) ||
			(member.JoinedAt.Equal(chosen.JoinedAt) && member.UserID < chosen.UserID) {
			chosen = member
		}
	}
	if chosen.UserID == "" {
		return Succession{}, false
	}
	return Succession{TripID: tripID, Successor: chosen.UserID}, true
}
