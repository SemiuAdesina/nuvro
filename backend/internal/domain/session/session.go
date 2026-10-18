package session

import (
	"crypto/sha256"
	"time"
)

const AccessLifetime = 15 * time.Minute
const RefreshLifetime = 90 * 24 * time.Hour
const Provider = "session"

type Session struct {
	ID        string
	UserID    string
	FamilyID  string
	IssuedAt  time.Time
	ExpiresAt time.Time
	RotatedAt *time.Time
	RevokedAt *time.Time
}

func (s Session) Usable(now time.Time) bool {
	return s.RevokedAt == nil && s.RotatedAt == nil && now.Before(s.ExpiresAt)
}

func (s Session) Reused() bool { return s.RotatedAt != nil || s.RevokedAt != nil }

type Grant struct {
	SessionID    string
	UserID       string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

func HashRefreshToken(token string) []byte {
	digest := sha256.Sum256([]byte(token))
	return digest[:]
}

func Start(sessionID, familyID, userID string, now time.Time) Session {
	return Session{ID: sessionID, UserID: userID, FamilyID: familyID, IssuedAt: now.UTC(), ExpiresAt: now.Add(RefreshLifetime).UTC()}
}
