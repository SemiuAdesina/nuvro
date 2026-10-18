package cursor_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/cursor"
)

func codec(t *testing.T) cursor.Codec {
	t.Helper()
	value, err := cursor.New([]byte(strings.Repeat("k", 32)))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestRoundTripPreservesPosition(t *testing.T) {
	position := cursor.Position{CreatedAt: time.Date(2026, 9, 9, 10, 30, 0, 123456789, time.UTC), ID: "trip-1"}
	token, err := codec(t).Encode("trips", "user-1", position)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := codec(t).Decode("trips", "user-1", token)
	if err != nil || !decoded.CreatedAt.Equal(position.CreatedAt) || decoded.ID != position.ID {
		t.Fatalf("%+v %v", decoded, err)
	}
	if strings.Contains(token, "trip-1") {
		t.Fatal("cursor exposed the underlying identifier verbatim")
	}
}

func TestCursorIsBoundToScopeAndPrincipal(t *testing.T) {
	token, err := codec(t).Encode("trips", "user-1", cursor.Position{CreatedAt: time.Now().UTC(), ID: "trip-1"})
	if err != nil {
		t.Fatal(err)
	}
	for name, attempt := range map[string][2]string{
		"another traveller": {"trips", "user-2"},
		"another listing":   {"saved", "user-1"},
	} {
		if _, err := codec(t).Decode(attempt[0], attempt[1], token); !errors.Is(err, cursor.ErrInvalid) {
			t.Fatalf("%s replayed the cursor", name)
		}
	}
	other, secretErr := cursor.New([]byte(strings.Repeat("z", 32)))
	if secretErr != nil {
		t.Fatal(secretErr)
	}
	if _, err := other.Decode("trips", "user-1", token); !errors.Is(err, cursor.ErrInvalid) {
		t.Fatal("cursor accepted under a different secret")
	}
}

func TestMalformedCursorsAreRejected(t *testing.T) {
	subject := codec(t)
	for _, token := range []string{"", ".", "abc", "abc.def", "!!!.!!!", strings.Repeat("a", 600)} {
		if _, err := subject.Decode("trips", "user-1", token); err == nil {
			t.Fatalf("accepted %q", token)
		}
	}
	if _, err := cursor.New([]byte("short")); !errors.Is(err, cursor.ErrWeakSecret) {
		t.Fatal("weak secret accepted")
	}
	if _, err := subject.Encode("", "user", cursor.Position{ID: "x"}); err == nil {
		t.Fatal("unscoped cursor issued")
	}
	ephemeral, err := cursor.NewEphemeral()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ephemeral.Encode("trips", "user", cursor.Position{CreatedAt: time.Now(), ID: "trip"}); err != nil {
		t.Fatal(err)
	}
}

func FuzzDecode(f *testing.F) {
	f.Add("abc.def")
	f.Add("")
	f.Fuzz(func(t *testing.T, token string) {
		subject, err := cursor.New([]byte(strings.Repeat("k", 32)))
		if err != nil {
			t.Fatal(err)
		}
		position, err := subject.Decode("trips", "user-1", token)
		if err == nil && position.ID == "" {
			t.Fatal("accepted a cursor without a position")
		}
	})
}
