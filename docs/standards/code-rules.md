# Code rules

These rules are enforced by `make verify` and in CI. `CLAUDE.md` is the short form; this is
the detail.

## File size

No handwritten source file exceeds **400 physical lines**.

| Range | Meaning |
| --- | --- |
| 0 to 200 | Ideal |
| 200 to 300 | Acceptable |
| 300 to 400 | Review for decomposition |
| Over 400 | CI failure |

Applies to `.go`, `.swift`, `.kt`, `.kts`, `.sql` and `.sh`, including tests. Handwritten
configuration stays under the limit wherever practical. OpenAPI is split with `$ref` rather
than allowed to grow into one large document.

Excluded: generated code, lockfiles, machine-generated project files, vendored third party
code and binary assets. The exclusion is never used to hide handwritten code.

```bash
./scripts/check-loc.sh
```

## No explanatory comments

Handwritten runtime source contains no explanatory comments. Names and decomposition carry
the meaning. Architecture explanation belongs in Markdown under `docs/`.

Forbidden:

```
// explanation          /* explanation */       // MARK:
TODO   FIXME   HACK   XXX          commented-out code
```

Allowed, because they are machine directives rather than prose:

```
//go:build            //go:embed           //go:generate
//nolint:...          // Code generated ... DO NOT EDIT.
// swift-tools-version:...
-- +goose Up          -- +goose Down
```

Linting is configured so that no linter demands a documentation comment on an exported
symbol. `.golangci.yml` disables `ST1000`, `ST1020`, `ST1021` and `ST1022` for this reason.

```bash
./scripts/check-comments.sh
```

This rule does not apply to Markdown.

## Functions

Target 40 lines or fewer. A function approaching 60 lines is reviewed for decomposition, and
`funlen` enforces that in the backend.

Avoid deep nesting, boolean argument explosions, large switch statements, hidden side
effects, global mutable state, god objects, god view models and god repositories. Prefer
early returns where they improve clarity.

Do not split a function into meaningless wrappers to satisfy a line count. Decomposition
follows responsibility.

## Visibility

Default to unexported, internal or private. Expose only what another package or module
genuinely needs.

- Go: server implementation lives under `internal`. Exported identifiers are a deliberate
  choice, not the default.
- Swift: prefer `internal` and `private`. `public` is for the module's intended API.
- Kotlin: prefer `internal` and `private`, and expose a narrow module API.

## Reuse

A component becomes reusable when at least one of these is true:

- more than one feature needs it
- it is a stable design system primitive
- it is platform infrastructure
- it is a shared domain concept

Do not build abstraction layers for reuse that has not happened.

## Avoid

```
clever abstractions          premature generic frameworks
deep inheritance             massive base types
unjustified metaprogramming  reflection heavy architecture
hidden global registries     service locator patterns
```

Prefer clarity, explicit dependencies, small modules, testability, typed boundaries and a
minimal public surface.

## Concurrency

Use concurrency where work is genuinely independent, such as parallel provider retrieval.

Every external operation carries a `context.Context` with cancellation and a timeout. Do not
create uncontrolled goroutines, leak goroutines, or start background goroutines inside a
request handler. Concurrency orchestration belongs in an engine or use case boundary, not in
a handler. Do not use concurrency merely because the language offers it.

## HTTP handlers

A handler decodes the request, validates transport concerns, calls a use case, maps the
result or error, and writes the response. It contains no travel business logic. It does not
build SQL, call a provider, or compute a score.

## Formatting

`gofmt` for Go, checked in CI. Swift and Kotlin use one formatter each, never two competing
ones. `.editorconfig` is the whitespace baseline: UTF-8, LF, final newline, no trailing
whitespace, spaces except where the language toolchain requires otherwise.

## Generated code

Generated code lives in clearly identifiable locations, is reproducible from its source
contract or schema, and is never edited by hand. It is excluded from the line count gate, the
comment gate and the one-to-one test rule. The source contract is authoritative; the
generated file is not.

## Dependencies

Before adding one: confirm it is actively maintained, confirm the standard library is
genuinely insufficient, confirm it is required now rather than anticipated, pin the version,
and record architecturally significant choices in `docs/architecture/overview.md`. Popularity
is not a reason.
