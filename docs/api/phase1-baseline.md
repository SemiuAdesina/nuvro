# Phase 1 baseline

Inspected on 2026-09-08 before edits. Root: the repository root.
Branch: `main`, unborn (no commits). Phase 0 files are staged; concurrent iOS
changes and prototype contract additions were already present. No index changes made.

Existing backend: Go 1.27.1, Chi health handlers, typed configuration, lifecycle,
worker supervision, money domain, safe error envelope, request IDs and slog.
OpenAPI 3.1 was split but not generated. sqlc configured; no migrations or queries.
No PostgreSQL/Redis connections, production authentication, or engine implementations.

`make verify` executed before editing:

| Gate | Result |
| --- | --- |
| LOC, comments, dependencies, format, vet, unit tests, race, lint | PASS |
| Swift tests | FAIL: Travel Setup compiler errors, including SetupReviewStep type checking and UserDefaults Sendable diagnostics |
| govulncheck | SKIPPED: executable not installed |
| Android | SKIPPED: Gradle project not initialized |

Full baseline output retained locally at `/tmp/nuvro-phase1-baseline-preserved.log`.
These failures are separate from any Phase 1 regressions.

## Continuation baseline, 2026-09-09

Inspected before completing the phase. The earlier engineer's work was present and
uncommitted; the index still held Phase 0. `make verify` before any edit:

| Gate | Result |
| --- | --- |
| LOC, comments, dependencies, api-validate, api-check, gofmt, vet, tests, race, lint, govulncheck, Swift tests | PASS |
| sql-check | SKIPPED by the script because sqlc was not on `PATH`; PASS when run with `SQLC=/tmp/nuvro-tools/sqlc` |
| Android | SKIPPED: Gradle project not initialised |

Integration tests passed against the disposable `nuvro-phase1-postgres` and
`nuvro-phase1-redis` containers. The Swift failures recorded above were fixed by the
concurrent iOS work before this continuation. One defect was found in the earlier work:
entity identifiers were generated as 32 hex characters and read back as hyphenated UUIDs,
so a created trip's identifier did not equal the stored one. `id.NewUUID` fixes that and
the integration test now asserts the round trip.

## Protected slice released, 2026-09-09

The six protected operations were promoted from drafts to `openapi/v1`, and the drafts root
was removed. `PATCH /v1/me` and `GET /v1/me/preferences` were added so profile and
preferences are symmetrical, giving ten released operations.

Identity is implemented rather than assumed. Sign in with Apple tokens are verified against
Apple's published keys, first sign-in enrols an account atomically, and a concurrent
enrolment recovers through the identity unique constraint. A development credential is
available only when explicitly configured in development, enforced twice.

The generator now runs with `nullable-type`, so PATCH bodies distinguish absent, null and
value. That mapping is carried into the domain by `change.Change` and covered at three
levels. Trip listing is keyset paginated with HMAC-signed cursors bound to the listing scope
and the principal.

Gates after the change: every gate in `make verify` passes with only Android skipped;
integration passes under the race detector against disposable PostgreSQL and Redis; all five
fuzz targets run clean. Nothing is committed.
