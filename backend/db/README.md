# Database sources

```
migrations/   Goose SQL migrations
queries/      sqlc source queries
seed/         local development seed data
```

The account/trip foundation is migration `20260908000100_account_trip_foundation.sql`.
It creates users, identities, profiles, structured preferences, trips, members and
saved destinations. Queries cover identity lookup and the first internal trip operation.
No seed fixtures, live provider payloads, passport data or document binaries are stored.

Run `make db-migrate`, `make sql-generate` and `make sql-check` from the root.
Generated Go lives in `../internal/adapter/postgres/sqlc/`. Integration tests use
`NUVRO_TEST_DATABASE_URL` pointing to a disposable empty database and exercise
up/down/up and repository rollback. Never point those tests at retained data.

## Migration naming

```
YYYYMMDDHHMMSS_description.sql
```

Lowercase snake_case description. Goose annotations are required machine directives and are
the documented exception to the no-comment rule:

```sql
-- +goose Up
-- +goose Down
```

Rules for schema changes, destructive migration sequencing, money columns, timezone handling
and constraints are in [docs/standards/database.md](../../docs/standards/database.md).
