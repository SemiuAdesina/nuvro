package user

import "errors"

var ErrSuspended = errors.New("user: account is not active")

const (
	StatusActive    = "active"
	StatusSuspended = "suspended"
	StatusDeleted   = "deleted"
)

type Identity struct {
	Provider string
	Subject  string
}

type Account struct {
	UserID      string
	IdentityID  string
	Provider    string
	Subject     string
	Profile     Profile
	Preferences Preferences
}

func NewAccount(userID, identityID string, identity Identity) (Account, error) {
	if userID == "" || identityID == "" || identity.Provider == "" || identity.Subject == "" {
		return Account{}, ErrInvalidProfile
	}
	profile := DefaultProfile(userID)
	if err := profile.Validate(); err != nil {
		return Account{}, err
	}
	preferences := DefaultPreferences()
	if err := preferences.Validate(); err != nil {
		return Account{}, err
	}
	return Account{
		UserID: userID, IdentityID: identityID, Provider: identity.Provider, Subject: identity.Subject,
		Profile: profile, Preferences: preferences,
	}, nil
}

func IsActive(status string) bool { return status == StatusActive }
