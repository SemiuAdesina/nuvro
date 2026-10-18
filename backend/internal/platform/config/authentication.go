package config

import (
	"encoding/base64"
	"strings"
	"time"
)

const minimumCursorSecretLength = 32
const minimumDevelopmentCredentialLength = 16

type Authentication struct {
	AppleAudience         string
	AppleIssuer           string
	AppleKeyEndpoint      string
	AppleKeyTTL           time.Duration
	DevelopmentCredential string
	DevelopmentSubject    string
	CursorSecret          []byte
	AppleTeamID           string
	AppleKeyID            string
	ApplePrivateKey       []byte
	AppleTokenEndpoint    string
	AppleRevokeEndpoint   string
	TokenSealingKey       []byte
	SessionSigningKey     []byte
}

func (a Authentication) IssuesSessions() bool {
	return len(a.SessionSigningKey) >= minimumCursorSecretLength
}

func (a Authentication) RevokesAppleTokens() bool {
	return a.AppleTeamID != "" && a.AppleKeyID != "" && len(a.ApplePrivateKey) > 0 && len(a.TokenSealingKey) > 0
}

func (a Authentication) UsesApple() bool { return a.AppleAudience != "" }

func (a Authentication) UsesDevelopmentCredential() bool { return a.DevelopmentCredential != "" }

func (r *reader) authentication() Authentication {
	return Authentication{
		AppleAudience:         r.text("NUVRO_APPLE_AUDIENCE", ""),
		AppleIssuer:           r.text("NUVRO_APPLE_ISSUER", "https://appleid.apple.com"),
		AppleKeyEndpoint:      r.text("NUVRO_APPLE_KEY_ENDPOINT", "https://appleid.apple.com/auth/keys"),
		AppleKeyTTL:           r.duration("NUVRO_APPLE_KEY_TTL", time.Hour),
		DevelopmentCredential: r.text("NUVRO_DEV_AUTH_CREDENTIAL", ""),
		DevelopmentSubject:    r.text("NUVRO_DEV_AUTH_SUBJECT", "local-traveller"),
		CursorSecret:          r.secret("NUVRO_CURSOR_SECRET"),
		AppleTeamID:           r.text("NUVRO_APPLE_TEAM_ID", ""),
		AppleKeyID:            r.text("NUVRO_APPLE_KEY_ID", ""),
		ApplePrivateKey:       r.secret("NUVRO_APPLE_PRIVATE_KEY"),
		AppleTokenEndpoint:    r.text("NUVRO_APPLE_TOKEN_ENDPOINT", "https://appleid.apple.com/auth/token"),
		AppleRevokeEndpoint:   r.text("NUVRO_APPLE_REVOKE_ENDPOINT", "https://appleid.apple.com/auth/revoke"),
		TokenSealingKey:       r.secret("NUVRO_TOKEN_SEALING_KEY"),
		SessionSigningKey:     r.secret("NUVRO_SESSION_SIGNING_KEY"),
	}
}

func (r *reader) secret(key string) []byte {
	raw := r.text(key, "")
	if raw == "" {
		return nil
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		r.problems = append(r.problems, key+" must be base64 encoded")
		return nil
	}
	return decoded
}

func (c Config) authenticationProblems() []string {
	var problems []string
	authentication := c.Authentication
	if authentication.UsesDevelopmentCredential() {
		if c.Environment != EnvironmentDevelopment {
			problems = append(problems, "NUVRO_DEV_AUTH_CREDENTIAL is only permitted when NUVRO_ENV is development")
		}
		if len(authentication.DevelopmentCredential) < minimumDevelopmentCredentialLength {
			problems = append(problems, "NUVRO_DEV_AUTH_CREDENTIAL must be at least 16 characters")
		}
	}
	if authentication.AppleKeyTTL <= 0 {
		problems = append(problems, "NUVRO_APPLE_KEY_TTL must be a positive duration")
	}
	if endpoint := authentication.AppleKeyEndpoint; !strings.HasPrefix(endpoint, "https://") {
		problems = append(problems, "NUVRO_APPLE_KEY_ENDPOINT must be an https URL")
	}
	if len(authentication.CursorSecret) > 0 && len(authentication.CursorSecret) < minimumCursorSecretLength {
		problems = append(problems, "NUVRO_CURSOR_SECRET must decode to at least 32 bytes")
	}
	problems = append(problems, c.appleCredentialProblems()...)
	if key := authentication.SessionSigningKey; len(key) > 0 && len(key) < minimumCursorSecretLength {
		problems = append(problems, "NUVRO_SESSION_SIGNING_KEY must decode to at least 32 bytes")
	}
	if !c.RequiresBackingServices() {
		return problems
	}
	return append(problems, authentication.deployedProblems()...)
}

func (a Authentication) deployedProblems() []string {
	var problems []string
	if !a.UsesApple() {
		problems = append(problems, "NUVRO_APPLE_AUDIENCE is required outside development")
	}
	if len(a.CursorSecret) == 0 {
		problems = append(problems, "NUVRO_CURSOR_SECRET is required outside development")
	}
	if !a.IssuesSessions() {
		problems = append(problems, "NUVRO_SESSION_SIGNING_KEY is required outside development")
	}
	if !a.RevokesAppleTokens() {
		problems = append(problems,
			"NUVRO_APPLE_TEAM_ID, NUVRO_APPLE_KEY_ID, NUVRO_APPLE_PRIVATE_KEY and NUVRO_TOKEN_SEALING_KEY are required "+
				"outside development so Sign in with Apple tokens are revoked when an account is deleted")
	}
	return problems
}
