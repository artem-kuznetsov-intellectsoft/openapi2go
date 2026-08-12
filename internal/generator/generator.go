// Package generator converts an OpenAPI document's component schemas into
// Go struct and enum declarations.
package generator

import (
	"fmt"
	"go/format"
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
	name    string
	typ     string
	jsonTag string
}

type structDef struct {
	name    string
	comment string
	fields  []fieldDef
}

type generator struct {
	schemas   map[string]*openapi.Schema
	generated map[string]bool
	enumIndex map[string]*enumDef

	enums   []*enumDef
	structs []*structDef
	usesTime bool
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

	fields := g.buildFields(schema)
	g.structs = append(g.structs, &structDef{
		name:    name,
		comment: fmt.Sprintf("%s is generated from components.schemas.%s.", name, name),
		fields:  fields,
	})

	return name
}

func (g *generator) buildFields(schema *openapi.Schema) []fieldDef {
	required := toSet(schema.Required)

	fields := make([]fieldDef, 0, len(schema.Properties))
	for _, propName := range sortedKeys(schema.Properties) {
		fields = append(fields, g.buildField(propName, schema.Properties[propName], required[propName]))
	}

	return fields
}

func (g *generator) buildField(propName string, ref *openapi.RefOr[*openapi.Schema], required bool) fieldDef {
	goName := toPascalCase(propName)

	typ, nullable := g.resolveRefOrType(propName, ref)
	if (nullable || !required) && isPointerable(typ) {
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

	switch schema.Type {
	case "object":
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

		return "string", nullable
	case "integer":
		return "int64", nullable
	case "number":
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

	if g.usesTime {
		b.WriteString("import \"time\"\n\n")
	}

	for _, e := range g.enums {
		if sentence := describeSentence(e.description); sentence != "" {
			fmt.Fprintf(&b, "// %s represents the %s.\n", e.name, sentence)
		}

		fmt.Fprintf(&b, "type %s string\n\n", e.name)

		b.WriteString("const (\n")
		for _, v := range e.values {
			fmt.Fprintf(&b, "%s%s %s = %q\n", e.name, v, e.name, v)
		}
		b.WriteString(")\n\n")
	}

	for _, s := range g.structs {
		if s.comment != "" {
			fmt.Fprintf(&b, "// %s\n", s.comment)
		}

		if len(s.fields) == 0 {
			fmt.Fprintf(&b, "type %s struct{}\n\n", s.name)
			continue
		}

		fmt.Fprintf(&b, "type %s struct {\n", s.name)
		for _, f := range s.fields {
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
