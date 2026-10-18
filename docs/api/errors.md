# Errors

All errors use RFC 9457 `application/problem+json`, without a success/error envelope.
The controlled schema contains `type`, `title`, `status`, `detail`, optional
`instance`, and extensions `code`, `requestId`, `errors`. The runtime currently
uses `about:blank`; title matches the HTTP status. `instance` and field errors are
optional and are not populated with raw request values or validator diagnostics.

Stable codes: validation_error, authentication_required, forbidden, not_found,
conflict, rate_limited, upstream_unavailable, provider_timeout, invalid_state,
idempotency_conflict, internal_error. Transport also defines method_not_allowed,
payload_too_large and unsupported_media_type.

Generated strict request/response error hooks and OpenAPI validator failures all
use this mapping. SQL, provider failures, panic text, tokens and paths never become
response detail. Internal causes remain wrapped for server diagnosis. Readiness
failure returns a safe 503 problem, not raw dependency errors.

Every response carries `X-Request-ID`. Phase 0's stronger rule is retained: caller
request IDs are never accepted or echoed. A failed entropy source uses a process
sequence/time fallback so correlation is not omitted. Request IDs are not secrets,
credentials or authorization inputs.

W3C traceparent/tracestate is propagated. No baggage or credentials are echoed.
Logs carry request_id, trace_id, method, matched route template, status and duration.
Unmatched paths are labeled `unmatched`; arbitrary URLs do not become metric labels.
