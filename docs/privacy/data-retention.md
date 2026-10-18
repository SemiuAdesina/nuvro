# Data retention and account deletion

Two mechanisms bound how long NUVRO holds data. A retention policy expires records that have
outlived their purpose while the account stays open. Account deletion ends access immediately
and erases what the account owns.

Both are versioned as `retention-v1` in `internal/domain/retention`.

## Retention policy

| Class | Kept for | Why that long |
| --- | --- | --- |
| `personalization_signals` | 730 days | Inference reads at most two years of behaviour, so older signals influence nothing. |
| `commerce_handoffs` | 400 days | The longest partner attribution window, after which a click can no longer be credited. |
| `provider_webhook_events` | 180 days | Replay protection only has to outlive a partner's redelivery window. |
| `trip_alerts` | 180 days | An alert describes a moment during a trip and is not a travel record. |
| `replan_proposals` | 180 days | A proposal is a transient offer to change a plan, not part of the plan. |
| `price_watches` (expired) | 90 days past expiry | A watch that has expired has no monitoring left to do. |

Every rule carries its rationale in code, and a test fails if a class is added without one. The
`retention-sweep` worker applies each rule on `NUVRO_RETENTION_INTERVAL`, six hours by default.

Trips, itineraries and profiles have no expiry. They are the traveller's own record and are
removed when the account is, not on a timer.

## Account deletion

`DELETE /v1/me` accepts a deletion and returns `202` with the schedule. Two things happen at
different times, and the split is deliberate.

Access ends at once. The request marks the account non-active in the same transaction that
records the deletion, so the next sign-in is refused and no new data can be written. From the
traveller's side the account is gone the moment they ask.

Erasure follows after a 30 day grace period. The window exists for operational recovery — a
support-initiated restore after a mistaken or compromised request — not as a user-facing undo.
There is deliberately no cancel endpoint: a closed account cannot authenticate, so an endpoint
it could never reach would be a fiction.

The request body carries at most a 64 character `reason` chosen from a fixed client-side list.
It is never free text, because a free-text field on a deletion form collects exactly the kind
of grievance nobody should be storing about a person who has just left.

## Sign in with Apple

Apple's account deletion guidance asks apps that support Sign in with Apple to revoke user
tokens through the REST API when an account is deleted. NUVRO does.

The iOS app hands the single-use authorization code to `PUT /v1/me/identity/apple` once. The
server exchanges it at Apple's token endpoint for a refresh token, seals that token with
AES-GCM bound to the account id, and stores only the sealed bytes. The plaintext never touches
the database.

Closing an account revokes it immediately, in the same request. If Apple is unreachable the
closure still succeeds — a provider outage must never trap someone in an account they are
trying to leave — and the identity stays queued for the retention sweep to retry. Once Apple
accepts the revocation the sealed token is deleted and the row is stamped `revoked_at`.

A sealed token that can no longer be opened, after a key rotation say, is retired rather than
retried forever. It can never be used again by anyone, including us, so holding it has no value.

## What erasure does

The purge runs as one transaction per account.

Memberships are removed first, then trips are resolved. A trip the departing traveller owned
alone is deleted with everything hanging off it. A trip shared with others passes to the
longest-standing remaining member, with ties broken on identifier so the outcome is
deterministic; that member is promoted to owner. Removing memberships before promoting matters:
a trip may only have one owner row, so the departing owner has to be gone before a successor
can take the role.

The account row is then deleted. Profile, preferences, identities, saved destinations,
personalization signals, learned preferences, trip feedback, price watches and commerce
handoffs all cascade from it.

`commerce_conversions` is the deliberate exception. A conversion is a financial record between
NUVRO and a partner, it carries no user column, and its `handoff_id` becomes null when the
handoff is erased. The money record survives; the link to a person does not.

## The deletion receipt

`account_deletions` outlives the account by design, which is why `user_id` there carries no
foreign key. It records that a deletion was asked for, when erasure ran, and how many trips
were transferred or deleted. It is the only proof the erasure happened, and it holds a bare
identifier with nothing left to join it to.
