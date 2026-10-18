# API conventions

JSON uses camelCase. Resources are returned directly, never wrapped in a success envelope.
Collections return `items`, `hasMore` and an optional `nextCursor`.

## Pagination

`GET /v1/trips` is cursor paginated, ordered by `createdAt` then `id`, both descending. The
repository asks for one row more than the page size to decide `hasMore` without a second
query, and returns only the requested number.

A cursor is opaque and authenticated. `platform/cursor` encodes the position and signs it
with HMAC-SHA256 over the listing scope, the principal and the payload. A cursor issued to
one traveller is rejected for another, a cursor for one listing is rejected for a different
listing, and any tampering fails the signature. `NUVRO_CURSOR_SECRET` is base64, at least 32
bytes, and required outside development; development generates an ephemeral per-process
secret, so cursors do not survive a restart there. Offsets and database identifiers are never
exposed.

## Money, time and place

Money is signed int64 minor units plus an ISO 4217 code. Coordinates carry validated ranges.
Calendar ranges use ISO dates; local presentation keeps an IANA timezone beside absolute
timestamps. Domain validation enforces the relational invariants JSON Schema cannot express,
such as an end date not preceding a start date.

## PATCH semantics

An omitted field is unchanged. A nullable field accepts explicit null to clear it. A
non-nullable field rejects null. The generated types use three-state nullable values, the
transport maps them to a domain `change.Change`, and the domain refuses to clear a required
field. Trip dates change as a unit: both dates, or both null. Sending only one is rejected as
`validation_error` before any write. All three states are covered by unit tests in the
domain, the handler mapping and the end-to-end contract tests.

## Vocabularies and freshness

Stable NUVRO vocabularies use enums; provider-owned vocabularies do not. Dynamic facts carry
source and freshness, and AI explanation never replaces authoritative evidence. Media assets
carry mobile-ready URLs, dimensions, alt text, attribution and expiry; provider photo
references stay behind S01.

## Limits and admission

Headers are capped and a 64 KiB body ceiling applies to the current surface; a future
operation that needs a different ceiling declares it explicitly. Every server timeout is
configured and validated as positive.

Rate admission has a tested middleware boundary with standard read, expensive search,
authentication and commerce-write classes. No limiter or numeric policy is enabled yet. Key
selection belongs to transport: the verified principal first, otherwise the network peer.
Forwarded headers require a trusted-proxy policy before they are trusted.

## Idempotency

`middleware.RequireIdempotencyKey` is the format boundary for future side-effecting
operations: one to 255 characters of `A-Z a-z 0-9 . _ : -`, anything else rejected as
`validation_error`, and the key placed in the request context. No released operation declares
it yet: trip creation is a cheap, owner-scoped insert with no external side effect. The
owning use case must atomically claim principal, operation and key, retain a request
fingerprint, replay completed outcomes, reject a different payload with
`idempotency_conflict`, and define expiry for incomplete operations. GET never requires a key.

## Redis keys

Keys follow `nuvro:v1:<capability>:<identifier>` and are built only by `adapter/redis.Key`,
which accepts a lowercase capability and a printable identifier without separators or glob
characters. No cache exists yet; adding one requires an owner, TTL, freshness, fallback and
invalidation. Losing Redis must never lose authoritative data.
