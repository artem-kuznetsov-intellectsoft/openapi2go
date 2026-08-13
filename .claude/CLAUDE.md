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

## Type-mapping rules

@rules/type-mapping.md

The above is the authoritative rule set for OpenAPI→Go type mapping. Treat it as binding
whenever generating or reviewing generator behavior, not just as background reading.

## Architecture

- **`openapi/schema.go`** — the OpenAPI Object model. Anything that can be either an inline
  object or a `$ref` string (schemas, parameters, responses, etc.) is wrapped in the generic
  `RefOr[T]`, which has custom `MarshalJSON`/`UnmarshalJSON` to round-trip either form.
- **`generator/generator.go`** — the actual OpenAPI→Go translation. A single
  unexported `generator` struct accumulates `structDef`/`enumDef` entries while walking schemas,
  then `render` emits source text that is passed through `go/format.Source` before being
  returned. Key behaviors baked into this code (see also the "Type-mapping rules" section above):
  - Struct fields are emitted in **alphabetical order of the JSON property name**, not spec
    declaration order (`sortedKeys`).
  - `$ref` resolution is recursive and memoized via `generator.generated`/`resolveNamedType` —
    each referenced schema is generated at most once, in the order first encountered.
  - The OpenAPI 3.0 `allOf: [{ $ref: ... }]` idiom (used to attach `nullable` to a bare `$ref`)
    is specifically unwrapped in `unwrapRef` — a single-entry `allOf` wrapping a `$ref` is
    treated as that ref, nullable if the wrapper schema says so.
  - Enum types are named after the *property*, not the schema (`toPascalCase(propName)`), and
    deduplicated by name in `enumIndex` so the same enum used on multiple properties emits once.
  - `usesTime` is a generator-wide flag: the `"time"` import is only emitted if some field
    actually became `time.Time` (string + `format: date-time`).
- **`cmd/openapi2go`** — the CLI entrypoint (`generate` subcommand). Note `reorderFlagsFirst`:
  it hoists `-o`/`-pkg` in front of the positional spec path before calling `flag.Parse`, since
  Go's `flag` package stops parsing at the first non-flag argument.
- **`cmd/parser`, `cmd/validator`, `cmd/validator-v2`** — additional command-line tools in this
  module. Read access to these directories is denied by this repo's local Claude settings
  (`.claude/settings.local.json`) — do not attempt to reopen them if a read is blocked.

## Testing pattern

Generator tests (`generator/generator_test.go`) are golden-file tests: each case reads
an input fixture OpenAPI document from `testdata/<Case>/`, runs `Generate`, and diffs
(`go-cmp`) the result against a `generated.ref.go` file in the same directory. When adding a new
generator behavior, add a new `testdata/<Case>/` fixture pair rather than asserting inline —
that's the existing convention. `testdata/` also contains scenario fixtures (arrays, maps,
polymorphism, oneOf/allOf composition, nullable fields, read/write-only, required vs optional,
generic schemas, etc.) that aren't all wired into `_test.go` yet — check before assuming a
scenario is covered.

## `api/openapi.json`

This is a large real-world spec. A hook blocks reading it directly with the `Read` tool
("forbidden to prevent context window bloat") — use `jq` via Bash instead, e.g.
`jq '.components.schemas | keys' api/openapi.json` or `jq -c '.components.schemas.Foo' api/openapi.json`.
Several files under `api/` and `openapi/references/*.md` are also read-denied by local settings — expect
this and route around it with `jq`/other tools rather than retrying the same read.
