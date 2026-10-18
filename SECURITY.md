# Security

## Reporting a vulnerability

Report suspected vulnerabilities privately. Do not open a public issue, pull request or
discussion describing an exploitable defect.

Send the report to the address published in the repository security policy on the hosting
platform. Include affected component, version or commit, reproduction steps, observed impact
and any proof of concept. Expect acknowledgement within three business days.

Do not test against production systems, do not access data belonging to other people and do
not perform denial of service testing.

## Secret handling

Never commit:

```
API keys, database credentials, private keys, tokens, OAuth secrets,
production URLs containing secrets, passport data, real user PII fixtures
```

`.env.example` carries variable names and non-secret sample values only. Real environment
files are ignored by Git. Configuration is validated at process startup and the process fails
fast when a required value is missing. The backend never falls back to an insecure default.

If a secret is committed, treat it as compromised. Rotate it first, then remove it from
history. Rewriting history without rotating is not a fix.

## Personal data

NUVRO handles travel data that is personal and, in the case of passport and nationality
context, sensitive.

- Passport, nationality, date of birth and document numbers are processed only where an entry
  requirement genuinely needs them.
- They are never written to logs, traces, metrics, analytics events or crash reports.
- They are never placed in URL paths or query strings.
- Test fixtures use synthetic data. Real user data is never copied into the repository.
- Client-side caches of personal data use platform-protected storage, never plain files.

## Logging rules

Structured logs carry a request identifier and enough context to debug. They never carry:

```
authentication credentials, bearer or refresh tokens, session identifiers,
passport or document data, full payment instruments, provider API keys,
raw provider payloads containing personal data
```

Errors returned to clients carry a stable code and a safe message. SQL, stack traces,
provider errors, credentials and filesystem paths never leave the backend.

## Transport and server posture

The HTTP server sets explicit read header, read, write and idle timeouts and a maximum header
size. It is never left unbounded. Outbound provider clients set explicit timeouts, honour
context cancellation and bound response body size. Retries apply only to idempotent
operations or operations protected by an idempotency key.

## Dependency scanning

Backend dependency and vulnerability scanning runs in CI:

```
go vet ./...
go test -race ./...
govulncheck ./...
golangci-lint run
```

`gosec` runs through the configured lint pipeline. Dependabot proposes dependency updates.
A vulnerability with a fixed version available is upgraded before new feature work in the
affected area.

## AI boundary

Model providers are adapters. Model output is validated, bounded, parsed and converted into
domain types before use. Model output is never trusted as domain state, and never used as the
source of truth for prices, availability, weather, hours, routes, entry requirements,
exchange rates or live events.
