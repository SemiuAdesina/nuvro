package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clock"
)

type recordingWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *recordingWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *recordingWriter) Write(payload []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	written, err := w.ResponseWriter.Write(payload)
	w.bytes += written
	return written, err
}

func AccessLog(logger *slog.Logger, timer clock.Clock) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startedAt := timer.Now()
			recorder := &recordingWriter{ResponseWriter: w}

			next.ServeHTTP(recorder, r)

			if recorder.status == 0 {
				recorder.status = http.StatusOK
			}
			logger.InfoContext(r.Context(), "http request",
				slog.String("request_id", RequestIDFrom(r.Context())),
				slog.String("method", r.Method),
				slog.String("route", Route(r)),
				slog.String("trace_id", trace.SpanContextFromContext(r.Context()).TraceID().String()),
				slog.Int("status", recorder.status),
				slog.Int("bytes", recorder.bytes),
				slog.Duration("duration", timer.Now().Sub(startedAt).Round(time.Microsecond)),
			)
		})
	}
}

func (w *recordingWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }
