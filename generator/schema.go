package generator

import (
	"fmt"
	"maps"

	"github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"
)

// goAny is the Go type used for a schema with no more specific mapping.
const goAny = "any"

// oneOfMemberCount is the exact number of oneOf members that maps to
// OneOf[A, B] (or its Discriminated[A, B] alias).
const oneOfMemberCount = 2

// resolveNamedType returns the Go type name for a components.schemas entry,
// generating its struct declaration on first use. Schemas that cannot be
// resolved (referenced but not defined in the given spec) become an empty
// struct stub.
func (g *generator) resolveNamedType(name string) string {
	if g.markGenerated(name) {
		return name
	}

	schema, ok := g.schemas[name]
	if !ok {
		g.addStruct(&structDef{name: name})

		return name
	}

	if alias, ok := g.discriminatedAlias(name, schema); ok {
		g.addStruct(&structDef{
			name:    name,
			comment: fmt.Sprintf("%s is generated from components.schemas.%s.", name, name),
			alias:   alias,
		})
		g.flushPendingInline()

		return name
	}

	g.addStruct(&structDef{
		name:    name,
		comment: fmt.Sprintf("%s is generated from components.schemas.%s.", name, name),
		fields:  g.buildFields(schema),
	})
	g.flushPendingInline()

	return name
}

// flushPendingInline renders every struct queued by resolveInlineObjectType,
// in discovery order, after the struct that referenced them.
func (g *generator) flushPendingInline() {
	for len(g.pendingInline) > 0 {
		p := g.pendingInline[0]
		g.pendingInline = g.pendingInline[1:]
		g.structs = append(g.structs, &structDef{name: p.name, fields: g.buildFields(p.schema)})
	}
}

// discriminatedAlias reports whether schema is a pure discriminated union —
// a oneOf of exactly two $ref members with an adjacent discriminator object
// and no properties/allOf of its own — and if so, returns the Go type
// expression it should alias to (openapi.Discriminated[A, B]).
func (g *generator) discriminatedAlias(name string, schema *openapi.Schema) (string, bool) {
	if schema.Discriminator == nil || len(schema.OneOf) != oneOfMemberCount || len(schema.Properties) > 0 || len(schema.AllOf) > 0 {
		return "", false
	}

	typA := g.resolveRefOrType(name, schema.OneOf[0]).typ
	typB := g.resolveRefOrType(name, schema.OneOf[1]).typ
	g.usesDiscriminated = true

	return fmt.Sprintf("Discriminated[%s, %s]", typA, typB), true
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

	rt := g.resolveRefOrType(propName, ref)
	typ := rt.typ
	if rt.nullable && isPointerable(typ) {
		typ = "*" + typ
	}

	tag := propName
	switch {
	case required:
		// present tag with no suffix
	case !rt.nullable && rt.isStruct:
		// encoding/json's omitempty never considers a struct value empty, so a
		// non-pointer struct field (DateTime, Date, OneOf, a $ref/generated
		// struct) needs omitzero's zero-value check instead.
		tag += ",omitzero"
	default:
		tag += ",omitempty"
	}

	return fieldDef{name: goName, typ: typ, jsonTag: tag}
}

// resolvedType is what resolving a property/item/parameter schema to a Go
// type produces: the type expression itself, whether it should be treated as
// nullable (wrapped in a pointer), and whether it's a struct Kind value (as
// opposed to scalar/enum/map/slice/any) — buildField needs isStruct to pick
// between the ",omitzero" and ",omitempty" JSON tag suffix.
type resolvedType struct {
	typ      string
	nullable bool
	isStruct bool
}

// resolveRefOrType resolves the Go type for a property or array-item value
// that may be a direct $ref, an allOf-wrapped nullable $ref, or an inline
// schema.
func (g *generator) resolveRefOrType(propName string, ref *openapi.RefOr[*openapi.Schema]) resolvedType {
	if refName, refNullable, ok := unwrapRef(ref); ok {
		return resolvedType{typ: g.resolveNamedType(refName), nullable: refNullable, isStruct: true}
	}

	return g.resolveSchemaType(propName, ref.Value)
}

func (g *generator) resolveSchemaType(propName string, schema *openapi.Schema) resolvedType {
	if schema == nil {
		return resolvedType{typ: goAny}
	}

	nullable := schema.Nullable

	if len(schema.OneOf) == oneOfMemberCount {
		typA := g.resolveRefOrType(propName, schema.OneOf[0]).typ
		typB := g.resolveRefOrType(propName, schema.OneOf[1]).typ
		g.usesOneOf = true

		return resolvedType{typ: fmt.Sprintf("OneOf[%s, %s]", typA, typB), nullable: nullable, isStruct: true}
	}

	// A schema whose only shape comes from a multi-member allOf (composition,
	// as opposed to the single-member allOf-wrapped-$ref nullable idiom
	// unwrapRef already handles) needs the same embed/merge treatment as a
	// named components.schemas entry — buildFields implements that. This
	// covers an inline object appearing this way as a property value, an
	// array item, or an operation's requestBody/response schema.
	if len(schema.AllOf) > 0 {
		return resolvedType{typ: g.resolveInlineObjectType(toPascalCase(propName), schema), nullable: nullable, isStruct: true}
	}

	rt := g.resolveScalarSchemaType(propName, schema)
	rt.nullable = nullable

	return rt
}

func (g *generator) resolveScalarSchemaType(propName string, schema *openapi.Schema) resolvedType {
	switch schema.Type {
	case "object":
		return g.resolveObjectSchemaType(propName, schema)
	case "array":
		itemType := goAny
		if schema.Items != nil {
			itemType = g.resolveArrayItemType(propName, schema.Items)
		}

		return resolvedType{typ: "[]" + itemType}
	case "string":
		return g.resolveStringSchemaType(propName, schema)
	case schemaTypeInteger, "number":
		return resolvedType{typ: resolveNumericSchemaType(schema)}
	case "boolean":
		return resolvedType{typ: "bool"}
	default:
		return resolvedType{typ: goAny}
	}
}

// schemaTypeInteger is the OpenAPI schema.Type value for an integer, used by
// both resolveScalarSchemaType's switch and resolveNumericSchemaType's.
const schemaTypeInteger = "integer"

// resolveNumericSchemaType resolves the "integer"/"number" case of
// resolveScalarSchemaType by its declared format.
func resolveNumericSchemaType(schema *openapi.Schema) string {
	switch {
	case schema.Type == schemaTypeInteger && schema.Format == "int32":
		return "int32"
	case schema.Type == schemaTypeInteger:
		return "int64"
	case schema.Format == "float":
		return "float32"
	default:
		return "float64"
	}
}

func (g *generator) resolveObjectSchemaType(propName string, schema *openapi.Schema) resolvedType {
	if ref, ok := schema.AdditionalProperties.(*openapi.RefOr[*openapi.Schema]); ok && ref != nil {
		valType := g.resolveRefOrType(propName, ref).typ

		return resolvedType{typ: "map[string]" + valType}
	}

	if len(schema.Properties) > 0 {
		return resolvedType{typ: g.resolveInlineObjectType(toPascalCase(propName), schema), isStruct: true}
	}

	return resolvedType{typ: "map[string]any"}
}

func (g *generator) resolveStringSchemaType(propName string, schema *openapi.Schema) resolvedType {
	if len(schema.Enum) > 0 {
		enumName := toPascalCase(propName)
		g.registerEnum(enumName, schema)

		return resolvedType{typ: enumName}
	}

	switch schema.Format {
	case "date-time":
		g.usesDateTime = true
		return resolvedType{typ: "DateTime", isStruct: true}
	case "date":
		g.usesDate = true
		return resolvedType{typ: "Date", isStruct: true}
	case "byte":
		return resolvedType{typ: "[]byte"}
	default:
		return resolvedType{typ: "string"}
	}
}

// resolveArrayItemType resolves the Go type for an array's items schema. A
// $ref (or allOf-wrapped nullable $ref) resolves to the referenced named
// type; an inline object schema with its own properties gets a named
// "<ArrayTypeName>Entry" struct generated on demand (see
// resolveInlineObjectType); anything else falls back to resolveSchemaType.
func (g *generator) resolveArrayItemType(propName string, items *openapi.RefOr[*openapi.Schema]) string {
	if refName, _, ok := unwrapRef(items); ok {
		return g.resolveNamedType(refName)
	}

	if items.Value != nil && items.Value.Type == "object" && len(items.Value.Properties) > 0 {
		return g.resolveInlineObjectType(toPascalCase(propName)+"Entry", items.Value)
	}

	return g.resolveSchemaType(propName, items.Value).typ
}

// resolveInlineObjectType generates (memoized) a named struct for an
// anonymous object schema appearing as a property value or an array's items.
// Rendering is deferred until the struct that references it has been
// emitted (see flushPendingInline).
func (g *generator) resolveInlineObjectType(name string, schema *openapi.Schema) string {
	if g.markGenerated(name) {
		return name
	}

	g.pendingInline = append(g.pendingInline, &pendingInlineStruct{name: name, schema: schema})

	return name
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
