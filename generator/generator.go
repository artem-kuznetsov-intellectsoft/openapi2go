// Package generator converts an OpenAPI document's component schemas into
// Go struct and enum declarations.
package generator

import (
	"fmt"
	"go/format"
	"maps"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"
)

type enumDef struct {
	name        string
	description string
	values      []string
}

// goAny is the Go type used for a schema with no more specific mapping.
const goAny = "any"

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

// Generate renders Go source declaring one struct per schema under
// components.schemas of spec, plus any enum types referenced by them, as
// code. It also walks spec.Paths so request bodies and 4xx/5xx responses
// are generated in operation order (ahead of the alphabetical
// components.schemas pass) and error responses get an `Error() string`
// method — see registerErrorResponse. It also returns supportFiles: any of
// openapi.SupportFiles that code depends on (e.g. "date.go", when a
// date/date-time field was generated), keyed by filename with their package
// clause rewritten to pkgName. Callers write these alongside code's own
// output file so the generated package defines DateTime, Date, OneOf, and
// Discriminated itself rather than importing this module's openapi package.
// The fourth return value, clientCode, is client.go's source: a Client type
// with one method per operation that has an operationId (see
// registerClientMethod/renderClient), built from the exact same type names
// code declares. It is "" when spec has no such operation.
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
	src := g.renderClient(pkgName)
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
// walked by g required, with each file's package clause rewritten from
// "package openapi" to pkgName so it compiles alongside the generated code
// without depending on this module.
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

	if len(names) == 0 {
		return nil
	}

	source := openapi.SupportFiles()

	files := make(map[string]string, len(names))
	for _, name := range names {
		files[name] = strings.Replace(source[name], "package openapi\n", "package "+pkgName+"\n", 1)
	}

	return files
}

// walkPaths pre-generates the types referenced by every operation's
// requestBody and 4xx/5xx responses, in path/method/status-code order, so
// they render in that order rather than the alphabetical order the
// components.schemas pass would otherwise produce. Anything not reached
// this way (2xx responses, schemas unused by any operation) still gets
// generated by the components.schemas pass that follows.
func (g *generator) walkPaths(paths openapi.Paths) {
	for _, path := range sortedKeys(paths) {
		item := paths[path]
		if item == nil {
			continue
		}

		for _, mo := range []struct {
			httpMethod string
			op         *openapi.Operation
		}{
			{"http.MethodGet", item.Get},
			{"http.MethodPut", item.Put},
			{"http.MethodPost", item.Post},
			{"http.MethodDelete", item.Delete},
			{"http.MethodOptions", item.Options},
			{"http.MethodHead", item.Head},
			{"http.MethodPatch", item.Patch},
			{"http.MethodTrace", item.Trace},
		} {
			if mo.op != nil {
				g.walkOperation(path, mo.httpMethod, mo.op, item.Parameters)
			}
		}
	}
}

// walkOperation registers op's Params/requestBody/response types (as before)
// and additionally records a clientMethodDef from the exact type names those
// steps resolve, so client.go generation can never drift from what was
// actually declared — see registerClientMethod.
func (g *generator) walkOperation(
	path, httpMethod string,
	op *openapi.Operation,
	pathParams []*openapi.RefOr[*openapi.Parameter],
) {
	paramsType := g.registerParamsStruct(op, pathParams)
	requestType := g.registerOperationRequestBody(op)
	responseType, clientErrors := g.walkOperationResponses(op)

	g.registerClientMethod(&clientMethodDef{
		name:         toPascalCase(op.OperationID),
		operationID:  op.OperationID,
		httpMethod:   httpMethod,
		path:         path,
		paramsType:   paramsType,
		requestType:  requestType,
		responseType: responseType,
		errors:       clientErrors,
	})
}

// registerOperationRequestBody generates op's requestBody schema, exactly
// like registerInlineResponse does for a response, and returns its Go type
// name ("" if op has no requestBody or content).
func (g *generator) registerOperationRequestBody(op *openapi.Operation) string {
	if op.RequestBody == nil || op.RequestBody.Value == nil {
		return ""
	}

	schema := firstJSONSchema(op.RequestBody.Value.Content)
	if schema == nil {
		return ""
	}

	requestType, _ := g.resolveRefOrType("requestBody", schema)
	g.flushPendingInline()

	return requestType
}

// minErrorStatus/maxErrorStatus and minSuccessStatus/maxSuccessStatus bound
// the 4xx/5xx and 2xx status-code ranges walkOperationResponses classifies
// each response by.
const (
	minErrorStatus   = 400
	maxErrorStatus   = 599
	minSuccessStatus = 200
	maxSuccessStatus = 299
)

// walkOperationResponses registers every response of op (error and
// non-error alike, via registerErrorResponse/registerInlineResponse) and
// reports the pieces a client method needs: the Go type name of the first
// 2xx response with a schema (in status-code order), and every 4xx/5xx case
// for the method's response switch.
func (g *generator) walkOperationResponses(op *openapi.Operation) (string, []clientErrorDef) {
	responseType := ""
	var clientErrors []clientErrorDef

	for _, code := range sortedKeys(op.Responses) {
		resp := op.Responses[code]
		if resp == nil || resp.Value == nil {
			continue
		}

		errDef, successType := g.registerOperationResponse(code, resp.Value)
		if errDef != nil {
			clientErrors = append(clientErrors, *errDef)
			continue
		}

		if responseType == "" && successType != "" {
			responseType = successType
		}
	}

	return responseType, clientErrors
}

// registerOperationResponse registers one response code's schema (via
// registerErrorResponse for 4xx/5xx, registerInlineResponse otherwise) and
// classifies it for walkOperationResponses: a 4xx/5xx code returns a non-nil
// errDef; a 2xx code with a schema returns its Go type name as successType.
func (g *generator) registerOperationResponse(code string, resp *openapi.Response) (*clientErrorDef, string) {
	status, statusErr := strconv.Atoi(code)

	if statusErr == nil && status >= minErrorStatus && status <= maxErrorStatus {
		hasSchema := firstJSONSchema(resp.Content) != nil
		typeName := g.registerErrorResponse(code, resp)
		g.flushPendingInline()

		return &clientErrorDef{code: code, typ: typeName, hasSchema: hasSchema}, ""
	}

	typeName := g.registerInlineResponse(code, resp)
	g.flushPendingInline()

	if statusErr == nil && status >= minSuccessStatus && status <= maxSuccessStatus {
		return nil, typeName
	}

	return nil, ""
}

// registerInlineResponse generates a type for a non-error (1xx/2xx/3xx, or
// "default") response's schema when it's inline, mirroring requestBody
// handling — see type-mapping.md §4. A schema that's a $ref (or
// allOf-wrapped nullable $ref) into components.schemas is left untouched
// here: it has no inline content of its own to generate, so it's picked up
// by the later alphabetical components.schemas pass instead, preserving the
// existing declaration order for named 2xx/3xx/default schemas. Returns the
// response's Go type name in both cases (even though the $ref case doesn't
// generate it here), or "" if resp has no schema at all.
func (g *generator) registerInlineResponse(code string, resp *openapi.Response) string {
	schema := firstJSONSchema(resp.Content)
	if schema == nil {
		return ""
	}

	if name, _, ok := unwrapRef(schema); ok {
		return name
	}

	typeName, _ := g.resolveRefOrType("Response"+toPascalCase(code), schema)

	return typeName
}

// registerParamsStruct generates a "<PascalCase(operationId)>Params" struct
// for op's path/query/header/cookie parameters, merging pathParams (declared
// on the enclosing PathItem) with op.Parameters (operation-level parameters
// override a path-level one of the same name). Skipped when op has no
// operationId (there would be no name to give the struct) or no parameters
// at all. Returns the struct's name, or "" when skipped.
func (g *generator) registerParamsStruct(
	op *openapi.Operation,
	pathParams []*openapi.RefOr[*openapi.Parameter],
) string {
	if op.OperationID == "" {
		return ""
	}

	byName := map[string]*openapi.Parameter{}
	for _, ref := range pathParams {
		if p := g.resolveParameter(ref); p != nil {
			byName[p.Name] = p
		}
	}
	for _, ref := range op.Parameters {
		if p := g.resolveParameter(ref); p != nil {
			byName[p.Name] = p
		}
	}

	if len(byName) == 0 {
		return ""
	}

	name := toPascalCase(op.OperationID) + "Params"
	if g.generated[name] {
		return name
	}
	g.generated[name] = true

	fields := make([]fieldDef, 0, len(byName))
	for _, paramName := range sortedKeys(byName) {
		fields = append(fields, g.buildParamField(byName[paramName]))
	}

	def := &structDef{
		name:    name,
		comment: fmt.Sprintf("%s is generated for operationId %s.", name, op.OperationID),
		fields:  fields,
	}
	g.structs = append(g.structs, def)
	g.structIndex[name] = def

	return name
}

// resolveParameter dereferences a parameter that may be a direct $ref into
// components.parameters.
func (g *generator) resolveParameter(ref *openapi.RefOr[*openapi.Parameter]) *openapi.Parameter {
	if ref == nil {
		return nil
	}

	if ref.Ref != "" {
		return g.parameters[lastPathSegment(ref.Ref)]
	}

	return ref.Value
}

// buildParamField resolves the Go field for a single operation parameter. A
// path parameter is always treated as required per the OpenAPI spec,
// regardless of its declared "required" value. Unlike schema fields,
// parameter fields carry no json tag (they're bound from path/query/header
// values, not JSON) and optionality is expressed purely with a pointer,
// since there's no ",omitempty" tag to do that job instead.
func (g *generator) buildParamField(param *openapi.Parameter) fieldDef {
	typ := goAny
	if param.Schema != nil {
		typ, _ = g.resolveRefOrType(param.Name, param.Schema)
	}

	required := param.Required || param.In == "path"
	if !required && isPointerable(typ) {
		typ = "*" + typ
	}

	return fieldDef{name: toPascalCase(param.Name), typ: typ, paramName: param.Name, in: param.In}
}

// errorTODOBody is the Error() method body for every 4xx/5xx response type:
// the generator has no way to know what the correct output is (which field
// of an arbitrary schema holds the error message, or what a content-less
// response should report), so it always stubs the body out for the
// developer to fill in by hand.
const errorTODOBody = `panic("TODO: define the output")`

// registerErrorResponse generates the type for a 4xx/5xx response and gives
// it an `Error() string` method: a schema-backed response resolves to its
// named struct, while a response with no content synthesizes an empty
// "Response<code>" struct. Both get the same errorTODOBody. Returns the
// error type's Go name.
func (g *generator) registerErrorResponse(code string, resp *openapi.Response) string {
	schema := firstJSONSchema(resp.Content)
	if schema == nil {
		return g.registerNoContentErrorResponse(code)
	}

	typeName, _ := g.resolveRefOrType("Response"+code, schema)
	if def, ok := g.structIndex[typeName]; ok && def.errorBody == "" {
		def.errorBody = errorTODOBody
	}

	return typeName
}

func (g *generator) registerNoContentErrorResponse(code string) string {
	name := "Response" + code
	if g.generated[name] {
		return name
	}
	g.generated[name] = true

	def := &structDef{name: name, errorBody: errorTODOBody}
	g.structs = append(g.structs, def)
	g.structIndex[name] = def

	return name
}

// firstJSONSchema returns the schema of the first media type in content,
// in sorted key order, or nil if content has none.
func firstJSONSchema(content map[string]*openapi.MediaType) *openapi.RefOr[*openapi.Schema] {
	for _, key := range sortedKeys(content) {
		if mt := content[key]; mt != nil && mt.Schema != nil {
			return mt.Schema
		}
	}

	return nil
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
		def := &structDef{name: name}
		g.structs = append(g.structs, def)
		g.structIndex[name] = def

		return name
	}

	if alias, ok := g.discriminatedAlias(name, schema); ok {
		def := &structDef{
			name:    name,
			comment: fmt.Sprintf("%s is generated from components.schemas.%s.", name, name),
			alias:   alias,
		}
		g.structs = append(g.structs, def)
		g.structIndex[name] = def
		g.flushPendingInline()

		return name
	}

	fields := g.buildFields(schema)
	def := &structDef{
		name:    name,
		comment: fmt.Sprintf("%s is generated from components.schemas.%s.", name, name),
		fields:  fields,
	}
	g.structs = append(g.structs, def)
	g.structIndex[name] = def
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
	if schema.Discriminator == nil || len(schema.OneOf) != 2 || len(schema.Properties) > 0 || len(schema.AllOf) > 0 {
		return "", false
	}

	typA, _ := g.resolveRefOrType(name, schema.OneOf[0])
	typB, _ := g.resolveRefOrType(name, schema.OneOf[1])
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
		return goAny, false
	}

	nullable = schema.Nullable

	if len(schema.OneOf) == 2 {
		typA, _ := g.resolveRefOrType(propName, schema.OneOf[0])
		typB, _ := g.resolveRefOrType(propName, schema.OneOf[1])
		g.usesOneOf = true

		return fmt.Sprintf("OneOf[%s, %s]", typA, typB), nullable
	}

	// A schema whose only shape comes from a multi-member allOf (composition,
	// as opposed to the single-member allOf-wrapped-$ref nullable idiom
	// unwrapRef already handles) needs the same embed/merge treatment as a
	// named components.schemas entry — buildFields implements that. This
	// covers an inline object appearing this way as a property value, an
	// array item, or an operation's requestBody/response schema.
	if len(schema.AllOf) > 0 {
		return g.resolveInlineObjectType(toPascalCase(propName), schema), nullable
	}

	switch schema.Type {
	case "object":
		if ref, ok := schema.AdditionalProperties.(*openapi.RefOr[*openapi.Schema]); ok && ref != nil {
			valType, _ := g.resolveRefOrType(propName, ref)

			return "map[string]" + valType, nullable
		}

		if len(schema.Properties) > 0 {
			return g.resolveInlineObjectType(toPascalCase(propName), schema), nullable
		}

		return "map[string]any", nullable
	case "array":
		itemType := goAny
		if schema.Items != nil {
			itemType = g.resolveArrayItemType(propName, schema.Items)
		}

		return "[]" + itemType, nullable
	case "string":
		if len(schema.Enum) > 0 {
			enumName := toPascalCase(propName)
			g.registerEnum(enumName, schema)

			return enumName, nullable
		}

		if schema.Format == "date-time" {
			g.usesDateTime = true

			return "DateTime", nullable
		}

		if schema.Format == "date" {
			g.usesDate = true

			return "Date", nullable
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
		return goAny, nullable
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

	typ, _ := g.resolveSchemaType(propName, items.Value)

	return typ
}

// resolveInlineObjectType generates (memoized) a named struct for an
// anonymous object schema appearing as a property value or an array's items.
// Rendering is deferred until the struct that references it has been
// emitted (see flushPendingInline).
func (g *generator) resolveInlineObjectType(name string, schema *openapi.Schema) string {
	if g.generated[name] {
		return name
	}
	g.generated[name] = true

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

func (g *generator) render(pkgName string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "package %s\n\n", pkgName)

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

		switch {
		case s.alias != "":
			fmt.Fprintf(&b, "type %s = %s\n\n", s.name, s.alias)
		case len(s.fields) == 0:
			fmt.Fprintf(&b, "type %s struct{}\n\n", s.name)
		default:
			fmt.Fprintf(&b, "type %s struct {\n", s.name)
			for _, f := range s.fields {
				switch {
				case f.embedded:
					fmt.Fprintf(&b, "%s\n", f.typ)
				case f.jsonTag == "":
					fmt.Fprintf(&b, "%s %s\n", f.name, f.typ)
				default:
					fmt.Fprintf(&b, "%s %s `json:%q`\n", f.name, f.typ, f.jsonTag)
				}
			}
			b.WriteString("}\n\n")
		}

		if s.errorBody != "" {
			fmt.Fprintf(&b, "func (r %s) Error() string {\n%s\n}\n\n", s.name, s.errorBody)
		}
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
	return typ != goAny && !strings.HasPrefix(typ, "map[") && !strings.HasPrefix(typ, "[]")
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
