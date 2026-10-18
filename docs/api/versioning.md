# Versioning and compatibility

Public product URLs use `/v1`; health infrastructure URLs are unversioned. Mobile release
versions, backend build versions and API major versions are independent. A patch version
never appears in a URL.

Released contract changes require a compatibility assessment. Prefer additive optional fields
and new operations. Removing a field or enum value, changing a type, requiring a previously
optional value, or changing resource semantics is breaking. A new enum value also requires
checking how generated clients decode it.

## Phase 1 changes against Phase 0

Phase 1 replaced health `checked_at` with `checkedAt` and the nested error envelope with RFC
9457 Problem Details, then released the protected profile and trip operations. No product API
and no generated mobile client existed at the Phase 0 baseline, so these established the
first released contract rather than breaking a deployed one. The same reasoning must not be
applied to a future `/v1` change once clients are shipping.

Trip deletion stays unreleased on purpose. When it arrives it will be an archive lifecycle
transition, so a traveller cannot destroy trip history with one request.
