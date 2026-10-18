# 0003. OpenAPI as the client and server contract

Date: 2026-09-07

## Status

Accepted

## Context

Two native clients consume one backend. Without a single authoritative contract, Swift and
Kotlin each invent their own response types, drift apart, and the same request starts
rendering differently on each platform.

## Decision

The OpenAPI documents under `contracts/openapi/` are the source of truth for the contract
between the backend and both clients. They are split with `$ref` into paths, schemas,
parameters and responses, so that no contract file grows past the repository line limit.

Clients consume the contract rather than inventing types. Backend contract tests read the
same documents and assert that responses carry exactly the declared fields, that required
fields are present, and that enumerated values are declared.

Generated client code, if introduced later, lives in an explicit `Generated` directory, is
never edited by hand, and is reproducible from the contract. No generator is selected in
Phase 0, because choosing one before there is an API to generate would be a guess.

## Consequences

A response shape change that is not reflected in the contract fails the backend build. This
is verified: adding an undeclared field to the liveness response makes the contract test fail
with the field name.

The contract carries a small test-only dependency, a YAML parser, so that the tests can read
the documents. That is a deliberate trade: a dependency in test code in exchange for drift
being impossible rather than merely discouraged.

The cost is discipline. A new endpoint means writing the contract file as part of the change,
not afterwards.
