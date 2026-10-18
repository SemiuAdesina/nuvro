# API platform

The released source is `contracts/openapi/v1/openapi.yaml`, OpenAPI 3.1.0. It carries ten
operations: two unauthenticated health probes and eight authenticated product operations for
profile, preferences and trips.

```
OpenAPI → generated strict Chi interface → transport mapper → use case
        → domain / engine → consumer ports → PostgreSQL / Redis / providers
```

`make api-generate` validates and bundles local references, then runs pinned `oapi-codegen`
2.8.0 with `nullable-type` so PATCH bodies distinguish absent, null and value. The generated
server, models and embedded specification live in
`backend/internal/transport/http/api/server.gen.go`. `make api-check` regenerates into a
temporary directory and compares, without touching the working tree or the index.
`make sql-generate` and `make sql-check` require sqlc 1.31.1; pass `SQLC=<path>` when it is
not on `PATH`.

The unversioned Phase 0 documents under `contracts/openapi/` are retained as compatibility
material for the concurrent iOS prototype. They use snake_case and the old error envelope,
they are not generation inputs, and they do not describe the released API. There is no
`drafts` root any more: every operation in the contract is implemented and served.

The runtime uses generated DTOs at the edge only. Domain models and sqlc rows are separate
types; PostgreSQL repositories map explicitly between them, and only `internal/app` knows
which concrete adapter implements a port.

Liveness performs no dependency I/O. Readiness probes the configured PostgreSQL and Redis
with a shared two-second deadline. With no dependencies configured, development readiness
means the health-only process can serve its released surface; staging and production require
both, and also require a configured Apple audience and cursor secret.
