# 0001. Native mobile clients

Date: 2026-09-07

## Status

Accepted

## Context

NUVRO is a travel companion that runs while a traveller is moving through an unfamiliar
place. It needs location, maps, offline behaviour, background updates, notifications, camera,
wallet-adjacent flows, and a level of interface polish that carries a premium product.

The alternatives were React Native, Flutter, or a web view shell.

## Decision

iOS is Swift and SwiftUI. Android is Kotlin and Jetpack Compose. Neither is a cross-platform
runtime, and a web view is never the primary interface architecture.

iOS ships first. Android follows, with its architectural home reserved from the start so that
adding it is not a repository rewrite.

## Consequences

Two interface implementations must be built and maintained. That cost is accepted, and it is
bounded by keeping all canonical travel intelligence in Go: Match Score, ranking, budget
rules, itinerary orchestration, provider aggregation and entry requirement reasoning are
written once. The clients present results, they do not compute them.

Android deliberately does not copy iOS visual behaviour. Liquid Glass is an Apple platform
behaviour; Android preserves NUVRO brand identity using native Android conventions.

The risk is behavioural divergence between the platforms. The mitigation is structural: the
OpenAPI contract is the single source of truth, and no client reimplements backend logic. The
same request returns the same result on both platforms because only one implementation of
that logic exists.
