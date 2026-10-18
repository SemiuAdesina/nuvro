package postgres

import (
	"testing"
	"time"
)

func TestIdentifierAndDateMapping(t *testing.T) {
	if _, err := identifier("bad"); err == nil {
		t.Fatal("invalid identifier")
	}
	if _, err := identifier("00112233445566778899aabbccddeeff"); err != nil {
		t.Fatal(err)
	}
	if datePointer(dateValue(nil)) != nil {
		t.Fatal("null date lost")
	}
	now := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
	if got := datePointer(dateValue(&now)); got == nil || !got.Equal(now) {
		t.Fatal("date lost")
	}
}

func FuzzIdentifier(f *testing.F) {
	f.Add("00112233445566778899aabbccddeeff")
	f.Add("invalid")
	f.Fuzz(func(t *testing.T, value string) {
		parsed, err := identifier(value)
		if err == nil && !parsed.Valid {
			t.Fatal("accepted invalid identifier")
		}
	})
}
