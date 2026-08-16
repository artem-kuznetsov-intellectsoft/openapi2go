// Package generator converts an OpenAPI document's component schemas into
// Go struct and enum declarations.
package generator

import (
	"fmt"
	"go/format"
	"strings"

	"github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"
)

type enumDef struct {
	name        string
	description string
	values      []string
}

type fieldDef struct {
	name     string
	typ      string
	jsonTag  string
	embedded bool

	// paramName and in are set only for a Params struct field (see
	// buildParamField) — the original (JSON) parameter name and its location
	// ("path", "query", "header", "cookie") — so client.go generation can
	// place the value correctly on the outgoing request. Unset for ordinary
	// schema fields.
	paramName string
	in        string
}

type structDef struct {
	name      string
	comment   string
	fields    []fieldDef
	alias     string // if set, this schema renders as "type name = alias" instead of a struct
	errorBody string // if set, an `Error() string` method with this body is emitted after the type
}

type generator struct {
	schemas     map[string]*openapi.Schema
	parameters  map[string]*openapi.Parameter
	generated   map[string]bool
	enumIndex   map[string]*enumDef
	structIndex map[string]*structDef

	enums             []*enumDef
	structs           []*structDef
	pendingInline     []*pendingInlineStruct
	clientMethods     []*clientMethodDef
	usesDateTime      bool
	usesDate          bool
	usesOneOf         bool
	usesDiscriminated bool
}

// pendingInlineStruct is a named struct generated from an anonymous array
// item schema, queued so it renders after the struct that references it
// rather than before (the ordering used for $ref-resolved struct fields).
type pendingInlineStruct struct {
	name   string
	schema *openapi.Schema
}

// markGenerated reports whether name was already generated, recording it as
// generated if not — the guard every resolve*/register* function uses to
// memoize by name.
func (g *generator) markGenerated(name string) bool {
	if g.generated[name] {
		return true
	}
	g.generated[name] = true

	return false
}

// hasClient reports whether the spec had any operation with an operationId,
// i.e. whether a client.go is emitted at all — and therefore whether the
// generated package needs the client runtime copied in alongside it. Both
// renderClient and supportFiles ask this same question, so it lives in one
// place rather than in a separate flag that could disagree with the
// clientMethods slice it would be derived from.
func (g *generator) hasClient() bool { return len(g.clientMethods) > 0 }

// addStruct appends def to g.structs and indexes it by name.
func (g *generator) addStruct(def *structDef) {
	g.structs = append(g.structs, def)
	g.structIndex[def.name] = def
}

// Generate renders Go source declaring one struct per schema under
// components.schemas of spec, plus any enum types referenced by them. It
// also walks spec.Paths first so request bodies and 4xx/5xx responses render
// in operation order ahead of the alphabetical components.schemas pass, and
// error responses get an `Error() string` method (see registerErrorResponse).
// supportFiles returns any of openapi.SupportFiles the generated code
// depends on, with the package clause rewritten to pkgName so the caller's
// output compiles without importing this module. clientCode is client.go's
// source — a Client type with one method per operation that has an
// operationId, built from the exact type names code declares — or "" if spec
// has no such operation.
func Generate(spec *openapi.OpenAPI, pkgName string) (string, map[string]string, string, error) {
	g := &generator{
		schemas:     map[string]*openapi.Schema{},
		parameters:  map[string]*openapi.Parameter{},
		generated:   map[string]bool{},
		enumIndex:   map[string]*enumDef{},
		structIndex: map[string]*structDef{},
	}

	g.loadComponents(spec.Components)
	g.walkPaths(spec.Paths)

	for _, name := range sortedKeys(g.schemas) {
		g.resolveNamedType(name)
	}

	src := g.render(pkgName)

	formatted, err := format.Source([]byte(src))
	if err != nil {
		return src, nil, "", fmt.Errorf("formatting generated code: %w", err)
	}

	clientCode, clientSrc, err := g.formatClientCode(pkgName)
	if err != nil {
		return string(formatted), g.supportFiles(pkgName), clientSrc, fmt.Errorf("formatting generated client code: %w", err)
	}

	return string(formatted), g.supportFiles(pkgName), clientCode, nil
}

// loadComponents copies every named schema/parameter out of components into
// g.schemas/g.parameters, skipping unresolved $refs (which can't occur in a
// valid spec's own components section, but the RefOr shape allows it).
func (g *generator) loadComponents(components *openapi.Components) {
	if components == nil {
		return
	}

	for name, ref := range components.Schemas {
		if ref != nil && ref.Value != nil {
			g.schemas[name] = ref.Value
		}
	}

	for name, ref := range components.Parameters {
		if ref != nil && ref.Value != nil {
			g.parameters[name] = ref.Value
		}
	}
}

// formatClientCode renders and formats client.go's source, returning "" for
// both if g recorded no client methods. On a format error it also returns
// the unformatted source, mirroring Generate's own error path above, so a
// caller can inspect what failed to format.
func (g *generator) formatClientCode(pkgName string) (string, string, error) {
	src, err := g.renderClient(pkgName)
	if err != nil {
		return "", "", err
	}
	if src == "" {
		return "", "", nil
	}

	formattedBytes, err := format.Source([]byte(src))
	if err != nil {
		return "", src, err
	}

	return string(formattedBytes), src, nil
}

// supportFiles returns the subset of openapi.SupportFiles that the schemas
// walked by g required, with each file's package clause rewritten to pkgName
// so it compiles alongside the generated code without depending on this
// module.
func (g *generator) supportFiles(pkgName string) map[string]string {
	var names []string
	if g.usesDate || g.usesDateTime {
		names = append(names, "date.go")
	}
	if g.usesOneOf {
		names = append(names, "oneof.go")
	}
	if g.usesDiscriminated {
		names = append(names, "discriminated.go")
	}
	if g.hasClient() {
		names = append(names, "client_runtime.go")
	}

	if len(names) == 0 {
		return nil
	}

	source := openapi.SupportFiles()

	files := make(map[string]string, len(names))
	for _, name := range names {
		outName := strings.TrimSuffix(name, ".go") + ".gen.go"
		// The header matters in the consumer's tree as much as in ours: a
		// support file is hand-written here but generated output there, and
		// without it their linter has no reason to skip it.
		files[outName] = generatedFileHeader + rewritePackageClause(source[name], pkgName)
	}

	return files
}

// rewritePackageClause replaces src's package clause, and any package doc
// comment ahead of it, with a bare "package pkgName".
//
// Support files no longer all declare "package openapi" — the client runtime
// lives in its own subpackage so its Client/HTTPResponse names do not collide
// with the OpenAPI object model — so the clause is located rather than
// matched literally. The doc comment goes with it: "Package clientruntime
// holds..." is wrong once the file has been copied into a package named
// something else.
func rewritePackageClause(src, pkgName string) string {
	start := 0
	if !strings.HasPrefix(src, "package ") {
		i := strings.Index(src, "\npackage ")
		if i < 0 {
			return src
		}
		start = i + 1
	}

	end := start + strings.IndexByte(src[start:], '\n')

	return "package " + pkgName + src[end:]
}
