package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/id"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/httperror"
)

type contextKey struct{}

var requestIDKey contextKey
var fallbackSequence atomic.Uint64

func RequestID(generator id.Generator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			value, err := generator.Generate()
			if err != nil {
				value = fmt.Sprintf("fallback-%x-%x", time.Now().UnixNano(), fallbackSequence.Add(1))
			}
			if value != "" {
				w.Header().Set(httperror.RequestIDHeader, value)
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, value)))
		})
	}
}

func RequestIDFrom(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey).(string)
	return value
}
