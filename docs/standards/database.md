# Database

PostgreSQL is the authoritative transactional datastore. Redis is ephemeral.

## Access

pgx/v5 for the driver, sqlc for type-safe Go from hand-written SQL. There is no ORM. SQL is
explicit and version controlled.

```
backend/db/migrations/   Goose SQL migrations
backend/db/queries/      sqlc source queries
backend/db/seed/         local development seed data
```

Generated Go lands in `backend/internal/adapter/postgres/sqlc/` and is never edited by hand.

`backend/sqlc.yaml` is configured for PostgreSQL with the pgx/v5 driver. Generation activates
with the first real migration and query. Running `sqlc generate` against empty schema and
query directories produces nothing, which is correct: there is no schema to model yet.

## Migrations

Goose SQL migrations named:

```
YYYYMMDDHHMMSS_description.sql
```

with a lowercase snake_case description. Goose annotations are a permitted exception to the
no-comment rule because they are machine directives:

```sql
-- +goose Up
-- +goose Down
```

Every migration has an intentional forward path. Rollback behaviour is considered per
migration rather than assumed.

## Destructive change

Never drop a column or table in the same release that removes its application usage. Use
expand, migrate, contract across releases:

1. **Expand.** Add the new column or table. Deploy. Nothing reads it yet.
2. **Migrate.** Write to both old and new. Backfill. Move reads to the new shape. Deploy.
3. **Contract.** Once no deployed version references the old shape, drop it in a later
   release.

A rollback in the middle of this sequence must leave a working system. That is the point of
the sequence.

## Naming

Tables and columns use `lowercase_snake_case`. Column names say what the value means. A
column is never called `data`, `value`, `thing`, `info` or `misc` unless the domain genuinely
has that concept.

## Constraints

Invariants that PostgreSQL can enforce are enforced by PostgreSQL. Application validation is
in addition to database constraints, never instead of them.

Use `NOT NULL` by default. Use foreign keys with an explicit delete behaviour. Use `CHECK`
for domain ranges. Use unique constraints for real uniqueness, including partial unique
indexes where uniqueness is conditional.

## Money

Money is stored as an integer minor unit column plus an ISO 4217 currency column. Never
`float`, never `double precision`, and never a bare numeric with an assumed currency.

```sql
amount_minor  bigint      not null,
currency      char(3)     not null check (currency ~ '^[A-Z]{3}$')
```

A stored currency conversion retains source currency, target currency, the rate, the rate
observation time and the provider.

## Time and timezone

Instants are `timestamptz` and are read and written as UTC. Destination local time is stored
as its own value together with an IANA timezone identifier column; a local time with no zone
is ambiguous and is not stored.

```sql
observed_at   timestamptz  not null,
timezone      text         not null
```

Dynamic travel information carries its observation time so that stale and live data are
distinguishable.

## Redis

Redis holds cache, rate limit state, short-lived session and trip state, coordination and
queue state. It is never the only home of critical trip or account information.

Every entry has:

- a namespaced key, prefixed by the owning capability, for example `nuvro:v1:weather:<id>`
- a single owner, the package responsible for writing and invalidating it
- a TTL where the value can go stale
- a documented invalidation strategy

Losing the entire Redis instance must degrade performance, never correctness.

## Personal data

Passport, nationality and document data are stored only where an entry requirement genuinely
needs them, are never written to logs or traces, and are never placed in a cache key. Test
fixtures are synthetic.
