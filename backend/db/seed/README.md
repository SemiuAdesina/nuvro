# Development seed data

`destinations.sql` is a **development and test reference catalog**, not a comprehensive
global destination database. It holds fourteen geographically diverse destinations across
Africa, Europe, Asia, the Middle East, North America, South America, the Caribbean and
Oceania, chosen so ranking behaviour can be exercised for travellers departing from any
region rather than only from one market.

It carries static descriptive metadata only: name, country, timezone, coordinates, primary
airport, and climate, interest and style tags. It deliberately holds no visa rules, no
prices, no weather and no availability. Those are dynamic facts that must come from
providers with their own freshness and source, never from a seeded table.

It is plain SQL, not a Goose migration, because seed rows are not schema. Apply it to a
development or test database after the schema migrations:

```bash
psql "${NUVRO_DATABASE_URL}" -f db/seed/destinations.sql
```

The `-- +seed statement` line separates the two inserts so a test harness can execute them
individually; `psql` ignores it as a comment.

Never apply it to a production catalog. A real catalog is a Phase 3 data-sourcing decision
recorded in `docs/backend/provider-gap-report.md`.
