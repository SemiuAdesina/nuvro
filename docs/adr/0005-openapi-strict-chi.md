# 0005. OpenAPI 3.1 and strict Chi generation

Date: 2026-09-08

Status: Accepted

## Context

Phase 1 establishes the backend contract pipeline while preserving ADRs 0001–0004.
The Phase 0 contracts were manually checked and had no selected generator.

## Decision

OpenAPI 3.1.0 is the released source; pinned oapi-codegen 2.8.0 emits models, chi-server, strict-server and an embedded specification. nethttp-middleware 1.2.0 performs schema and security validation before adapters. Drafts never drive production generation. Generated files are committed, compiled and exempt from handwritten gates. CI compares reproducible temporary output.

## Consequences

Contract changes require regeneration and compatibility review. Domain and persistence
models remain separate from generated API types. See docs/api for commands and gates.
