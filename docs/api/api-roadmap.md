# API roadmap

Released operations are implemented, authenticated, tested against the contract and served.
Everything else is a proposal recorded here, not a registered handler.

| Group | Operation | State |
| --- | --- | --- |
| System | `GET /health/live` | Released, unauthenticated |
| System | `GET /health/ready` | Released, unauthenticated |
| Profile | `GET /v1/me` | Released |
| Profile | `PATCH /v1/me` | Released |
| Profile | `GET /v1/me/preferences` | Released |
| Profile | `PATCH /v1/me/preferences` | Released |
| Trips | `POST /v1/trips` | Released |
| Trips | `GET /v1/trips` | Released, cursor paginated |
| Trips | `GET /v1/trips/{tripId}` | Released |
| Trips | `PATCH /v1/trips/{tripId}` | Released |
| Trips | `DELETE /v1/trips/{tripId}` | Planned as an archive transition, not a destructive delete |
| Trip members | `GET/POST /v1/trips/{tripId}/members`, `DELETE .../members/{memberId}` | Planned |
| Discover | `POST /v1/discover/matches` | Released; E01, E15, E02 and E03 over unconfigured provider ports |
| Discover | `POST /v1/discover/budget-search` | Planned; overlaps `discoverMatches` and will only ship if it earns distinct semantics |
| Destinations | `GET /v1/destinations/{destinationId}`, `.../media`, `POST /v1/destinations/compare` | Planned, needs E04 and S01 |
| Saved | `GET /v1/saved-destinations`, `PUT`/`DELETE .../{destinationId}` | Planned |
| Itinerary | `GET .../itinerary`, `POST .../itinerary/generate`, `PATCH .../items/{itemId}`, `POST .../replan` | Planned, needs E05/E06/E07 |
| Budget | `GET /v1/trips/{tripId}/budget` | Released, needs E12 |
| Nearby | `POST /v1/trips/{tripId}/nearby/search` | Released, needs E09 |
| Language | `POST /v1/trips/{tripId}/translate` | Released, needs E10 |
| Documents | `GET /v1/trips/{tripId}/documents` | Planned, needs a document store decision |
| Alerts | `GET .../alerts`, `PATCH .../alerts/{alertId}` | Planned |
| Stays, Activities, Transport | `POST /v1/trips/{tripId}/{stays,activities,transport}/search` | Planned, needs provider adapters |
| Summary | `GET /v1/trips/{tripId}/summary` | Planned |
| Travel Mode | `GET /v1/trips/{tripId}/active` | Released |
| Itinerary | `POST /v1/trips/{tripId}/replan/{proposalId}/apply` | Released |
| Language | `GET /v1/trips/{tripId}/phrases` | Released |
| Offline | `GET /v1/trips/{tripId}/offline-pack` | Released |
| Commerce | `POST /v1/trips/{tripId}/flights/search` | Released; partner adapters pending |
| Commerce | `POST /v1/commerce/handoffs` | Released |
| Commerce | `GET /v1/me/price-watches`, `POST /v1/me/price-watches`, `DELETE /v1/me/price-watches/{watchId}` | Released; entitlement backend pending |
| Personalization | `GET /v1/me/personalization`, `PATCH`, `DELETE` | Released |
| Personalization | `GET /v1/me/personalization/insights`, `PATCH .../insights/{insightId}` | Released |
| Personalization | `GET /v1/trips/{tripId}/feedback`, `PUT /v1/trips/{tripId}/feedback` | Released |
| Sessions | `POST /v1/auth/sessions`, `POST /v1/auth/sessions/refresh`, `DELETE /v1/auth/sessions` | Released; consumed by iOS |
| Account | `DELETE /v1/me`, `PUT /v1/me/identity/apple` | Released; consumed by iOS |

A planned operation gets a contract file when its vertical slice is real. Empty placeholder
documents are not created to make the table look complete.

## Promotion checklist

1. Domain vocabulary and invariants, with tests.
2. Use case with authorization, transaction boundary and tests.
3. Repository queries, sqlc generation and real database tests.
4. Contract under `openapi/v1`, generated server, handler mapping and contract tests.
5. Wiring in `internal/app`, then `make verify` and `make backend-integration`.
