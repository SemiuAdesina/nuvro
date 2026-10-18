# Request admission

Two things stand between a request and a handler: a rate limit that decides whether the
request is allowed at all, and an idempotency record that decides whether it has already
been answered. Both are enforced in the router, not left to callers.

## Order

```
Guard (per address)  →  Validate (contract + authentication)  →  Admit (per traveller)  →  Idempotent (replay)  →  handler
```

`Guard` runs before authentication so that a flood of unauthenticated or badly-credentialled
requests is shed at the door rather than after the work of validating them. `Admit` runs
after authentication because a per-traveller budget needs to know which traveller. Health
probes pass through both untouched so a load balancer can never be throttled out of service.

## Budgets

Four classes, each an independent token bucket:

| Class | Keyed by | Covers |
| --- | --- | --- |
| `authentication` | client address | every request at the door, before identity is known |
| `standard_read` | traveller | ordinary reads and small writes |
| `expensive_search` | traveller | anything that fans out to providers or runs a planner |
| `commerce_write` | traveller | handoff creation and price watch changes |

A route is classified by its pattern, not its handler, so a new endpoint inherits a sensible
budget the moment it is registered. Classification is tested against the released route list.

The bucket lives in Redis and is evaluated by one Lua script, so refill and spend are atomic
across replicas. Time comes from the injected clock, which is why the refill behaviour is
testable rather than a matter of sleeping. Buckets expire once they would be full again, so
the keyspace does not grow with traffic.

If Redis cannot answer, the request is refused with 503 rather than admitted. Admission is a
safety control; failing open would make it decorative under exactly the conditions it exists
for. Outside development, starting without Redis is a configuration error rather than a
silent downgrade.

The client address comes from the connection, never from `X-Forwarded-For`, which a caller
can set freely. A deployment behind a trusted proxy must set the real address on the
connection or the guard collapses to one bucket for the whole fleet.

## Idempotent creation

Three operations create something a traveller can be charged or committed to, and all three
require an `Idempotency-Key` header declared in the contract:

- `POST /v1/trips`
- `POST /v1/commerce/handoffs`
- `POST /v1/me/price-watches`

The first request runs normally and its response is recorded against the traveller and key.
A retry with the same key and the same body replays the recorded response verbatim, marked
`Idempotent-Replay: true`, and never reaches the handler. A retry with the same key but a
different body is refused with 409, because that is a client defect, not a retry.

Records are scoped to the traveller, so one traveller can never replay another's response.
Server errors are not recorded: a 500 must remain retryable. Records expire after 24 hours.

## Parameter binding

The generated server binds path, query and header parameters before any middleware runs. A
malformed identifier therefore cannot reach the authentication or admission layers, so the
router installs an error handler that answers those failures with an ordinary Problem
Details document carrying no parser internals.
