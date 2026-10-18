package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

const KeyLength = 32

var ErrNoKey = errors.New("secret: a 32 byte key is required")
var ErrCorrupt = errors.New("secret: the sealed value cannot be opened")

type Sealer struct {
	aead cipher.AEAD
}

func NewSealer(key []byte) (*Sealer, error) {
	if len(key) != KeyLength {
		return nil, ErrNoKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrNoKey
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrNoKey
	}
	return &Sealer{aead: aead}, nil
}

func (s *Sealer) Seal(plaintext string, context []byte) ([]byte, error) {
	if s == nil {
		return nil, ErrNoKey
	}
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return s.aead.Seal(nonce, nonce, []byte(plaintext), context), nil
}

func (s *Sealer) Open(sealed []byte, context []byte) (string, error) {
	if s == nil {
		return "", ErrNoKey
	}
	size := s.aead.NonceSize()
	if len(sealed) < size {
		return "", ErrCorrupt
	}
	plaintext, err := s.aead.Open(nil, sealed[:size], sealed[size:], context)
	if err != nil {
		return "", ErrCorrupt
	}
	return string(plaintext), nil
}
