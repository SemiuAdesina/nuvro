# 0007. Generated native mobile clients

Date: 2026-09-08

Status: Accepted

## Context

Phase 1 establishes the backend contract pipeline while preserving ADRs 0001–0004.
The Phase 0 contracts were manually checked and had no selected generator.

## Decision

The released OpenAPI source will drive Apple Swift OpenAPI Generator with URLSession and a type-safe generated Kotlin client. Keep native clients and map generated transport models into presentation models. Mobile generation waits until its toolchain and networking modules can verify it; no production client is generated from prototypes or drafts.

## Consequences

Contract changes require regeneration and compatibility review. Domain and persistence
models remain separate from generated API types. See docs/api for commands and gates.
