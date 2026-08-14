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
```

Run the generator CLI directly:

```sh
go run ./cmd/openapi2go generate <openapi-spec-path> [-o output.go] [-pkg name]
```

## Codebase navigation

See @.claude/rules/LSP.md for LSP-based navigation guidance (use LSP tools instead of
grep/Explore where possible).

## Architecture

- **`openapi/`** — the OpenAPI Object model (`schema.go`), used to unmarshal an OpenAPI 3.0.x
  document into Go structs, plus small runtime support types (`date.go`, `oneof.go`,
  `discriminated.go`) that generated code can copy for itself, and a reference client example
  (`client_example.go`).
- **`generator/`** — the OpenAPI→Go translation. `generator.go` handles struct/enum generation
  from `components.schemas`; `client.go` handles generation of a `Client` type with one method
  per operation from `spec.Paths`. `fixtures/` holds the golden-file test fixtures (see Testing
  pattern below).
- **`cmd/openapi2go`** — the CLI entrypoint (`generate` subcommand).
- **`make/`** — Makefile includes (`lint.mk`), pulled in by the root `Makefile`.

## Testing pattern

Generator tests (`generator/generator_test.go`) are golden-file tests: each case reads
an input fixture OpenAPI document from `fixtures/<Case>/`, runs `Generate`, and diffs
(`go-cmp`) the result against a `generated.ref.go` file in the same directory. When adding a new
generator behavior, add a new `fixtures/<Case>/` fixture pair rather than asserting inline —
that's the existing convention.

Client-generation coverage is wired up per-fixture via an opt-in `clientRefFile` field on the
test table entry, separate from `refFile` — most fixtures leave it unset. When a fixture's
operation resolves to no client method at all, the convention is to have **no** golden file on
disk rather than an empty one (`UPDATE_GOLDEN` removes a stale file instead of writing an empty
one).
