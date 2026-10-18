# Contributing to NUVRO

## Before you write code

Read [CLAUDE.md](CLAUDE.md). It is the enforceable contract, not a suggestion.
Read the relevant document under [docs/](docs/) for the area you are touching.

Inspect existing files before editing. Never overwrite a meaningful file you have not read.

## Branches

```
feat/       new product capability
fix/        defect repair
refactor/   structural change with no behaviour change
test/       test-only change
docs/       documentation only
chore/      tooling, dependencies, CI
```

Keep commits focused. Do not mix a large refactor with unrelated product behaviour.

## Quality gates

Run before pushing:

```bash
make verify
```

This runs the line-count gate, the comment gate, the Go dependency-direction gate and the
backend checks available in your environment. It fails immediately on a violated mandatory
rule. It never reports success for a check it skipped.

## Mandatory rules in review

- No handwritten source file above 400 physical lines.
- No explanatory comments, `TODO`, `FIXME`, `HACK`, `XXX` or commented-out code in
  handwritten runtime source.
- Every behaviour-bearing file has meaningful tests. A test that asserts nothing is a defect.
- No junk drawer packages.
- No business logic in HTTP handlers, SwiftUI Views or Compose UI.
- No canonical travel logic duplicated in a mobile client.
- Dependency direction: `domain ← engine ← usecase ← transport`.
- No secrets, credentials or real personal data in any commit.

## Adding a dependency

1. Confirm it is actively maintained.
2. Confirm the standard library is genuinely insufficient.
3. Confirm it is required now, not anticipated.
4. Pin the version through the package manager.
5. Record architecturally significant dependencies in `docs/architecture/overview.md`.

## Adding a migration

Migrations are Goose SQL files under `backend/db/migrations/` named
`YYYYMMDDHHMMSS_description.sql` with a lowercase snake_case description.
Destructive changes use expand, migrate, contract across releases. Never drop a column or
table in the same release that removes its application usage.
See [docs/standards/database.md](docs/standards/database.md).

## Generated code

Generated code is never edited by hand. Regenerate it from its source contract or schema.
It is excluded from the line-count gate, the comment gate and the one-to-one test rule.
