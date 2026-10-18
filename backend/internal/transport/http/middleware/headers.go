package middleware

import "net/http"

const strictTransport = "max-age=63072000; includeSubDomains"
const contentPolicy = "default-src 'none'; frame-ancestors 'none'"

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("Cache-Control", "no-store")
		header.Set("Strict-Transport-Security", strictTransport)
		header.Set("Content-Security-Policy", contentPolicy)
		header.Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}
