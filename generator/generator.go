// Package generator converts an OpenAPI document's component schemas into
// Go struct and enum declarations.
package generator

import (
	"fmt"
	"go/format"
	"maps"
	"sort"
	"strings"
	"unicode"

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
}

type structDef struct {
	name    string
	comment string
	fields  []fieldDef
	alias   string // if set, this schema renders as "type name = alias" instead of a struct
}

type generator struct {
	schemas   map[string]*openapi.Schema
	generated map[string]bool
	enumIndex map[string]*enumDef

	enums             []*enumDef
	structs           []*structDef
	usesTime          bool
	usesDate          bool
	usesOneOf         bool
	usesDiscriminated bool
}

// Generate renders Go source declaring one struct per schema under
// components.schemas of spec, plus any enum types referenced by them.
func Generate(spec *openapi.OpenAPI, pkgName string) (string, error) {
	g := &generator{
		schemas:   map[string]*openapi.Schema{},
		generated: map[string]bool{},
		enumIndex: map[string]*enumDef{},
	}

	if spec.Components != nil {
		for name, ref := range spec.Components.Schemas {
			if ref != nil && ref.Value != nil {
				g.schemas[name] = ref.Value
			}
		}
	}

	for _, name := range sortedKeys(g.schemas) {
		g.resolveNamedType(name)
	}

	src := g.render(pkgName)

	formatted, err := format.Source([]byte(src))
	if err != nil {
		return src, fmt.Errorf("formatting generated code: %w", err)
	}

	return string(formatted), nil
}

// resolveNamedType returns the Go type name for a components.schemas entry,
// generating its struct declaration on first use. Schemas that cannot be
// resolved (referenced but not defined in the given spec) become an empty
// struct stub.
func (g *generator) resolveNamedType(name string) string {
	if g.generated[name] {
		return name
	}
	g.generated[name] = true

	schema, ok := g.schemas[name]
	if !ok {
		g.structs = append(g.structs, &structDef{name: name})
		return name
	}

	if alias, ok := g.discriminatedAlias(name, schema); ok {
		g.structs = append(g.structs, &structDef{
			name:    name,
			comment: fmt.Sprintf("%s is generated from components.schemas.%s.", name, name),
			alias:   alias,
		})

		return name
	}

	fields := g.buildFields(schema)
	g.structs = append(g.structs, &structDef{
		name:    name,
		comment: fmt.Sprintf("%s is generated from components.schemas.%s.", name, name),
		fields:  fields,
	})

	return name
}

// discriminatedAlias reports whether schema is a pure discriminated union —
// a oneOf of exactly two $ref members with an adjacent discriminator object
// and no properties/allOf of its own — and if so, returns the Go type
// expression it should alias to (openapi.Discriminated[A, B]).
func (g *generator) discriminatedAlias(name string, schema *openapi.Schema) (string, bool) {
	if schema.Discriminator == nil || len(schema.OneOf) != 2 || len(schema.Properties) > 0 || len(schema.AllOf) > 0 {
		return "", false
	}

	typA, _ := g.resolveRefOrType(name, schema.OneOf[0])
	typB, _ := g.resolveRefOrType(name, schema.OneOf[1])
	g.usesDiscriminated = true

	return fmt.Sprintf("openapi.Discriminated[%s, %s]", typA, typB), true
}

// fieldCollector accumulates the pieces buildFields needs while walking a
// schema's allOf members: embedded structs (from bare $ref members) plus
// the merged property/required sets contributed by inline members.
type fieldCollector struct {
	embedded []fieldDef
	props    map[string]*openapi.RefOr[*openapi.Schema]
	required map[string]bool
}

// buildFields builds the field list for schema, mapping each allOf member
// that is a bare $ref to an embedded (anonymous) struct field — the Go
// idiom for OpenAPI composition — and merging every inline allOf member's
// properties directly into this struct's own property set.
func (g *generator) buildFields(schema *openapi.Schema) []fieldDef {
	c := &fieldCollector{
		props:    map[string]*openapi.RefOr[*openapi.Schema]{},
		required: map[string]bool{},
	}
	g.collectInlineProperties(schema, c)

	fields := c.embedded
	for _, propName := range sortedKeys(c.props) {
		fields = append(fields, g.buildField(propName, c.props[propName], c.required[propName]))
	}

	return fields
}

func (g *generator) collectInlineProperties(schema *openapi.Schema, c *fieldCollector) {
	for _, member := range schema.AllOf {
		if member.Ref != "" {
			typeName := g.resolveNamedType(lastPathSegment(member.Ref))
			c.embedded = append(c.embedded, fieldDef{name: typeName, typ: typeName, embedded: true})
			continue
		}

		if member.Value != nil {
			g.collectInlineProperties(member.Value, c)
		}
	}

	maps.Copy(c.props, schema.Properties)
	maps.Copy(c.required, toSet(schema.Required))
}

func (g *generator) buildField(propName string, ref *openapi.RefOr[*openapi.Schema], required bool) fieldDef {
	goName := toPascalCase(propName)

	typ, nullable := g.resolveRefOrType(propName, ref)
	if nullable && isPointerable(typ) {
		typ = "*" + typ
	}

	tag := propName
	if !required {
		tag += ",omitempty"
	}

	return fieldDef{name: goName, typ: typ, jsonTag: tag}
}

// resolveRefOrType resolves the Go type for a property or array-item value
// that may be a direct $ref, an allOf-wrapped nullable $ref, or an inline
// schema.
func (g *generator) resolveRefOrType(propName string, ref *openapi.RefOr[*openapi.Schema]) (typ string, nullable bool) {
	if refName, refNullable, ok := unwrapRef(ref); ok {
		return g.resolveNamedType(refName), refNullable
	}

	return g.resolveSchemaType(propName, ref.Value)
}

func (g *generator) resolveSchemaType(propName string, schema *openapi.Schema) (typ string, nullable bool) {
	if schema == nil {
		return "any", false
	}

	nullable = schema.Nullable

	if len(schema.OneOf) == 2 {
		typA, _ := g.resolveRefOrType(propName, schema.OneOf[0])
		typB, _ := g.resolveRefOrType(propName, schema.OneOf[1])
		g.usesOneOf = true

		return fmt.Sprintf("openapi.OneOf[%s, %s]", typA, typB), nullable
	}

	switch schema.Type {
	case "object":
		if ref, ok := schema.AdditionalProperties.(*openapi.RefOr[*openapi.Schema]); ok && ref != nil {
			valType, _ := g.resolveRefOrType(propName, ref)

			return "map[string]" + valType, nullable
		}

		return "map[string]any", nullable
	case "array":
		itemType := "any"
		if schema.Items != nil {
			itemType, _ = g.resolveRefOrType(propName, schema.Items)
		}

		return "[]" + itemType, nullable
	case "string":
		if len(schema.Enum) > 0 {
			enumName := toPascalCase(propName)
			g.registerEnum(enumName, schema)

			return enumName, nullable
		}

		if schema.Format == "date-time" {
			g.usesTime = true

			return "time.Time", nullable
		}

		if schema.Format == "date" {
			g.usesDate = true

			return "openapi.Date", nullable
		}

		if schema.Format == "byte" {
			return "[]byte", nullable
		}

		return "string", nullable
	case "integer":
		if schema.Format == "int32" {
			return "int32", nullable
		}

		return "int64", nullable
	case "number":
		if schema.Format == "float" {
			return "float32", nullable
		}

		return "float64", nullable
	case "boolean":
		return "bool", nullable
	default:
		return "any", nullable
	}
}

func (g *generator) registerEnum(name string, schema *openapi.Schema) {
	if _, ok := g.enumIndex[name]; ok {
		return
	}

	values := make([]string, len(schema.Enum))
	for i, v := range schema.Enum {
		values[i] = fmt.Sprint(v)
	}

	e := &enumDef{name: name, description: schema.Description, values: values}
	g.enumIndex[name] = e
	g.enums = append(g.enums, e)
}

func (g *generator) render(pkgName string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "package %s\n\n", pkgName)

	var imports []string
	if g.usesTime {
		imports = append(imports, `"time"`)
	}
	if g.usesOneOf || g.usesDiscriminated || g.usesDate {
		imports = append(imports, `"github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"`)
	}

	switch len(imports) {
	case 0:
	case 1:
		fmt.Fprintf(&b, "import %s\n\n", imports[0])
	default:
		b.WriteString("import (\n")
		for _, imp := range imports {
			fmt.Fprintf(&b, "%s\n", imp)
		}
		b.WriteString(")\n\n")
	}

	for _, e := range g.enums {
		if sentence := describeSentence(e.description); sentence != "" {
			fmt.Fprintf(&b, "// %s represents the %s.\n", e.name, sentence)
		}

		fmt.Fprintf(&b, "type %s string\n\n", e.name)

		b.WriteString("const (\n")
		for _, v := range e.values {
			fmt.Fprintf(&b, "%s%s %s = %q\n", e.name, capitalizeFirst(v), e.name, v)
		}
		b.WriteString(")\n\n")
	}

	for _, s := range g.structs {
		if s.comment != "" {
			fmt.Fprintf(&b, "// %s\n", s.comment)
		}

		if s.alias != "" {
			fmt.Fprintf(&b, "type %s = %s\n\n", s.name, s.alias)
			continue
		}

		if len(s.fields) == 0 {
			fmt.Fprintf(&b, "type %s struct{}\n\n", s.name)
			continue
		}

		fmt.Fprintf(&b, "type %s struct {\n", s.name)
		for _, f := range s.fields {
			if f.embedded {
				fmt.Fprintf(&b, "%s\n", f.typ)
				continue
			}

			fmt.Fprintf(&b, "%s %s `json:%q`\n", f.name, f.typ, f.jsonTag)
		}
		b.WriteString("}\n\n")
	}

	return b.String()
}

// unwrapRef resolves a property value to a referenced schema name, handling
// both a direct $ref and the common allOf-wrapped-single-$ref pattern used
// to attach `nullable` to a reference in OpenAPI 3.0.
func unwrapRef(ref *openapi.RefOr[*openapi.Schema]) (name string, nullable bool, ok bool) {
	if ref.Ref != "" {
		return lastPathSegment(ref.Ref), false, true
	}

	schema := ref.Value
	if schema == nil {
		return "", false, false
	}

	if len(schema.AllOf) == 1 && schema.AllOf[0].Ref != "" {
		return lastPathSegment(schema.AllOf[0].Ref), schema.Nullable, true
	}

	return "", false, false
}

func lastPathSegment(ref string) string {
	if i := strings.LastIndex(ref, "/"); i >= 0 {
		return ref[i+1:]
	}

	return ref
}

func isPointerable(typ string) bool {
	return typ != "any" && !strings.HasPrefix(typ, "map[") && !strings.HasPrefix(typ, "[]")
}

// toPascalCase converts a camelCase, snake_case, or kebab-case JSON property
// name into an exported Go identifier.
func toPascalCase(s string) string {
	var b strings.Builder

	upperNext := true
	for _, r := range s {
		if r == '_' || r == '-' || r == ' ' {
			upperNext = true
			continue
		}

		if upperNext {
			b.WriteRune(unicode.ToUpper(r))
			upperNext = false
		} else {
			b.WriteRune(r)
		}
	}

	return b.String()
}

// capitalizeFirst upper-cases the first rune of s, leaving the rest as-is —
// used to turn enum values like "aggressive" into valid, exported const
// name suffixes without disturbing already-uppercase values like
// "PENDING_VERIFICATION".
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}

	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])

	return string(r)
}

// describeSentence lowercases the first letter of a description and strips
// any trailing period, for embedding into a "X represents the <desc>."
// doc comment.
func describeSentence(desc string) string {
	d := strings.TrimSuffix(strings.TrimSpace(desc), ".")
	if d == "" {
		return ""
	}

	return strings.ToLower(d[:1]) + d[1:]
}

func toSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, item := range items {
		m[item] = true
	}

	return m
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}
