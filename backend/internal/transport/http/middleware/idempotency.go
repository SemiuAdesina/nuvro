package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/replay"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/httperror"
)

const IdempotencyKeyHeader = "Idempotency-Key"
const IdempotencyRetention = 24 * time.Hour
const maximumIdempotencyKeyLength = 255
const maximumReplayBodyBytes = 256 << 10

type idempotencyKey struct{}

func ParseIdempotencyKey(raw string) (string, bool) {
	if len(raw) == 0 || len(raw) > maximumIdempotencyKeyLength {
		return "", false
	}
	for index := 0; index < len(raw); index++ {
		if !isIdempotencyKeyByte(raw[index]) {
			return "", false
		}
	}
	return raw, true
}

func isIdempotencyKeyByte(value byte) bool {
	switch {
	case value >= 'a' && value <= 'z', value >= 'A' && value <= 'Z', value >= '0' && value <= '9':
		return true
	case value == '.', value == '_', value == ':', value == '-':
		return true
	default:
		return false
	}
}

func IdempotencyKeyFrom(ctx context.Context) (string, bool) {
	key, ok := ctx.Value(idempotencyKey{}).(string)
	return key, ok && key != ""
}

var replayedRoutes = map[string]bool{
	"POST /v1/trips":             true,
	"POST /v1/commerce/handoffs": true,
	"POST /v1/me/price-watches":  true,
}

func RequiresIdempotency(method, pattern string) bool {
	return replayedRoutes[method+" "+pattern]
}

func Idempotent(store replay.Store, principals Principals) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !RequiresIdempotency(r.Method, routePattern(r)) {
				next.ServeHTTP(w, r)
				return
			}
			key, ok := ParseIdempotencyKey(r.Header.Get(IdempotencyKeyHeader))
			if !ok {
				httperror.WriteStatus(w, http.StatusBadRequest, "validation_error",
					"A valid Idempotency-Key header is required for this request.")
				return
			}
			subject, identified := principals(r)
			request := r.WithContext(context.WithValue(r.Context(), idempotencyKey{}, key))
			if store == nil || !identified {
				next.ServeHTTP(w, request)
				return
			}
			body, err := drain(request)
			if err != nil {
				httperror.WriteStatus(w, http.StatusBadRequest, "validation_error", "Request is invalid.")
				return
			}
			serveOnce(w, request, next, store, subject, key, fingerprintOf(request, body))
		})
	}
}

func serveOnce(w http.ResponseWriter, r *http.Request, next http.Handler,
	store replay.Store, subject, key, fingerprint string) {
	if recorded, found := store.Recorded(r.Context(), subject, key); found {
		if recorded.Fingerprint != fingerprint {
			httperror.WriteStatus(w, http.StatusConflict, "conflict",
				"This idempotency key was already used for a different request.")
			return
		}
		write(w, recorded)
		return
	}
	capture := &recorder{ResponseWriter: w, status: http.StatusOK}
	next.ServeHTTP(capture, r)
	if capture.status >= 500 || capture.overflowed {
		return
	}
	store.Record(r.Context(), subject, key, replay.Response{
		Status: capture.status, ContentType: w.Header().Get("Content-Type"),
		Body: capture.body.Bytes(), Fingerprint: fingerprint,
	}, IdempotencyRetention)
}

func write(w http.ResponseWriter, recorded replay.Response) {
	if recorded.ContentType != "" {
		w.Header().Set("Content-Type", recorded.ContentType)
	}
	w.Header().Set("Idempotent-Replay", "true")
	w.WriteHeader(recorded.Status)
	_, _ = w.Write(recorded.Body)
}

func drain(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

func fingerprintOf(r *http.Request, body []byte) string {
	sum := sha256.New()
	sum.Write([]byte(r.Method))
	sum.Write([]byte{0})
	sum.Write([]byte(r.URL.Path))
	sum.Write([]byte{0})
	sum.Write(body)
	return hex.EncodeToString(sum.Sum(nil)[:16])
}

type recorder struct {
	http.ResponseWriter
	status     int
	body       bytes.Buffer
	written    bool
	overflowed bool
}

func (r *recorder) WriteHeader(status int) {
	if r.written {
		return
	}
	r.status = status
	r.written = true
	r.ResponseWriter.WriteHeader(status)
}

func (r *recorder) Write(payload []byte) (int, error) {
	if !r.written {
		r.WriteHeader(http.StatusOK)
	}
	if r.body.Len()+len(payload) > maximumReplayBodyBytes {
		r.overflowed = true
	} else {
		r.body.Write(payload)
	}
	return r.ResponseWriter.Write(payload)
}
