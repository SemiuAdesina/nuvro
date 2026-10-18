# 0004. PostgreSQL authoritative, Redis ephemeral

Date: 2026-09-07

## Status

Accepted

## Context

NUVRO mixes two very different kinds of data. Trips, travellers, preferences and bookings
must survive anything. Weather forecasts, price quotes, provider responses and rate limit
counters are worth caching and worthless once stale.

Treating both the same way produces either a slow product or a product that loses a
traveller's trip.

## Decision

PostgreSQL is the authoritative transactional datastore, accessed through pgx/v5 with
hand-written SQL compiled by sqlc. No ORM.

Redis is ephemeral. It may hold destination, weather and price caches, rate limit state,
short-lived session and trip state, distributed coordination and queue state. Critical trip or
account information is never stored only in Redis.

Every Redis entry has a namespaced key prefixed by its owning capability, a single owning
package, a TTL where the value can go stale, and a documented invalidation strategy.

Money is stored as integer minor units plus an ISO 4217 currency code, never floating point.
Instants are `timestamptz` read and written as UTC. Destination local times are stored with an
IANA timezone identifier.

## Consequences

Losing the entire Redis instance degrades performance and never correctness. That property is
worth the discipline it demands.

Hand-written SQL means schema changes are explicit and reviewable, and query performance is
visible rather than emergent. It also means more typing than an ORM, which is accepted.

Destructive schema changes use expand, migrate, contract across releases. A column is never
dropped in the same release that removes its usage, so a rollback mid-sequence still leaves a
working system.
