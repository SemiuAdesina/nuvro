# Caching

Every cached value is keyed `nuvro:v1:<capability>:<digest>`, so a namespace change is a
version bump and never a silent collision. Redis is a cache and an admission store, never a
system of record: everything in it can be lost without losing user data.

| What | TTL | Why that long | Invalidated by |
| --- | --- | --- | --- |
| Destination descriptions | provider freshness | The upstream describes how stale it may be. | Freshness metadata |
| Commerce offers | 20 minutes | Long enough to survive a handoff, short enough that a quoted price is still real. | Expiry only |
| Personalization context | 10 minutes | Inference changes slowly; a stale minute costs a slightly worse ranking. | Any insight edit, history toggle or clear |
| Idempotency replays | per operation | A retry of the same write must return the first answer. | Expiry only |
| Rate limit buckets | per class | | Expiry only |
| Apple signing keys | `NUVRO_APPLE_KEY_TTL`, 1h | Apple rotates keys; a shorter window costs a fetch. | An unknown key id forces a refetch, floored at one minute |

Two rules hold everywhere. A cache miss must produce the same answer as a hit, only slower —
no code path returns a degraded result because Redis was empty. And a cache failure is never
fatal to a read: the value is recomputed.

The one place caching was a correctness risk was commerce offers. `money.Amount` has unexported
fields, so a cached offer round-tripped through JSON came back with a zero price. It now has
explicit `MarshalJSON` and `UnmarshalJSON` with tests, because a silently zeroed price is worse
than no cache at all.
