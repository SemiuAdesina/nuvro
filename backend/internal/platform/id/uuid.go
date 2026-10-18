package id

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type UUID struct {
	source func([]byte) (int, error)
}

func NewUUID() UUID {
	return UUID{source: rand.Read}
}

func NewUUIDFrom(source func([]byte) (int, error)) UUID {
	return UUID{source: source}
}

func (u UUID) Generate() (string, error) {
	raw := make([]byte, byteLength)
	if _, err := u.source(raw); err != nil {
		return "", fmt.Errorf("id: read random bytes: %w", err)
	}
	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(raw)
	return encoded[:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:], nil
}
