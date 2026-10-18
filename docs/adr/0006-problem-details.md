# 0006. RFC 9457 errors

Date: 2026-09-08

Status: Accepted

## Context

Phase 1 establishes the backend contract pipeline while preserving ADRs 0001–0004.
The Phase 0 contracts were manually checked and had no selected generator.

## Decision

Replace the Phase 0 error envelope with application/problem+json. Standard fields and controlled code/requestId/errors extensions carry safe client messages. No validator/SQL/provider diagnostic is returned verbatim. All generated and middleware error hooks share the same renderer.

## Consequences

Contract changes require regeneration and compatibility review. Domain and persistence
models remain separate from generated API types. See docs/api for commands and gates.
