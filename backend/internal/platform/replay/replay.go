package replay

import (
	"context"
	"time"
)

type Response struct {
	Status      int               `json:"status"`
	ContentType string            `json:"contentType"`
	Body        []byte            `json:"body"`
	Fingerprint string            `json:"fingerprint"`
	Headers     map[string]string `json:"headers,omitempty"`
}

type Store interface {
	Recorded(ctx context.Context, subject, key string) (Response, bool)
	Record(ctx context.Context, subject, key string, recorded Response, ttl time.Duration)
}
