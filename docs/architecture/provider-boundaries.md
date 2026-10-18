# Provider boundaries

External providers are adapters. They live at the edge of the system, behind interfaces that
their consumers declare.

## The rule

```
engine or use case  →  declares a narrow interface
adapter             →  implements it against a real provider
internal/app        →  wires one to the other
```

An engine never imports an adapter. An engine never knows a provider's name. Swapping a
weather provider changes one adapter and one line of wiring, and no engine code.

## Adapter slots

These categories will exist under `backend/internal/adapter/external`. None of them is
implemented, and none of them has a credential in this repository.

| Category | Purpose |
| --- | --- |
| AI models | Intent interpretation, explanation, translation, itinerary candidates, replanning |
| Places | Points of interest, activities, opening hours |
| Maps | Geocoding, place identity, static context |
| Routes | Travel time and route feasibility between places |
| Weather | Forecast and climate context |
| Flights | Availability and pricing search |
| Stays | Accommodation availability and pricing |
| Activities | Bookable experiences |
| Entry requirements | Visa, documentation and eligibility facts |
| Currency | Exchange rates with observation time and source |
| Events | Dated local events |
| eSIM | Connectivity purchase |
| Insurance | Travel insurance quotes |
| Airport transfers | Ground transport between airport and stay |
| Affiliate systems | Attribution, click tracking, conversion, reconciliation |

## Providers under future evaluation

Named for planning only. No commercial access is assumed, no contract exists, and no
integration has been built or tested.

```
OpenAI
Google Places, Google Routes, Google Weather
Sherpa, Timatic
Expedia and the Booking ecosystem
Viator, GetYourGuide
Airalo
flight search partners
insurance partners
```

Selecting any of them requires an architecture decision record covering commercial terms,
data freshness guarantees, rate limits, failure behaviour and personal data handling.

## Destination media providers

| Slot | Candidates | Notes |
|------|------------|-------|
| Place photos (primary) | Google Places Photos | Attribution metadata is mandatory in the response. |
| Editorial fallback | Unsplash, Pexels | Used when place photos are missing or fail quality checks. |
| Curated library (future) | NUVRO media library | Editorially selected imagery for signature destinations. |

The media adapter implements the small interface S01 Destination Media Intelligence defines.
Selection, scoring, and rotation stay in the engine; adapters only fetch candidates.

## Client requirements

A provider client is never the default `http.Client`. Before an adapter ships it must have:

- an explicit request timeout, and a context deadline honoured on every call
- connection reuse through a configured transport
- a bounded response body, so a hostile or broken provider cannot exhaust memory
- error mapping into `apperror` codes, so provider failure never reaches a client verbatim
- a retry policy applied only to idempotent operations, or operations protected by an
  idempotency key
- traces and metrics, so a slow provider is visible in a request trace

Non-idempotent operations, notably anything that creates a booking, are never retried
automatically without an idempotency mechanism.

## Failure behaviour

Provider failure degrades a result; it does not fail the product. A weather provider outage
should produce a destination result without weather, marked as unavailable, rather than a 500.
Which providers are required and which are optional for a given operation is a decision each
use case makes explicitly.

## Secrets

Provider credentials are read from the environment, grouped by provider, validated at
startup, and never committed. `.env.example` lists the variable names with empty or non-secret
values. A provider key never appears in a log, a trace attribute, an error message or a URL.

## Phase 2 discovery ports

`engine/feasibility` declares `EntryRequirementProvider`, `FlightEvidenceProvider`,
`TripCostEstimator` and `FXRateProvider`. `engine/understanding` declares `Interpreter` and
`engine/recommendation` declares `ExplanationGenerator`. Every one is defined by its consumer
and implemented by nothing today. Status, candidates and selection requirements are recorded
in [the provider gap report](../backend/provider-gap-report.md).

## Failing providers

Every provider call goes through one HTTP client that carries three protections.

**Retry.** A network failure, a 429, or a 502/503/504 is retried with exponential backoff and
jitter, three attempts by default. A 4xx is never retried, because the request itself is
wrong and repeating it only adds load. A malformed body is never retried for the same reason.
GET is always retryable; POST is retried only where the caller marks it safe, so a search may
be repeated but a state-changing call may not.

**Circuit breaking.** After five consecutive failures the circuit opens for thirty seconds
and calls fail immediately with `ErrCircuitOpen` instead of queuing behind a dead dependency.
After the cooldown exactly one probe is allowed through; it closes the circuit on success and
reopens it on failure. The breaker uses the injected clock, so its behaviour is tested rather
than timed.

**Instrumentation.** Each attempt records a provider request, its duration, and a failure
when it failed, labelled by provider name only. Every attempt is counted, so a retried call
shows as three requests and two failures rather than hiding the recovery.

These are attempt-level protections around a single dependency. They do not change the
freshness rules above: a provider that is down still yields an honest unavailable section
rather than a fabricated value.
