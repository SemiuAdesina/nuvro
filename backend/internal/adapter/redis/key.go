package redis

import (
	"errors"
	"strings"
)

const keyPrefix = "nuvro:v1"
const maximumKeyLength = 512

var ErrInvalidKey = errors.New("redis: invalid key")

func Key(capability, identifier string) (string, error) {
	if !isCapabilityName(capability) || identifier == "" || !isIdentifier(identifier) {
		return "", ErrInvalidKey
	}
	key := keyPrefix + ":" + capability + ":" + identifier
	if len(key) > maximumKeyLength {
		return "", ErrInvalidKey
	}
	return key, nil
}

func isCapabilityName(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if (r < 'a' || r > 'z') && r != '_' {
			return false
		}
	}
	return true
}

func isIdentifier(value string) bool {
	for _, r := range value {
		if r <= ' ' || r > '~' || strings.ContainsRune(":*?[]", r) {
			return false
		}
	}
	return true
}
