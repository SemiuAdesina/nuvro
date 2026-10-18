# Security standards

The reporting process, secret handling policy and personal data rules live in
[SECURITY.md](../../SECURITY.md) at the repository root, so they are visible to anyone
reporting a vulnerability. This document holds the engineering detail.

## Configuration and secrets

Configuration is typed, read from the environment, and validated at process startup. The
process fails fast with a message naming every missing or malformed variable. It never starts
with an insecure default.

`.env.example` lists variable names with empty or non-secret values. Real environment files
are ignored by Git. Provider configuration is grouped by provider.

```
NUVRO_ENV              NUVRO_HTTP_PORT        NUVRO_DATABASE_URL
NUVRO_REDIS_URL        NUVRO_LOG_LEVEL        OPENAI_API_KEY
```

`NUVRO_DATABASE_URL` and `NUVRO_REDIS_URL` are optional in development, where the health
server runs without backing services, and required in staging and production. This is
enforced in `internal/platform/config`, not by convention.

## Error exposure

A client receives a stable code, a safe message and a request identifier. Nothing else.

The `apperror` model deliberately returns a generic message for internal errors, so a wrapped
provider or database error cannot leak through the envelope by accident. That behaviour is
tested, including a test that asserts a database error string never reaches the response.

Never exposed to a client:

```
SQL   stack traces   credentials   API keys
provider errors      internal filesystem paths
```

## Logging

Structured `log/slog` records carry a request identifier, method, route, status and duration.
The request identifier is issued by the server. An identifier supplied by a caller is not
trusted and is not echoed back, so a caller cannot poison correlation.

Never logged: authentication credentials, bearer or refresh tokens, session identifiers,
passport or document data, full payment instruments, provider API keys, or raw provider
payloads containing personal data.

## HTTP server

Read header, read, write and idle timeouts and a maximum header size are set from
configuration and validated as positive. The server is never left unbounded.

Panics in request flow are recovered by middleware, logged with a stack trace on the server
side, and returned to the client as a generic internal error with no panic value or stack.
`http.ErrAbortHandler` is re-panicked so that the standard library can handle it as intended.

## Client address and trusted proxies

Rate limiting at the door keys unauthenticated callers by client address, so the address has
to be one an attacker cannot choose. NUVRO reads it from the connection unless the deployment
is explicitly told otherwise.

| Variable | Default | Meaning |
| --- | --- | --- |
| `NUVRO_TRUSTED_PROXY_HOPS` | `0` | How many trusted proxies in front of NUVRO append to `X-Forwarded-For`. |
| `NUVRO_TRUSTED_PROXY_CIDRS` | empty | Comma separated CIDR ranges of those proxies. |

With the defaults, `X-Forwarded-For` is ignored entirely and the connection peer is used, so a
direct deployment is safe with no configuration at all. Forwarding is honoured only when both
settings are present and the immediate connection peer falls inside one of the trusted ranges;
a header from any other peer is discarded. A hop count above zero without a trusted range is
refused at startup rather than silently trusting nobody.

The client is taken by walking the chain from the right by the configured hop count, never by
reading the leftmost entry. A caller who prepends addresses of their own only lengthens the
part of the chain that is skipped. Chains that are shorter than the hop count, contain an empty
entry, carry more than 32 entries or exceed 1024 bytes are treated as unusable, and the peer
address is used instead. That fails towards over-counting a shared proxy, never towards letting
a caller pick their own bucket.

Two deployments are supported. A single managed load balancer sets hops to `1` and the load
balancer subnet as the trusted range. A CDN in front of a load balancer sets hops to `2` and
both ranges. The PROXY protocol is deliberately not implemented; it belongs here only if the
infrastructure NUVRO actually runs on requires it.

## Outbound calls

Provider clients set explicit timeouts, honour context cancellation, reuse connections and
bound response body size. Retries apply only to idempotent operations, or to operations
protected by an idempotency key. Booking creation is never retried automatically.

## AI

Model providers are adapters. Model output is requested as structured output, then validated,
bounded, parsed and converted into domain types before it reaches application logic. Model
output is never trusted as domain state, and never used as the source of truth for a live
travel fact.

## Dependency and vulnerability scanning

```
go vet ./...
go test -race ./...
govulncheck ./...
golangci-lint run
```

`gosec` runs inside the lint pipeline. `govulncheck` runs on every push and pull request and
weekly on a schedule. A committed secret scan runs over full history. Dependabot proposes
updates for Go modules, GitHub Actions, Swift packages and container base images.

A vulnerability with a fix available is upgraded before new feature work in the affected area.
