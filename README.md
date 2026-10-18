# NUVRO

**The world, made personal.**

NUVRO is a global vacation intelligence and travel companion platform.
The product flow is **Discover → Plan → Travel**, built on a global foundation of language,
locale, currency, home country, departure airport, units, timezone, passport context,
accessibility, profile and cloud sync.

## Status

**Phase 1 — backend platform and the first product slice.**
The Go backend serves a generated OpenAPI 3.1 contract with request validation, Problem
Details, tracing, PostgreSQL and Redis readiness and bounded shutdown. Sign in with Apple
identity tokens are verified against Apple's published keys, and a first sign-in enrols an
account atomically. Profile, preferences and trips are released and authenticated, with
three-state PATCH semantics and signed cursor pagination. Goose migrations and sqlc
repositories back them. The travel engines and external providers are not implemented.
Native iOS UI work is in progress independently. See [API overview](docs/api/overview.md),
[authentication](docs/api/authentication.md) and
[Phase 1 foundation](docs/architecture/phase1-foundation.md).

## Repository layout

```
backend/     Go modular monolith: api and worker binaries
clients/ios  Native SwiftUI application and local Swift packages
clients/android  Native Kotlin and Jetpack Compose application
contracts/   OpenAPI contract shared by both clients
design/      Brand direction, onboarding art plan, design tokens
docs/        Architecture, product, standards and decision records
infra/       Local Docker Compose and backend container build
scripts/     Repository quality gates
```

Full tree, including reserved and intentionally empty directories, is in
[docs/architecture/repository-tree.md](docs/architecture/repository-tree.md).

## Quick start

```bash
make verify
```

Run the API locally:

```bash
cd backend && go run ./cmd/api
```

```bash
curl localhost:8080/health/live
```

Local PostgreSQL and Redis:

```bash
docker compose -f infra/local/compose.yaml up -d
```

## Signature capabilities

1. Vacation Match Score
2. Where Can My Money Take Me?
3. Everything About This Place
4. Adaptive Travel Mode

## Rules that will not change

- Vacation Match Score is deterministic backend logic. AI explains it, never invents it.
- Affiliate payout never influences organic Match Score or ranking.
- AI is never the source of truth for live travel facts.
- Canonical travel intelligence lives in Go, never duplicated in Swift or Kotlin, so iOS and
  Android return identical results for the same request.

The full contract is in [CLAUDE.md](CLAUDE.md).

## Documentation

- [Architecture overview](docs/architecture/overview.md)
- [Dependency rules](docs/architecture/dependency-rules.md)
- [Engine registry](docs/architecture/engine-registry.md)
- [Data flow](docs/architecture/data-flow.md)
- [Provider boundaries](docs/architecture/provider-boundaries.md)
- [Code rules](docs/standards/code-rules.md)
- [Testing](docs/standards/testing.md)
- [Security](SECURITY.md)
- [Contributing](CONTRIBUTING.md)
