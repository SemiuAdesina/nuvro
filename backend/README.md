# NUVRO backend

Go modular monolith. Two binaries, one module, strong package boundaries.

## Running

```bash
go run ./cmd/api
```

```bash
curl -i localhost:8080/health/live
```

```bash
go run ./cmd/worker
```

Both read configuration from the environment. `.env.example` at the repository root lists
every variable. In development, `NUVRO_DATABASE_URL` and `NUVRO_REDIS_URL` are optional, so
the API runs with no backing services. In staging and production they are required, and the
process refuses to start without them.

Local PostgreSQL and Redis:

```bash
docker compose -f ../infra/local/compose.yaml up -d
```

## Testing

```bash
go test ./...
```

```bash
go test -race ./...
```

```bash
go vet ./...
```

From the repository root, `make verify` runs these along with the line count, comment and
dependency direction gates.

## What exists

| Area | State |
| --- | --- |
| `cmd/api`, `cmd/worker` | Thin entrypoints. Configuration, construction, signal handling. |
| `internal/app` | Composition root. HTTP server lifecycle and graceful shutdown. |
| `internal/apperror` | Stable codes, safe messages, internal cause preserved but never exposed. |
| `internal/domain/money` | Integer minor units, ISO 4217 currency, conversion provenance. |
| `internal/platform` | Typed configuration, `log/slog` construction, injectable clock, hex request identifiers and canonical UUID entity identifiers, OpenTelemetry runtime. |
| `internal/transport/http` | Generated strict Chi server, OpenAPI request validation, Problem Details, request identity, W3C trace propagation, body limits, idempotency-key and rate-admission boundaries, fail-closed bearer authentication hook. |
| `internal/transport/http/handler/health` | `/health/live` and `/health/ready`. |
| `internal/usecase/account` | Verified identity to internal account, with atomic first-sign-in enrolment. |
| `internal/usecase/profile` | Profile and preference reads and patches inside one transaction. |
| `internal/usecase/discover` | Orchestrates E01, the destination catalog, E15, E02 and E03 with bounded concurrency. |
| `internal/usecase/plan` | Atomic trip creation with owner membership, membership-checked reads, keyset listing with signed cursors, role-checked updates. |
| `internal/usecase/system` | Readiness orchestration. |
| `internal/adapter/external/apple` | Sign in with Apple identity token verification with bounded, cached key retrieval. |
| `internal/adapter/postgres`, `internal/adapter/redis` | pgx pool and sqlc-backed store with a use-case-owned transaction boundary; go-redis client, probe and namespaced key formatter. |
| `internal/domain/{trip,user,destination,evidence,change}` | Trip lifecycle, roles and patches, principal, profile, preferences and enrolment, Match Score and media models, freshness, and the absent/null/value change vocabulary. |
| `internal/engine/{understanding,feasibility,matching,recommendation}` | E01 intent, E15 access and affordability, E02 deterministic Match Score, E03 ranking and explanation. |
| `internal/engine/{media,requirements}` | Consumer-owned ports for S01 candidates and E11 evidence. |
| `internal/worker/runner` | Cancellation-aware job supervision. Zero jobs registered. |
| `tests/contract` | Binds `contracts/openapi` to the implementation. |

Everything else in the tree is a reserved boundary documented in
`docs/architecture/repository-tree.md`. No placeholder implementations exist.

## Layers

```
domain  ←  engine  ←  usecase  ←  transport
```

Adapters implement interfaces declared by their consumers. `internal/app` is the only place
that wires a concrete implementation to an interface. `internal/apperror` and
`internal/platform` are cross-cutting and may be imported anywhere.

The forbidden edges are enforced:

```bash
../scripts/check-go-deps.sh
```

## Health endpoints

`/health/live` reports that the process is running. It never consults backing services, so a
dependency outage does not get the process restarted.

`/health/ready` consults every registered component probe and returns 503 when any is down.
Failure detail is reduced to `up` or `down`, so the probe cannot leak a connection string or
a credential. Configured PostgreSQL and Redis probes are registered, with a shared two-second deadline. Failure returns safe RFC 9457 Problem Details.

## Database

PostgreSQL through pgx/v5 with SQL compiled by sqlc. No ORM. Migrations are Goose SQL under
`db/migrations`, queries under `db/queries`, and generated Go lands in
`internal/adapter/postgres/sqlc` where it is never edited by hand.

`sqlc.yaml` generates the foundational queries using pinned sqlc 1.31.1. Run `make sql-generate` from the root; point `SQLC` at the binary when it is not on `PATH`. Entity identifiers are canonical RFC 4122 version 4 strings from `id.NewUUID`, so a value written by a use case reads back byte-for-byte equal. Goose applies the seven-table foundation with `make db-migrate`. See `docs/architecture/phase1-foundation.md`.

## Adding an endpoint

1. Add domain vocabulary and invariants with tests.
2. Implement the capability use case with authorization, a transaction boundary and tests.
3. Add queries, run `make sql-generate`, and cover the repository with real database tests.
4. Add the path and schema under `contracts/openapi/v1`, then `make api-generate`.
5. Implement the generated strict interface with explicit DTO mappers.
6. Wire it in `internal/app` and add contract tests that exercise the router.
7. Run `make verify` and `make backend-integration` against disposable services.

Generated API code lives in `internal/transport/http/api/server.gen.go`. sqlc output
lives in `internal/adapter/postgres/sqlc`. Neither is edited manually. Domain types
import neither. Build tools are pinned in go.mod (oapi-codegen) and scripts (sqlc).
The Goose command imports only the required PostgreSQL driver.

See [API docs](../docs/api/overview.md) for commands, draft promotion, errors and
mobile generation. OTel is initialized for local instrumentation; exporter/collector
selection is still a deployment decision. No external travel provider is integrated.
