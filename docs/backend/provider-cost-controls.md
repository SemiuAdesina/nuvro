# Provider cost controls

Every provider call costs money. These limits are part of the architecture, not tuning.

## Field masks

Places API charges by the fields requested, so each operation asks for exactly what it uses.

| Operation | Fields requested |
| --- | --- |
| Nearby and text search | id, displayName, formattedAddress, location, primaryType, types, rating, userRatingCount, priceLevel, currentOpeningHours.openNow, photos |
| Place details | the above plus regularOpeningHours and utcOffsetMinutes |

No operation requests reviews, website, phone numbers or editorial summaries. Adding a field
to a mask is a deliberate change with a cost consequence.

## Call budget per itinerary generation

The pipeline is bounded by construction.

| Stage | Calls |
| --- | --- |
| Candidate discovery | 1 nearby search, cached 15 minutes |
| Route evidence | 1 route matrix over at most 12 points |
| Weather | 1 window per trip, cached 1 hour |
| Media | reuses photo metadata already returned by search |

There is no per-place details call, no per-place photo call and no per-leg route call. A
twenty-place result costs one search, not sixty requests.

## Caps

| Limit | Value |
| --- | --- |
| Candidate pool after pruning | 24 |
| Route matrix points | 12 |
| Nearby results returned to a client | 20 |
| Destination preview places | 8 |
| Destination media assets | 5 |
| Provider response body | 4 MiB |
| Provider timeout | 5 seconds, configurable to at most 30 |

## Failure economics

Rate-limited responses are not retried. A 429 maps straight to a safe error and the affected
section degrades. There is no retry loop and no circuit breaker, because neither is justified
before real provider behaviour is observed.

## Commerce partners

No commerce partner is configured, so no request is billed today. These limits are in place
before any adapter is written.

| Control | Value |
| --- | --- |
| Offer cache lifetime | 20 minutes, keyed by opaque offer identifier |
| Offers returned per search | 30 |
| Price watch batch per worker tick | 100 |
| Price watch interval | 15 minutes, configurable 1 minute to 6 hours |
| Repricing frequency | 6 to 48 hours, tightening as departure nears |
| Active watches per traveller | 20 |
| Webhook payload ceiling | 256 KiB |

A search fans out across configured partners once and caches the result. There is no
per-offer detail call and no repeated pricing call inside one search. Watch repricing reruns
a search rather than polling an expired offer, so it costs one partner query per due watch,
not one per offer previously shown.
