# Contracts

`openapi/v1/openapi.yaml` is the only released backend/mobile generation input.
It uses OpenAPI 3.1 with split paths, schemas, parameters and responses, and every source
file stays under 400 lines. It carries ten operations: two unauthenticated health probes and
eight authenticated operations for profile, preferences and trips. Every one of them is
implemented and served; there is no drafts root. Future groups are listed in
[the roadmap](../docs/api/api-roadmap.md) and get a contract file when their slice is real.

The unversioned Phase 0 documents are retained for concurrent iOS prototype work.
They are compatibility artifacts, not production code generation inputs. Their
snake_case types and old error envelope do not define the released Phase 1 API.

`make api-validate`, `make api-generate`, `make api-check` validate references and
reproduce committed Go output. Backend tests validate requests, responses, schema
constraints, Problem Details and draft syntax. Production DTOs use camelCase.

See [API overview](../docs/api/overview.md), [conventions](../docs/api/conventions.md),
[authentication](../docs/api/authentication.md) and [mobile generation](../docs/api/mobile-codegen.md).
