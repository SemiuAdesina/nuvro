package id

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

const byteLength = 16

type Generator interface {
	Generate() (string, error)
}

type Random struct {
	source func([]byte) (int, error)
}

func NewRandom() Random {
	return Random{source: rand.Read}
}

func NewRandomFrom(source func([]byte) (int, error)) Random {
	return Random{source: source}
}

func (r Random) Generate() (string, error) {
	raw := make([]byte, byteLength)
	if _, err := r.source(raw); err != nil {
		return "", fmt.Errorf("id: read random bytes: %w", err)
	}
	return hex.EncodeToString(raw), nil
}
