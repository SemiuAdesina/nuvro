# Engine registry

The 14 NUVRO engines are Go package boundaries under `backend/internal/engine/`. They are not
14 services, and they are not 14 machine learning models.

## What an engine is

```
NUVRO Go engine
      │
      ├── deterministic rules
      ├── scoring
      ├── orchestration
      ├── live provider data
      ├── persistence
      ├── cache
      └── AI where reasoning is genuinely useful
```

An engine owns a capability. It declares the narrow interfaces it needs from the outside
world and never imports a concrete provider, an adapter or the transport layer.

## The registry

| ID | Package | Capability |
| --- | --- | --- |
| E01 | `engine/understanding` | Turns a traveller request into structured travel intent: budget, dates, party, constraints, vibe. |
| E02 | `engine/matching` | Vacation Match Score. Deterministic, explainable, NUVRO-owned. |
| E03 | `engine/recommendation` | Candidate generation and organic ranking of destinations. |
| E04 | `engine/intel` | Everything About This Place. Assembles destination intelligence from trusted facts. |
| E05 | `engine/itinerary` | Day and trip structure, sequencing, feasibility. |
| E06 | `engine/routing` | Route and time between places, including transfer feasibility. |
| E07 | `engine/replanning` | Adaptive Travel Mode. Repairs a plan when reality changes. |
| E08 | `engine/weather` | Weather intelligence over provider forecasts and climate context. |
| E09 | `engine/places` | Places, activities and local discovery. |
| E10 | `engine/language` | Language companion, translation and phrase context. |
| E11 | `engine/requirements` | Entry and visa requirements. High consequence, provider sourced. |
| E12 | `engine/budget` | Trip cost and budget. Where Can My Money Take Me? |
| E13 | `engine/commerce` | Booking and affiliate. Strictly separated from organic ranking. |
| E14 | `engine/personalization` | Profile, history and preference signals applied to ranking inputs. |
| E15 | `engine/feasibility` | Travel Feasibility Intelligence. Combines the traveller's residence, passport countries, permits, departure airport, currency and paperwork preference with entry, flight, cost and FX evidence to expose Access and Affordability separately from the Personal Match. Core since Phase 2; the client never computes any of it. |

E01, E02, E03 and E15 are implemented and serve `POST /v1/discover/matches`; see
[discover-intelligence.md](discover-intelligence.md). E04, E05, E06, E08, E09, E12 and S01 are
implemented as of Phase 3 and serve destination detail, itinerary, budget and nearby search;
see [destination-intelligence.md](destination-intelligence.md),
[itinerary-engine.md](itinerary-engine.md) and
[live-provider-flow.md](live-provider-flow.md). E15 is a core engine, not a proposal. E11
requirement evidence has domain and port definitions only. E07, E10, E13 booking and affiliate is implemented as of Phase 5 and serves the commerce search, handoff
and price watch endpoints; see [commerce.md](commerce.md),
[affiliate-attribution.md](affiliate-attribution.md) and
[price-monitoring.md](price-monitoring.md). E14 personalization is implemented as of Phase 6
and serves the personalization settings, insight and trip feedback endpoints; explicit
preferences always outrank inferences, and personalization applies only at the Discover
intent boundary, never inside a ranking engine. See
[personalization.md](personalization.md) and
[../privacy/personalization-data.md](../privacy/personalization-data.md). Git does not track empty directories, and no placeholder file
is added to keep one alive. This table is the canonical record.

## Access and affordability are separate from the match

A recommendation exposes three independent readings: Personal Match (E02), Access (E15 over
E11) and Affordability (E15 over E12 with FX normalisation). They are never collapsed into one
unexplained percentage. Access is described in logistics language ("Easier entry", "More
preparation", "Visa required", "Online application available", "Check requirements"), never
in terms of the traveller's passport or identity. A profile without passport context is marked
incomplete and access results are not presented as personalised.

## Vacation Match Score

This is the invariant that defines the product.

```
Live and trusted data
        +
Traveller preferences
        +
NUVRO scoring weights
        ↓
Deterministic score
        ↓
AI explanation
```

A language model never produces the number. It explains a score that has already been
computed. A score must be reproducible from its inputs and weights, and explainable in terms
of which factors contributed.

## Commerce independence

**Affiliate payout never influences Vacation Match Score or organic recommendation ranking.**

`engine/commerce` may attach booking options, attribution and partner links to a result that
has already been ranked. It cannot participate in producing the ranking. If sponsored
placement is ever introduced, it is explicitly labelled and rendered separately from organic
results.

This rule is not an implementation detail. A change that lets commerce data reach the ranking
inputs is a defect regardless of how it is written.

## AI boundary

AI may interpret intent, summarise, explain, personalise, generate itinerary candidates,
replan and translate over facts it was given.

AI must never be the source of truth for:

```
flight prices        hotel availability   weather
opening hours        routes               visa requirements
entry requirements   exchange rates       booking availability
live events
```

Model output that feeds application logic is requested as structured output, then validated,
bounded, parsed and converted into domain types before use. Model selection is configuration,
not a constant scattered through the codebase, so cheap and expensive models can be routed
per operation later.

## Entry requirements

Visa and entry requirements are high-consequence travel information. They come from
authoritative, current provider data. A result retains:

```
source
retrieval timestamp
traveller nationality and passport context
destination
travel dates
```

AI may simplify and explain supplied requirements. It may not invent them.

## Destination media

Photography for destinations and places is **S01 Destination Media Intelligence**, a supporting backend subsystem consumed by E04 and other engines. It is not a client concern and does not replace any of E01–E14. The client never chooses a provider, looks up a photo, or
scores a candidate. It renders the media the API returns:

```text
Destination / place
      ↓
candidate photo retrieval (provider adapters)
      ↓
quality validation and crop suitability
      ↓
context and seasonal filtering (trip dates, content context)
      ↓
provider attribution
      ↓
controlled selection, stable for the session
      ↓
DestinationMedia in the API response
```

The released shared MediaAsset schema carries URL, dimensions, aspect ratio, optional dominant colour, alt text, attribution, freshness and optional expiry. Provider-specific references stay backend-only. Clients show the placeholder while the image loads and
keep it as the fallback when no image is selected. Rotation between candidates happens between
sessions or screens, never while a traveller is reading. Providers under evaluation are listed
in provider-boundaries.md.

## Freshness

Dynamic travel information carries freshness semantics. Weather, price, availability, flight
state, entry requirements, events, opening hours and transport status all change. Domain
models must make stale and live information distinguishable; a value with no observation time
is a design defect.
