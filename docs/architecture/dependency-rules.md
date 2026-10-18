# Dependency rules

Dependency direction is the property that keeps NUVRO a modular monolith rather than a large
ball of Go. It is enforced by `scripts/check-go-deps.sh` in `make verify` and in CI.

## Direction

```
domain  ←  engine  ←  usecase  ←  transport
```

Adapters sit outside the business core. They implement interfaces declared by their
consumers. The composition root wires concrete implementations to those interfaces.

## Allowed

| Layer | May import |
| --- | --- |
| `domain` | Standard library and approved value dependencies only |
| `engine` | `domain` |
| `usecase` | `domain`, `engine` |
| `transport` | `usecase`, `domain` |
| `adapter` | External systems, `domain`, and the contracts required to implement a port |
| `worker` | `usecase` |
| `platform` | Standard library, and configuration types |
| `app` | Anything required to compose the application |

`internal/apperror` and `internal/platform` are cross-cutting. Any layer may import them.
They contain no travel logic.

## Forbidden

```
domain  → adapter
domain  → transport
domain  → usecase
domain  → engine

engine  → adapter
engine  → transport

usecase → transport
usecase → a concrete external provider

adapter → transport
adapter → usecase
```

Nothing outside `internal/app` imports `internal/app`.

## Interfaces belong to the consumer

An engine that needs weather declares the smallest interface that satisfies it, inside the
engine package:

```go
package weather

type Forecaster interface {
    Forecast(ctx context.Context, at Place, on Window) (Outlook, error)
}
```

An adapter under `internal/adapter/external` implements that interface against a real
provider. The engine never imports the adapter, never learns the provider's name, and can be
tested with a stub.

There is no repository-wide `ports` package. A single package holding every interface in the
system recreates the coupling these rules exist to prevent.

## No junk drawer packages

These directory and package names are rejected by the gate:

```
utils  util   helpers  helper  common  misc
shared stuff  base     general manager service services
```

Name a package for the capability it provides: `money`, `clock`, `telemetry`, `routing`,
`matching`, `budget`, `localization`. If a package cannot be named for one capability, it is
holding more than one and should be split.

## Client rules

The same principle applies to the mobile clients.

- A SwiftUI View or a Composable renders state and emits intent. It performs no networking,
  no persistence access and no business orchestration.
- Feature modules depend on models, domain, the design system, localization, and location
  where genuinely required. They do not depend on each other.
- `NuvroData` implements repositories. `NuvroNetworking` owns transport. `NuvroDomain` holds
  client-side repository protocols and client use cases.
- No client reimplements Match Score, destination ranking, budget scoring, itinerary
  orchestration or entry requirement reasoning. Those live in Go so that iOS and Android
  return identical results for the same request.

## Running the gate

```bash
./scripts/check-go-deps.sh
```

It reports the file, the offending direction and the exact import. It also rejects junk
drawer package names anywhere under `backend/internal`.
