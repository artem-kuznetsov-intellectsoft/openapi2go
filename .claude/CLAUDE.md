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
  returned. This is where the "Type-mapping rules" above are implemented; the notes below are
  code-structure/implementation detail, not a restatement of what gets mapped to what:
  - `sortedKeys` gives deterministic struct-field and top-level-schema iteration order (the
    mapping rule this serves — alphabetical field order — is in type-mapping.md).
  - `$ref` resolution is recursive and memoized via `generator.generated`/`resolveNamedType` —
    each referenced schema is generated at most once, in the order first encountered.
  - `unwrapRef` is the single place that recognizes both a direct `$ref` and the
    `allOf`-wrapped-`$ref` nullable idiom (see type-mapping.md); it's called from both plain
    property resolution and array-item resolution, so fixes there apply to both.
  - `collectInlineProperties`/`buildFields` implement the `allOf` composition rules (embedding
    vs. property-merge) from type-mapping.md.
  - `discriminatedAlias` implements the two-member discriminated-`oneOf` → type-alias rule from
    type-mapping.md, short-circuiting normal struct generation for schemas matching that shape.
  - `enumIndex` deduplicates enum type/const declarations by the resolved type name (see
    type-mapping.md for how that name is derived) so a repeated enum shape emits once.
  - `usesDateTime`/`usesDate`/`usesOneOf`/`usesDiscriminated` are generator-wide flags — set
    whenever a field mapping requires it — that `Generate` uses to pick which of
    `openapi.SupportFiles()` (`date.go`, `oneof.go`, `discriminated.go`, embedded verbatim from
    this package via `openapi/support.go`) to return alongside the generated code, package
    clause rewritten to the output package name, so `DateTime`/`Date`/`OneOf`/`Discriminated`
    live in the output package instead of being imported from this module (see type-mapping.md
    for which formats/keywords trigger each flag). `cmd/openapi2go` writes these files next to
    `-o`'s output file; in stdout mode (no `-o`) it just lists their names on stderr.
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
that's the existing convention. Nearly every `testdata/` fixture is wired into `_test.go`
(arrays, maps, polymorphism, oneOf/allOf composition, nullable fields, required vs optional,
formats, etc.); `testdata/GenericSchema/` is the one exception left unwired — a monolithic spec
covering the same primitive/format/required scenarios plus `readOnly`/`writeOnly` (unmapped
keywords — see type-mapping.md). Check before assuming a scenario is covered.

## `api/openapi.json`

This is a large real-world spec. A hook blocks reading it directly with the `Read` tool
("forbidden to prevent context window bloat") — use `jq` via Bash instead, e.g.
`jq '.components.schemas | keys' api/openapi.json` or `jq -c '.components.schemas.Foo' api/openapi.json`.
Several files under `api/` and `openapi/references/*.md` are also read-denied by local settings — expect
this and route around it with `jq`/other tools rather than retrying the same read.
