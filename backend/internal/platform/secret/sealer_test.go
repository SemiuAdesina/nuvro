package secret_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/secret"
)

const distinguishableLength = 8

func key(fill byte) []byte { return bytes.Repeat([]byte{fill}, secret.KeyLength) }

func sealer(t *testing.T, fill byte) *secret.Sealer {
	t.Helper()
	value, err := secret.NewSealer(key(fill))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestASealedValueOpensBackToItself(t *testing.T) {
	sealed, err := sealer(t, 'k').Seal("apple-refresh-token", []byte("user-1"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(sealed, []byte("apple-refresh-token")) {
		t.Fatal("the sealed value still contains the plaintext")
	}
	opened, err := sealer(t, 'k').Open(sealed, []byte("user-1"))
	if err != nil {
		t.Fatal(err)
	}
	if opened != "apple-refresh-token" {
		t.Fatalf("opened = %q", opened)
	}
}

func TestSealingTheSameValueTwiceProducesDifferentBytes(t *testing.T) {
	first, err := sealer(t, 'k').Seal("token", nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := sealer(t, 'k').Seal("token", nil)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(first, second) {
		t.Fatal("sealing is deterministic, so equal secrets are visible as equal ciphertexts")
	}
}

func TestAValueSealedForOneAccountWillNotOpenForAnother(t *testing.T) {
	sealed, err := sealer(t, 'k').Seal("token", []byte("user-1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sealer(t, 'k').Open(sealed, []byte("user-2")); !errors.Is(err, secret.ErrCorrupt) {
		t.Fatalf("err = %v, want a refusal to open another account's secret", err)
	}
}

func TestADifferentKeyCannotOpenTheValue(t *testing.T) {
	sealed, err := sealer(t, 'k').Seal("token", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sealer(t, 'j').Open(sealed, nil); !errors.Is(err, secret.ErrCorrupt) {
		t.Fatalf("err = %v, want a refusal", err)
	}
}

func TestATamperedValueIsRefused(t *testing.T) {
	sealed, err := sealer(t, 'k').Seal("token", nil)
	if err != nil {
		t.Fatal(err)
	}
	tampered := bytes.Clone(sealed)
	tampered[len(tampered)-1] ^= 0xff
	if _, err := sealer(t, 'k').Open(tampered, nil); !errors.Is(err, secret.ErrCorrupt) {
		t.Fatalf("err = %v, want a refusal", err)
	}
	if _, err := sealer(t, 'k').Open(sealed[:4], nil); !errors.Is(err, secret.ErrCorrupt) {
		t.Fatalf("err = %v, want a refusal of a truncated value", err)
	}
}

func TestAKeyOfTheWrongLengthIsRefused(t *testing.T) {
	for _, length := range []int{0, 16, 31, 33} {
		if _, err := secret.NewSealer(bytes.Repeat([]byte{'k'}, length)); !errors.Is(err, secret.ErrNoKey) {
			t.Fatalf("a %d byte key was accepted", length)
		}
	}
}

func TestAnAbsentSealerRefusesRatherThanStoringPlaintext(t *testing.T) {
	var absent *secret.Sealer
	if _, err := absent.Seal("token", nil); !errors.Is(err, secret.ErrNoKey) {
		t.Fatalf("err = %v, want a refusal to seal with no key", err)
	}
	if _, err := absent.Open([]byte("anything"), nil); !errors.Is(err, secret.ErrNoKey) {
		t.Fatalf("err = %v, want a refusal to open with no key", err)
	}
}

func FuzzSealedValuesNeverOpenWithoutTheKey(f *testing.F) {
	f.Add("token", "user-1")
	f.Add("", "")
	f.Add(strings.Repeat("a", 4096), "user-2")
	sealing, err := secret.NewSealer(bytes.Repeat([]byte{'k'}, secret.KeyLength))
	if err != nil {
		f.Fatal(err)
	}
	other, err := secret.NewSealer(bytes.Repeat([]byte{'j'}, secret.KeyLength))
	if err != nil {
		f.Fatal(err)
	}
	f.Fuzz(func(t *testing.T, plaintext, context string) {
		sealed, sealErr := sealing.Seal(plaintext, []byte(context))
		if sealErr != nil {
			t.Fatal(sealErr)
		}
		if len(plaintext) >= distinguishableLength && bytes.Contains(sealed, []byte(plaintext)) {
			t.Fatal("the plaintext survives in the sealed value")
		}
		if _, err := other.Open(sealed, []byte(context)); !errors.Is(err, secret.ErrCorrupt) {
			t.Fatalf("a foreign key opened the value: %v", err)
		}
		opened, err := sealing.Open(sealed, []byte(context))
		if err != nil || opened != plaintext {
			t.Fatalf("round trip failed: %q %v", opened, err)
		}
	})
}
