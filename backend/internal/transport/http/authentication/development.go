package authentication

import (
	"context"
	"crypto/subtle"
	"errors"

	"github.com/SemiuAdesina/nuvro/backend/internal/domain/user"
)

const DevelopmentProvider = "development"

var ErrDevelopmentUnavailable = errors.New("authentication: the development verifier is only available in development")

type DevelopmentVerifier struct {
	credential string
	subject    string
}

func NewDevelopmentVerifier(isDevelopment bool, credential, subject string) (*DevelopmentVerifier, error) {
	if !isDevelopment {
		return nil, ErrDevelopmentUnavailable
	}
	if len(credential) < 16 || subject == "" {
		return nil, errors.New("authentication: the development credential must be at least 16 characters with a subject")
	}
	return &DevelopmentVerifier{credential: credential, subject: subject}, nil
}

func (v *DevelopmentVerifier) Verify(_ context.Context, credential string) (user.Identity, error) {
	if subtle.ConstantTimeCompare([]byte(credential), []byte(v.credential)) != 1 {
		return user.Identity{}, errors.New("authentication: credential rejected")
	}
	return user.Identity{Provider: DevelopmentProvider, Subject: v.subject}, nil
}
