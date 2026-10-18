package apperror_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
)

func TestNewCarriesCodeAndMessage(t *testing.T) {
	err := apperror.New(apperror.CodeNotFound, "Destination not found.")

	if err.Code() != apperror.CodeNotFound {
		t.Errorf("Code() = %q, want not_found", err.Code())
	}
	if err.Message() != "Destination not found." {
		t.Errorf("Message() = %q", err.Message())
	}
	if err.Unwrap() != nil {
		t.Errorf("Unwrap() = %v, want nil", err.Unwrap())
	}
}

func TestWrapPreservesCauseForInternalInspection(t *testing.T) {
	cause := errors.New("pq: relation \"trips\" does not exist")
	err := apperror.Wrap(apperror.CodeInternal, "Could not load the trip.", cause)

	if !errors.Is(err, cause) {
		t.Fatal("errors.Is did not find the cause")
	}
	if !strings.Contains(err.Error(), cause.Error()) {
		t.Fatalf("Error() = %q, want it to include the cause", err.Error())
	}
}

func TestCodeOfFindsWrappedApplicationError(t *testing.T) {
	err := fmt.Errorf("loading trip: %w", apperror.New(apperror.CodeConflict, "Trip already exists."))

	if got := apperror.CodeOf(err); got != apperror.CodeConflict {
		t.Fatalf("CodeOf = %q, want conflict", got)
	}
}

func TestCodeOfTreatsUnknownErrorsAsInternal(t *testing.T) {
	if got := apperror.CodeOf(errors.New("boom")); got != apperror.CodeInternal {
		t.Fatalf("CodeOf = %q, want internal", got)
	}
	if got := apperror.CodeOf(nil); got != apperror.CodeInternal {
		t.Fatalf("CodeOf(nil) = %q, want internal", got)
	}
}

func TestSafeMessageReturnsAuthoredMessageForClientErrors(t *testing.T) {
	err := apperror.New(apperror.CodeInvalidInput, "Departure date must be in the future.")

	if got := apperror.SafeMessage(err); got != "Departure date must be in the future." {
		t.Fatalf("SafeMessage = %q", got)
	}
}

func TestSafeMessageHidesInternalDetail(t *testing.T) {
	cause := errors.New("dial tcp 10.0.0.4:5432: connection refused")
	err := apperror.Wrap(apperror.CodeInternal, "postgres pool exhausted", cause)

	message := apperror.SafeMessage(err)
	for _, leak := range []string{"postgres", "10.0.0.4", "connection refused"} {
		if strings.Contains(strings.ToLower(message), leak) {
			t.Fatalf("SafeMessage = %q, which leaks %q", message, leak)
		}
	}
	if message == "" {
		t.Fatal("SafeMessage returned an empty string")
	}
}

func TestSafeMessageHidesUnknownErrors(t *testing.T) {
	if got := apperror.SafeMessage(errors.New("pq: syntax error at or near SELECT")); strings.Contains(got, "SELECT") {
		t.Fatalf("SafeMessage = %q, which leaks the underlying error", got)
	}
}

func TestUnknownApplicationCodeHidesDetail(t *testing.T) {
	err := apperror.New(apperror.Code("unknown"), "private diagnostic")
	if apperror.CodeOf(err) != apperror.CodeInternal || strings.Contains(apperror.SafeMessage(err), "private") {
		t.Fatal("unknown error exposed detail")
	}
}
