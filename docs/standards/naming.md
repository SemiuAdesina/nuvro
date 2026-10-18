# Naming

Names communicate purpose. Avoid abbreviations unless universally understood, and avoid
generic names entirely.

Rejected:

```
data1  temp  thing  stuff  util  helper  manager
obj    info2 newService     misc  common  sharedStuff
```

Prefer the domain vocabulary:

```
Trip  Destination  Traveler  Itinerary  MatchScore  Budget
Locale  Requirement  Activity  Booking  Recommendation
```

Booleans read as statements:

```
isEligible  hasPassport  canReplan  shouldNotify
```

## Go

Packages are lowercase single words, with no underscore and no mixed caps: `matching`,
`routing`, `budget`, `commerce`.

Identifiers use `MixedCaps` and `mixedCaps`. Avoid package stutter: `matching.Score`, never
`matching.MatchingScore`.

Interfaces are small and defined by the consumer, not the implementation. A one-method
interface takes a behaviour name: `Reader`, `Writer`, `Finder`, `Matcher`, `Resolver`.

Multiword filenames use `snake_case.go`. A test file sits beside its source:

```
score.go
score_test.go
```

## Swift

Modules and types use `UpperCamelCase`. Functions and properties use `lowerCamelCase`. A file
is named for the primary type it declares.

```
DestinationCard.swift        DestinationCardTests.swift
TripViewModel.swift          TripViewModelTests.swift
```

Suffixes are used only when accurate: `View`, `ViewModel`, `Repository`, `Client`, `Store`,
`Coordinator`. Protocols are never prefixed with `I`, and Hungarian notation is never used.

## Kotlin

Packages are lowercase. Types and files use `UpperCamelCase`. Functions and properties use
`lowerCamelCase`.

```
DestinationRepository.kt     DestinationRepositoryTest.kt
```

Gradle module paths follow the boundary:

```
:feature:onboarding   :feature:discover
:core:network         :core:designsystem
```

## SQL

Tables and columns use `lowercase_snake_case`. Migration descriptions use the same.

Never name a column `data`, `value`, `thing`, `info` or `misc` unless the domain genuinely
has that concept. A column name says what the value means.

Constraints are explicit. Invariants PostgreSQL can enforce are enforced by PostgreSQL, not
only by application validation.

## Money, time and locale in names

- A monetary field carries its unit: `amount_minor`, never `amount` or `price`.
- A timestamp carries its zone semantics: `observed_at` for a UTC instant,
  `local_departure_time` with a separate `timezone` column for destination local time.
- Language and locale are distinct concepts and never share a name. Passport and nationality
  context is distinct from current location.
