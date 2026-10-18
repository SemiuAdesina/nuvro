# Phase 1 foundation

The modular monolith, native clients, package names and two serving processes are
unchanged. Tool commands `cmd/contract` and `cmd/migrate` are operator/build tools,
not additional services. Runtime implementation remains under internal.

## Persistence and transaction boundary

Goose migration `20260908000100_account_trip_foundation.sql` creates users,
user_identities, user_profiles, traveler_preferences, trips, trip_members and
saved_destinations. Internal identifiers are independent of provider subjects.
Preferences use typed columns and arrays, not a catch-all JSON blob. Nationality,
passport scans, itinerary, budget, documents, commerce and bookings are not persisted.

sqlc 1.31.1 targets PostgreSQL/pgx/v5. Generated rows remain in adapter/postgres/sqlc.
Store maps UUID/date/timestamp values into domain models. `plan.Creator` owns the
transaction operation: insert trip then owner membership through the same Store.
The persistence adapter places the transaction in a private context key tied to
its Store, rejects nested transactions and rolls back on error/panic/cancellation.
No third-party I/O is performed inside the transaction.

The initial migration supports destructive down for disposable development/test
rollback only. Never run down against retained product data. Once shared/applied,
fixes use a new timestamped migration; future destructive evolution uses
expand → backfill → contract. PostgreSQL timestamps are timestamptz and connections
set UTC. Calendar trip dates use date. Origin/destination remain optional.

Trip owner consistency is established by the atomic application operation. Database
foreign keys, unique owner membership and checks enforce independently expressible
constraints. Future ownership transfer must update both owner representations in
one authorized transaction; no ownership-transfer or invite API exists.

## Observability and lifecycle

OpenTelemetry SDK traces and metrics are initialized without a vendor/exporter.
W3C trace propagation and request spans work; HTTP metrics are recorded. Provider
request/latency/failure and DB/Redis latency instruments are defined for adapters.
Provider instrumentation is activated with actual integrations. An exporter/collector
configuration remains a deployment decision; this phase does not claim remote export.
Use slog, not experimental OTel logging. Labels exclude user IDs and raw URLs.

Configured DB/Redis probes use driver cancellation and a shared two-second readiness
deadline. HTTP shutdown is bounded, force-closes if draining times out, closes pools
and Redis and shuts down telemetry. Worker cancellation remains the Phase 0 runner.
Worker has no jobs/dependencies to close until a real job is registered.

## Capability boundaries

All engine inputs below are domain evidence/preferences, never OpenAPI/sqlc structs.
Ports are declared by their consuming engine when it is implemented. This table
refines the registry without populating unused packages with placeholder code.

| Engine | Input → output | Consumer-owned capabilities |
| --- | --- | --- |
| E01 understanding | traveler text + locale → validated intent | IntentInterpreter, structured output only |
| E02 matching | trusted evidence + preferences + weights → score/components | DestinationEvidenceReader; no AI-generated score |
| E03 recommendation | intent + candidate evidence → organic ranked destinations | CandidateReader; no commerce inputs |
| E04 intel | destination + dates → attributed destination facts | DestinationEvidenceReader, S01 selection |
| E05 itinerary | trip constraints + facts → feasible itinerary proposal | ItineraryReasoner, PlaceReader, RouteReader |
| E06 routing | points + local dates/modes → timed routes | RouteReader |
| E07 replanning | current plan + disruptions → validated plan changes | DisruptionReader, ItineraryReasoner |
| E08 weather | coordinates + dates → freshness-aware outlook | Forecaster |
| E09 places | geographic context + preferences → places/events | PlaceReader, EventReader, ActivityReader |
| E10 language | text/audio context + languages → translation | TranslationReasoner, AudioTranslator |
| E11 requirements | consented passport/nationality/residence + destination/dates → sourced requirements | EvidenceReader; model exists, no rules invented |
| E12 budget | quantities + flight/stay/activity prices + FX → integer-minor-unit estimate | FlightReader, StayReader, ActivityReader, CurrencyReader |
| E13 commerce | already-ranked options + booking intent → attributed offers/outcomes | ESIMReader, InsuranceReader, TransferReader, ConversionRecorder |
| E14 personalization | consented profile/preferences/history → preference signals | PreferenceReader |
| S01 media | destination/place + dates/context → selected media | CandidateReader; domain and port exist, no provider |

AI may explain the computed MatchScore via ExplanationGenerator. Score metadata
includes weightsVersion and evidenceVersion. The model contains no affiliate payout,
commission or provider SDK type. E13 can consume ranked outputs but cannot feed
organic ranking/scoring. Scoring/commerce behavioral invariant tests are required
when those engines become real; current enforcement is architectural.

Provider calls accept context, bounded response sizes, explicit timeouts and safe
error mapping. Actual independent fan-out uses cancellation and bounded concurrency.
Retries require safe/idempotent operations, retryable failure classes, remaining
deadline, bounded exponential backoff/jitter and maximum attempts; respect provider
retry headers. Do not add retries or circuit breakers before an actual provider.
No DocumentStore port is added until document implementation; binaries eventually
use object storage, with a separate vendor, retention and security decision.

## S01 Destination Media Intelligence

S01 supports E04 and other engines; it does not replace E01–E14. CandidateReader
returns candidate metadata and short-lived provider references privately to S01.
Selection eventually considers quality, season, trip dates, landscape/crop suitability,
attribution, stable rotation, expiry and fallbacks. Mobile receives only MediaAsset.
Google Places Photos, Unsplash, Pexels and a curated library are future candidates.
There is no fake provider and no invented image URL in runtime code.

The concurrently proposed E15 feasibility entry remains a proposal. It does not
change the locked fourteen core engine registry in this phase.

## Released protected surface

Identity is real. `adapter/external/apple` verifies Sign in with Apple identity tokens
against Apple's published keys with bounded, cached retrieval; `usecase/account` maps a
verified identity to an internal account and enrols an unknown one atomically, recovering
from a concurrent first sign-in through the identity unique constraint. A configured
development credential is the only alternative and is refused outside development by both
configuration validation and its constructor.

`usecase/profile` reads and patches the profile and preferences inside one transaction per
operation. `usecase/plan` gained keyset listing over `(created_at, id)` with HMAC-signed
cursors bound to the listing scope and the principal, and an update path that checks the
member role, reads and writes inside a single transaction. Transport maps generated
three-state nullable DTOs to the domain `change.Change` vocabulary, which refuses to clear a
required field. See [authentication](../api/authentication.md) and
[conventions](../api/conventions.md).
