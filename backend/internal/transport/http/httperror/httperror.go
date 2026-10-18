package httperror

import (
	"encoding/json"
	"net/http"

	"github.com/SemiuAdesina/nuvro/backend/internal/apperror"
	"github.com/SemiuAdesina/nuvro/backend/internal/transport/http/api"
)

const RequestIDHeader = "X-Request-ID"

func Status(code apperror.Code) int {
	switch code {
	case apperror.CodeInvalidInput:
		return http.StatusBadRequest
	case apperror.CodeUnauthenticated:
		return http.StatusUnauthorized
	case apperror.CodeForbidden:
		return http.StatusForbidden
	case apperror.CodeNotFound:
		return http.StatusNotFound
	case apperror.CodeConflict, apperror.CodeInvalidState, apperror.CodeIdempotencyConflict:
		return http.StatusConflict
	case apperror.CodeRateLimited:
		return http.StatusTooManyRequests
	case apperror.CodeUnavailable:
		return http.StatusServiceUnavailable
	case apperror.CodeProviderTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

func Write(w http.ResponseWriter, err error) {
	code := apperror.CodeOf(err)
	WriteStatus(w, Status(code), string(code), apperror.SafeMessage(err))
}

func WriteStatus(w http.ResponseWriter, status int, code, detail string) {
	response := api.Problem{
		Type: "about:blank", Title: http.StatusText(status), Status: status,
		Detail: detail, Code: api.ProblemCode(code),
	}
	if requestID := w.Header().Get(RequestIDHeader); requestID != "" {
		response.RequestId = &requestID
	}
	w.Header().Set("Content-Type", "application/problem+json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}

func NotFoundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		Write(w, apperror.New(apperror.CodeNotFound, "The requested resource does not exist."))
	}
}

func MethodNotAllowedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		WriteStatus(w, http.StatusMethodNotAllowed, "method_not_allowed", "The requested method is not supported.")
	}
}
