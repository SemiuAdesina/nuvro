package config

import (
	"strings"

	"github.com/SemiuAdesina/nuvro/backend/internal/adapter/external/apple"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/secret"
)

func (c Config) appleCredentialProblems() []string {
	var problems []string
	settings := c.Authentication
	if len(settings.TokenSealingKey) > 0 && len(settings.TokenSealingKey) != secret.KeyLength {
		problems = append(problems, "NUVRO_TOKEN_SEALING_KEY must decode to exactly 32 bytes")
	}
	if len(settings.ApplePrivateKey) > 0 {
		if _, err := apple.ParsePrivateKey(settings.ApplePrivateKey); err != nil {
			problems = append(problems, "NUVRO_APPLE_PRIVATE_KEY must decode to a PKCS#8 EC private key")
		}
	}
	for name, value := range map[string]string{
		"NUVRO_APPLE_TOKEN_ENDPOINT":  settings.AppleTokenEndpoint,
		"NUVRO_APPLE_REVOKE_ENDPOINT": settings.AppleRevokeEndpoint,
	} {
		if !strings.HasPrefix(value, "https://") {
			problems = append(problems, name+" must be an https URL")
		}
	}
	return problems
}
