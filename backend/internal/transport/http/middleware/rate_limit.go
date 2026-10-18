package middleware

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/SemiuAdesina/nuvro/backend/internal/platform/admission"
	"github.com/SemiuAdesina/nuvro/backend/internal/platform/clientip"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/httperror"
)

type Principals func(*http.Request) (string, bool)

func RateLimit(limiter admission.Limiter, class admission.Class, key func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !admit(w, r, limiter, class, key(r)) {
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Admit(limiter admission.Limiter, principals Principals) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			subject, identified := principals(r)
			if identified && !admit(w, r, limiter, ClassOf(r.Method, routePattern(r)), subject) {
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Guard(limiter admission.Limiter, addresses *clientip.Resolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			class := admission.Authentication
			if ClassOf(r.Method, routePattern(r)) == admission.Exempt {
				class = admission.Exempt
			}
			if !admit(w, r, limiter, class, ClientAddress(addresses, r)) {
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func admit(w http.ResponseWriter, r *http.Request, limiter admission.Limiter, class admission.Class, subject string) bool {
	if limiter == nil || class == admission.Exempt {
		return true
	}
	allowed, err := limiter.Allow(r.Context(), class, subject)
	if err != nil {
		httperror.WriteStatus(w, http.StatusServiceUnavailable, "upstream_unavailable", "Request admission is unavailable.")
		return false
	}
	if !allowed {
		httperror.WriteStatus(w, http.StatusTooManyRequests, "rate_limited", "Request limit reached.")
		return false
	}
	return true
}

var expensiveSuffixes = []string{
	"/search", "/generate", "/replan", "/translate", "/compare", "/matches", "/offline-pack",
}

func ClassOf(method, pattern string) admission.Class {
	switch {
	case strings.HasPrefix(pattern, "/health"):
		return admission.Exempt
	case strings.HasPrefix(pattern, "/v1/auth/"):
		return admission.Authentication
	case strings.HasPrefix(pattern, "/v1/commerce/"), strings.HasPrefix(pattern, "/v1/me/price-watches"):
		return commerceClass(method)
	case isExpensive(pattern):
		return admission.ExpensiveSearch
	default:
		return admission.StandardRead
	}
}

func commerceClass(method string) admission.Class {
	if method == http.MethodGet {
		return admission.StandardRead
	}
	return admission.CommerceWrite
}

func isExpensive(pattern string) bool {
	for _, suffix := range expensiveSuffixes {
		if strings.HasSuffix(pattern, suffix) {
			return true
		}
	}
	return false
}

func routePattern(r *http.Request) string {
	if context := chi.RouteContext(r.Context()); context != nil && context.RoutePattern() != "" {
		return context.RoutePattern()
	}
	return r.URL.Path
}

func ClientAddress(addresses *clientip.Resolver, r *http.Request) string {
	return addresses.Resolve(r.RemoteAddr, r.Header.Values(clientip.ForwardedHeader))
}
