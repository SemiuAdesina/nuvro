# Price monitoring

## What is watched

A watch stores the search criteria needed to rerun a query, not a transient offer identifier.
An offer expires in minutes; a watch lives for weeks.

| Product | Target |
| --- | --- |
| Flight | Origin, destination, dates, travellers, cabin |
| Stay | Property, check-in, check-out, travellers |
| Activity | Place, date, travellers |

The watch holds only what rerunning that search requires. It references the trip and user
rather than copying a profile into every row.

## Scheduling

Frequency follows how close the trip is, so distant trips are not polled aggressively before
partner rate limits are known.

| Departure | Next check |
| --- | --- |
| Over 30 days out | 48 hours |
| 7 to 30 days | 12 hours |
| Within 7 days | 6 hours |

The worker asks only for watches that are active, unexpired and actually due, in batches of a
hundred. It never scans every watch on every tick.

## Meaningful change

A change must clear both a percentage and an absolute threshold, defaulting to five percent
and a thousand minor units, before anything is raised. A drop from 400.00 to 399.00 is
recorded as an observation and raises nothing. Arithmetic is integer minor units throughout;
there is no floating point anywhere in the comparison. Mixed currencies are refused rather
than converted.

A meaningful drop moves the baseline down, so a slow decline produces one alert per step
rather than a repeated alert against a stale starting price.

## Alerts

Price monitoring reuses the Phase 4 alert model rather than introducing a second notification
system. It produces `price_dropped`, `price_increased` and `availability_changed` alerts with
the same fingerprint deduplication, so evaluating the same unchanged price twice yields one
alert. Canonical state is a code plus structured parameters; the English sentence is rendered
by the client.

## Expiry

A watch ends when its travel date passes, when the traveller deletes it, or when the trip is
removed by cascade. Expiry runs before every evaluation pass, so an expired watch is never
repriced. Each traveller may hold twenty active watches.

## Entitlement

Price monitoring is Premium. Entitlement is checked server-side on every creation; a client
flag is never trusted. No entitlement backend exists yet, so the boundary returns false and
the endpoint answers that monitoring is unavailable. This is a documented gap, not a fake
authorisation: nothing grants access by default.

Following a NUVRO recommendation to a partner is deliberately not paywalled. Search and
handoff are free; only ongoing monitoring is a subscription feature.

## Repricing

No approved commerce partner is configured, so no repricer is wired. The worker still runs its
expiry pass and does nothing else. Monitoring is not claimed to work until a partner can
actually be re-queried.
