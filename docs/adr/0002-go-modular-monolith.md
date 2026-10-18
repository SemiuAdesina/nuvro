# 0002. Go modular monolith

Date: 2026-09-07

## Status

Accepted

## Context

NUVRO has fourteen identified capabilities: understanding, matching, recommendation, destination
intelligence, itinerary, routing, replanning, weather, places, language, entry requirements,
budget, commerce and personalization.

The tempting reading is fourteen services. At this stage that would produce fourteen
deployment targets, fourteen failure modes and a distributed system before a single traveller
has used the product, with no operational evidence that any of it is needed.

The backend language options were Go, Node with NestJS, and a JVM stack.

## Decision

The backend is a Go modular monolith.

Go, `net/http` and Chi for routing and HTTP composition. No heavyweight web framework.
PostgreSQL with pgx/v5 and sqlc. Redis for cache and coordination. `log/slog` for structured
logging. Constructor dependency injection with no DI framework. Standard library first.

The fourteen engines are Go package boundaries under `backend/internal/engine`, not services.
Two binaries share one module: `cmd/api` and `cmd/worker`.

Node and NestJS are rejected for the backend. There is no Node in the backend.

## Consequences

Boundaries are enforced by a dependency direction gate rather than by network calls, which is
faster to develop against and far cheaper to operate. A refactor across two engines is a
compile-time change, not a coordinated multi-service deployment.

Package discipline becomes essential, because nothing physically prevents a bad import. That
is why `scripts/check-go-deps.sh` runs in CI and why the forbidden edges are written down.

Extraction remains available. An engine already has a package boundary and consumer-declared
interfaces, so extracting one later is a mechanical change. It happens when operational
evidence demands it: an isolation, scaling or failure-domain problem the monolith cannot
solve. Not before.
