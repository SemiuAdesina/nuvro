package middleware

import (
	"io"
	"net/http"

	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/httperror"
)

func BodyLimit(limit int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > limit {
				httperror.WriteStatus(w, 413, "payload_too_large", "Request body is too large.")
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}

func NoBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 1))
		if err != nil || len(body) != 0 {
			httperror.WriteStatus(w, 400, "validation_error", "This operation does not accept a body.")
			return
		}
		next.ServeHTTP(w, r)
	})
}
