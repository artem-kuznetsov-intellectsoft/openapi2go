# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A Go code generator that converts OpenAPI 3.0.x component schemas into Go struct/enum
declarations. `openapi/schema.go` is a hand-written struct model of the OpenAPI 3.0.x Object
tree (unmarshaled via `encoding/json`); `generator` walks `components.schemas` in that
model and emits formatted Go source.

## Commands

```sh
go build ./...
go vet ./...
go test ./...
go test ./generator -run TestGenerate/CompanyDetailResponseDto -v            # single test
golangci-lint run                                                            # lint (config: .golangci.yaml)
gofmt -l .                                                                   # formatting check
make update-golden                                                           # regenerate goldens
```

`make update-golden` runs in three steps on purpose: regeneration first, on its own, then
build, then the full test suite. `go test ./...` builds every test binary up front, so a
single combined invocation would compile the fixture packages' hand-written exec tests
against the goldens `TestGenerate` is concurrently rewriting.

Run the generator CLI directly:

```sh
go run ./cmd/openapi2go generate <openapi-spec-path> [-o output-dir] [-pkg name]
```

## Architecture

- **`openapi/`** — the OpenAPI Object model (`schema.go`), used to unmarshal an OpenAPI 3.0.x
  document into Go structs, plus small runtime support types (`date.go`, `oneof.go`,
  `discriminated.go`) that generated code can copy for itself.
  - **`openapi/clientruntime/`** — `client_runtime.go`, the HTTP plumbing every generated
    client calls: `Client`, `NewClient`, `RequestOption`, `APIError`, and the unexported
    `do`/`decodeJSON`/`decodeError`/`pathParam`/`formatValue` helpers. Like the other support
    files it is copied verbatim into the generated package (as `client_runtime.gen.go`) with
    only its package clause rewritten, so **it must import stdlib only**. It is a subpackage
    because `package openapi` already declares `Response`, and because these names belong to
    generated code rather than to the object model. Its identifiers also share a namespace
    with every generated schema type — hence `HTTPResponse` rather than `Response`, and the
    `reservedClientNames` check in `generator/client.go` that fails generation on a collision.
    Change behavior here first and `generator/client.go` second.
- **`generator/`** — the OpenAPI→Go translation, split by responsibility:
  - `generator.go` — core `generator` type and the `Generate` entrypoint; loads
    `components`, orchestrates the walk/resolve/render passes.
  - `operations.go` — walks `spec.Paths`, registering each operation's Params/requestBody/
    response types (feeding `client.go`'s method generation) in path/method/status order.
  - `schema.go` — the OpenAPI schema → Go type resolution core: struct fields, enums,
    oneOf/discriminated unions, inline object/array types.
  - `render.go` — renders resolved struct/enum definitions to Go source text.
  - `naming.go` — string/identifier utilities (`toPascalCase`, `$ref` unwrapping, sorted-key
    helpers, etc.) shared across the above.
  - `client.go` — one method per operation that has an `operationId`, built from the exact
    type names the walk/resolve passes declared. Each method is a thin wrapper that builds
    query/header/cookie values, makes a single `c.do` call, dispatches on the status code,
    and hands off to the copied runtime — the `Client` type and all the plumbing come from
    `openapi/clientruntime`, not from here, so declaring them in both would be a duplicate
    declaration.
  - `fixtures/` — golden-file test fixtures (see Testing pattern below).
- **`cmd/openapi2go`** — the CLI entrypoint (`generate` subcommand).
- **`make/`** — Makefile includes (`lint.mk`), pulled in by the root `Makefile`.

## Testing pattern

Three layers, and a golden diff alone is not coverage — a new client behavior needs a fixture
entry **and** an assertion in an exec test.

**1. Runtime unit tests** — `openapi/clientruntime/client_runtime_test.go`, an internal test
package so it can reach the unexported helpers. This is where URL joining, path escaping,
value formatting, body draining, size caps, option ordering, and error construction are
actually pinned; every generated client runs this code.

**2. Golden files** — `generator/generator_test.go`. Each case reads an input fixture OpenAPI
document from `generator/fixtures/<Case>/`, runs `Generate`, and diffs (`go-cmp`) the result
against `types.gen.go` in the same directory. When adding a generator behavior, add a
`generator/fixtures/<Case>/` fixture rather than asserting inline. `generator/fixtures_test.go`
adds structural guards over the whole tree: every fixture with a spec must be referenced,
generated methods must not contain plumbing identifiers, and support files must import stdlib
only.

`clientRefFile` on a table entry is where that fixture's `client.gen.go` lives. Leaving it
unset **asserts the spec generates no client at all**, which is the right expectation for the
schema-only fixtures — it is an assertion, not a skip. When a spec has paths but its operation
resolves to no client method (no `operationId` — see `Formats`), set `clientRefFile` and keep
**no** golden file on disk; absence is the signal, and `UPDATE_GOLDEN` removes a stale file
rather than writing an empty one.

**3. Execution tests** — hand-written `client_exec_test.go` inside a fixture package, driving
the generated client against an `httptest.Server`. These are what catch wrong URLs, wrong
encodings, and wrong status dispatch; compilation alone never did. `ClientFullFeatureSet`
covers the six core operation shapes, `ClientStatusDispatch` covers 5xx/`default`/multiple-2xx,
and `ClientParamStyles` covers array queries, cookies, and value formatting.

Because these files are *not* regenerated, a behavioral regression fails a test instead of
being rubber-stamped in a golden diff. That only holds if the naming rule holds:

> Inside `generator/fixtures/<Case>/`, `*.gen.go` is generated output and is overwritten by
> `UPDATE_GOLDEN=1`. Every other file is hand-written and is never touched.

`.golangci.yaml` excludes `generator/fixtures/.*\.gen\.go$` — anchored on `.gen.go` so the
hand-written exec tests stay linted. The exclusion is needed because the client runtime is
copied whole into each generated package, so any package using only part of it reports the
rest as unused; the runtime is linted and tested at its source instead.
