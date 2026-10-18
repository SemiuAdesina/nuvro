# Deployment

## What runs

Two binaries from one image. `cmd/api` serves HTTP. `cmd/worker` runs background jobs. They
share configuration and backing services and are deployed separately so that a slow job never
competes with request handling.

`cmd/migrate` applies database migrations. It is a separate command so that migrations are an
explicit deployment step rather than something a starting API process does to itself.

```bash
docker build -f infra/docker/backend.Dockerfile --build-arg TARGET=api -t nuvro-api .
docker build -f infra/docker/backend.Dockerfile --build-arg TARGET=worker -t nuvro-worker .
```

The image is `distroless/static-debian12:nonroot` with a statically linked, stripped binary and
no shell. There is nothing in it to exec into, which is deliberate.

## Order of a release

1. Apply migrations with `cmd/migrate`. Every migration so far is additive, so the running
   version keeps working while the new one rolls out.
2. Roll the API.
3. Roll the worker.

## Replicas

The API is stateless and scales horizontally. Rate limiting and idempotency both live in Redis,
so admission decisions are shared across replicas rather than per process.

The worker may also run more than one replica, but each job runs on exactly one of them. Every
job takes a PostgreSQL advisory lock named after itself for as long as it runs; a replica that
does not get the lock stays idle and waits. This matters most for the retention sweep, where two
replicas purging the same account concurrently would race. A single-process deployment with no
lease configured simply runs everything.

## Health

`GET /health/live` reports that the process is up. `GET /health/ready` additionally checks the
database and Redis. Both are exempt from rate limiting. Point the liveness probe at the first
and the readiness probe at the second.

Shutdown is graceful: the server stops accepting connections and finishes in-flight requests
within `NUVRO_HTTP_SHUTDOWN_TIMEOUT` before the process exits.

## Behind a load balancer

Set `NUVRO_TRUSTED_PROXY_HOPS` and `NUVRO_TRUSTED_PROXY_CIDRS` to match the actual topology, or
unauthenticated rate limiting keys every request behind the load balancer to the same bucket.
See [security standards](../standards/security.md).

## Cross-origin requests

The API serves native mobile clients and sends no CORS headers at all. There is no browser
origin to allow, and a permissive policy added "just in case" would be a standing invitation.
A browser client would be a deliberate decision, with its own allowlist, not a default.

## Secrets

Secrets arrive as environment variables and are validated at startup: the sealing key must
decode to exactly 32 bytes, the Sign in with Apple key must parse as a PKCS#8 EC private key,
and the cursor secret must decode to at least 32 bytes. None of them are logged, and the access
log records the matched route rather than the request path, so identifiers in URLs stay out of
logs as well.
