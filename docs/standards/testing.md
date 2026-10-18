# Testing

## The rule

Every behaviour-bearing handwritten source file has meaningful tests.

```
score.go              score_test.go
rank.go               rank_test.go
DestinationCard.swift DestinationCardTests.swift
TripViewModel.kt      TripViewModelTest.kt
```

A test must validate behaviour. A test that constructs a value and asserts nothing, or exists
only to satisfy a filename rule, is a defect. There are no empty test files.

## Documented exceptions

These file kinds do not require a one-to-one test file. They are covered by the tests of the
code that consumes them.

| Kind | Reason | Example in this repository |
| --- | --- | --- |
| Pure data structures | No behaviour to exercise | Request and response payload structs |
| Interfaces and protocols | No implementation to exercise | `health.Probe`, `runner.Job`, `clock.Clock` |
| Constant declarations | No behaviour | `apperror` code constants, exercised through `Status` |
| Generated code | Reproducible from its source contract | `internal/adapter/postgres/sqlc` |
| Entrypoint wiring | Exercised through the application it builds | `cmd/api/main.go`, `cmd/worker/main.go`, `internal/app/wiring.go` |
| Resource declarations | No behaviour | Asset catalogues, string catalogues |
| Test support code | Exercised by the tests that use it, in the same package | `backend/internal/contract` (tested directly) |
| Platform adapters over Apple frameworks | Their behaviour is the framework's; the logic they feed is tested through the protocol they implement | `KeychainSessionStore`, `SystemPermissionAuthority`, `ConnectivityMonitor.start` |
| SwiftUI view bodies | Exercised through the models they render and the UI tests | `NuvroAuthentication` views, picker sheets |

Every current exception in the repository is in the table above. Adding a new kind of
exception means adding a row here in the same change.

## Backend test types

| Type | Where | What it proves |
| --- | --- | --- |
| Unit | Beside the package | A single behaviour in isolation |
| Handler | Beside the handler, using `net/http/httptest` | Status, headers, envelope and safety of the response |
| Integration | `backend/tests/integration` | Real wiring across layers, with real backing services |
| Contract | `backend/tests/contract` | The OpenAPI document and the implementation agree |
| Fuzz | Beside the parser | Hostile and malformed input does not panic or misparse |

Contract tests read the documents under `contracts/openapi` and assert that responses carry
exactly the declared fields, that every required field is present, and that enumerated values
are declared. A response field that is not in the contract fails the build.

Fuzz targets that exist: `money.ParseCurrency`, PostgreSQL identifier parsing in
`adapter/postgres`, `middleware.ParseIdempotencyKey`, the Redis key builder and
`cursor.Decode`. Each runs
as a seed-corpus unit test under `go test ./...` and as a fuzzer with
`go test -run '^$' -fuzz=<Name> -fuzztime=10s ./<package>`. Future candidates, once the
code exists: cursor decoding, provider response parsers, affiliate callback parsers,
external payload normalisation and traveller text normalisation.

## Race detector

The backend suite runs under `-race` locally through `make verify` and in CI. Concurrency
code, notably the worker runner and anything orchestrating parallel provider retrieval, is
tested with cancellation and failure paths, not only the happy path.

## iOS tests

Package unit tests live in `Tests/<Module>Tests` beside the module. State and use case
behaviour is tested directly. UI tests cover important end-to-end flows, and accessibility
behaviour is tested where a screen makes an accessibility promise.

Tests use Swift Testing. Model tests assert against the exact JSON the backend contract
declares, so a contract change breaks the client tests rather than shipping silently.

## Android tests

Unit tests, view model and state tests, repository tests and Compose UI tests live beside
their module. Instrumented tests are used only where platform behaviour genuinely requires a
device or emulator.

## What tests must not do

- Assert on a real external provider. Providers are stubbed through the consumer's interface.
- Depend on wall clock time. Inject `clock.Clock` and use `clock.NewFixed`.
- Contain real personal data, credentials or passport information. Fixtures are synthetic.
- Assert on log text as a substitute for asserting on behaviour, except where the log is the
  behaviour under test.
- Share mutable state between test functions.

## Running

```bash
make verify
```

```bash
make backend-test
```

```bash
make backend-race
```

```bash
make ios-test
```

```bash
make android-test
```

A command that cannot run because a toolchain or project does not exist reports SKIPPED with
the reason. It never reports success for work it did not do.

## Phase 1 integration

`make backend-integration` requires disposable PostgreSQL and Redis URLs in
`NUVRO_TEST_DATABASE_URL` and `NUVRO_TEST_REDIS_URL`. Missing URLs fail explicitly.
Tests apply and reverse migrations, exercise sqlc-backed domain mapping and atomic
trip/membership operations, enrol an account and persist profile and preference patches,
page trips with real cursors and reject a replayed one, and validate constraints and real
Redis cancellation.
Ordinary `go test ./...` needs no Docker. CI runs integrations with ephemeral services.
