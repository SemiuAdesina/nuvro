## What changed

<!-- One paragraph. What behaviour or structure is different after this change. -->

## Why

<!-- The problem this solves. Link the issue or decision record. -->

## Architecture

- [ ] Dependency direction holds: `domain ← engine ← usecase ← transport`
- [ ] Engines depend on consumer-declared interfaces, not concrete providers
- [ ] No business logic added to an HTTP handler, a SwiftUI View or a Composable
- [ ] No canonical travel scoring or ranking logic added to a mobile client
- [ ] Affiliate or commerce code cannot influence organic Match Score or ranking
- [ ] AI output is not treated as the source of truth for a live travel fact

## Repository rules

- [ ] No handwritten source file exceeds 400 lines
- [ ] No explanatory comments, `TODO`, `FIXME`, `HACK` or commented-out code added
- [ ] Every behaviour-bearing file added or changed has meaningful tests
- [ ] No junk drawer package introduced
- [ ] Generated code was regenerated, not hand edited

## Data and security

- [ ] Money is integer minor units with an ISO 4217 currency code
- [ ] Timestamps are UTC RFC3339, and local times keep an IANA timezone identifier
- [ ] No secret, credential, token or real personal data is committed
- [ ] No passport data, token or credential can reach a log, trace or metric
- [ ] Client-facing errors expose a stable code and a safe message only

## Contract

- [ ] `contracts/openapi/` updated when a client-visible request or response changed
- [ ] Contract tests updated

## Verification

```
make verify
```

<!-- Paste the result. If a gate was skipped, say which and why. -->
