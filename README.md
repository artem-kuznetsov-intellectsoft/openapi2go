# openapi2go

A Go code generator that converts OpenAPI 3.0.x component schemas into Go struct and
enum declarations.

`openapi/schema.go` is a hand-written struct model of the OpenAPI 3.0.x Object tree
(unmarshaled via `encoding/json`). `generator` walks `components.schemas` in that model
and emits formatted Go source. `cmd/openapi2go` wraps it in a small CLI.

## Install

```sh
go install github.com/artem-kuznetsov-intellectsoft/openapi2go/cmd/openapi2go
```

## Usage

```sh
openapi2go generate <openapi-spec-path> [-o output-dir] [-pkg name]
openapi2go version
```

- `-o` — output directory for the generated Go code (default: stdout); written as
  `types.gen.go`, plus `client.gen.go` and any support files (e.g. `date.gen.go`) as needed
- `-pkg` — package name for the generated Go code (default: `generated`)

Example:

```sh
openapi2go generate api/openapi.json -o generated -pkg generated
```

## Development

```sh
go build ./...
go vet ./...
go test ./...
golangci-lint run   # config: .golangci.yaml
gofmt -l .          # formatting check
```

Generator tests are golden-file tests: each case in `generator/fixtures/<Case>/` pairs
an input OpenAPI fixture with an expected `generated.ref.go`. See
[`.claude/CLAUDE.md`](.claude/CLAUDE.md) for architecture notes and repo-specific
conventions.
