package generator

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// clientMethodDef captures everything renderClient needs to emit one Client
// method for an operation. Every type name here was already resolved by the
// rest of walkOperation (Params struct, requestBody, responses), so
// client.go can never reference a type the main generated file didn't
// declare.
type clientMethodDef struct {
	name         string // PascalCase(operationId); also the Client method name.
	operationID  string // original spelling, for the doc comment.
	httpMethod   string // Go http.Method* constant expression, e.g. "http.MethodGet".
	path         string // OpenAPI path template, e.g. "/customer/{id}".
	paramsType   string // "" if the operation has no Params struct.
	requestType  string // "" if the operation has no request body.
	responseType string // "" if the operation's success response has no schema.
	errors       []clientErrorDef

	// summary and description are the operation's OpenAPI summary/description,
	// or "" if the spec declared none. They render as their own paragraph in
	// the method's doc comment, after the generated boilerplate lines.
	summary     string
	description string

	// defaultErrorType is the Go type of the spec's "default" response, or ""
	// if it declared none. It backs every non-2xx status no explicit case
	// claimed — which is exactly what "default" means in OpenAPI.
	defaultErrorType string
}

// clientErrorDef is one 4xx/5xx case in a client method's response switch.
// Only schema-backed errors get a case at all: a content-less one needs no
// arm, because expectSuccess already turns every non-2xx into an *APIError
// carrying the status, headers, and body.
type clientErrorDef struct {
	code      string
	typ       string
	hasSchema bool
}

// ErrReservedName reports a schema whose Go name collides with a declaration
// the copied client runtime brings into the generated package.
var ErrReservedName = errors.New("schema name is reserved by the client runtime")

// ErrDuplicateMethod reports two operationIds that produce the same Go method
// name.
var ErrDuplicateMethod = errors.New("duplicate client method name")

// reservedClientNames are the identifiers the copied client runtime declares
// in the generated package. A spec whose components.schemas collides with one
// of them would produce a duplicate declaration, so generation fails instead.
//
//nolint:gochecknoglobals // a fixed lookup table, not mutable state
var reservedClientNames = map[string]bool{
	"APIError":            true,
	"Client":              true,
	"ErrDecode":           true,
	"ErrResponseTooLarge": true,
	"HTTPResponse":        true,
	"NewClient":           true,
	"RequestOption":       true,
	"WithBearerToken":     true,
	"WithHeader":          true,
}

// registerClientMethod records m as a Client method, unless the operation it
// was built from had no operationId — same rule as registerParamsStruct —
// since there'd be no name to give the method.
func (g *generator) registerClientMethod(m *clientMethodDef) {
	if m.operationID == "" {
		return
	}

	g.clientMethods = append(g.clientMethods, m)
}

// importSet collects the packages the rendered client methods actually
// referenced. Recording an import at the point the code needing it is emitted
// — rather than predicting it in a second pass over clientMethods, as an
// earlier set of usesX booleans did — means a new emission site cannot forget
// to widen a flag.
type importSet map[string]bool

func (s importSet) add(path string) { s[path] = true }

func (s importSet) render(b *strings.Builder) {
	b.WriteString("import (\n")
	for _, path := range sortedKeys(s) {
		fmt.Fprintf(b, "%q\n", path)
	}
	b.WriteString(")\n\n")
}

// renderClient renders client.go: one method per operation recorded by
// registerClientMethod. The Client type itself, its options, and all of the
// request/response plumbing come from the copied client runtime
// (client_runtime.gen.go), not from here. Returns "" if the spec had no such
// operation, so callers know not to emit a client.go file at all.
func (g *generator) renderClient(pkgName string) (string, error) {
	if !g.hasClient() {
		return "", nil
	}

	if err := g.checkClientNames(); err != nil {
		return "", err
	}

	// context and net/http appear in every method signature and request.
	imports := importSet{"context": true, "net/http": true}

	var methods strings.Builder
	for _, m := range g.clientMethods {
		g.renderClientMethod(&methods, m, imports)
	}

	var b strings.Builder

	b.WriteString(generatedFileHeader)
	fmt.Fprintf(&b, "package %s\n\n", pkgName)
	imports.render(&b)
	b.WriteString(methods.String())

	return b.String(), nil
}

// checkClientNames rejects a spec whose declared types or operation names
// would collide with each other or with the client runtime, rather than
// emitting Go that cannot compile.
func (g *generator) checkClientNames() error {
	for _, name := range sortedKeys(g.structIndex) {
		if reservedClientNames[name] {
			return fmt.Errorf("%w: rename %q in the spec", ErrReservedName, name)
		}
	}

	seen := make(map[string]string, len(g.clientMethods))
	for _, m := range g.clientMethods {
		if prev, ok := seen[m.name]; ok {
			return fmt.Errorf("%w: operationIds %q and %q both produce %q",
				ErrDuplicateMethod, prev, m.operationID, m.name)
		}
		seen[m.name] = m.operationID
	}

	return nil
}

// paramsFields returns the field list of paramsType's Params struct, or nil
// if paramsType is "".
func (g *generator) paramsFields(paramsType string) []fieldDef {
	if paramsType == "" {
		return nil
	}

	return g.structIndex[paramsType].fields
}

// clientErrRet builds the "return" expression for a method's error paths: a
// method with a success response returns (*T, error), so an error needs a
// leading "nil, "; a method with none just returns error directly.
func clientErrRet(hasResp bool) func(string) string {
	if hasResp {
		return func(expr string) string { return "nil, " + expr }
	}

	return func(expr string) string { return expr }
}

// renderClientMethod emits one Client method for m: build the query and
// header values, hand the request to the runtime's do, then dispatch on the
// status code.
func (g *generator) renderClientMethod(b *strings.Builder, m *clientMethodDef, imports importSet) {
	hasResp := m.responseType != ""
	errRet := clientErrRet(hasResp)

	g.renderMethodSignature(b, m, hasResp)
	fmt.Fprintf(b, "const op = %q\n\n", m.name)

	g.renderQueryValues(b, m, imports)
	g.renderHeaderValues(b, m)
	g.renderCookieValues(b, m)
	g.renderDoCall(b, m, errRet)
	g.renderErrorCases(b, m, errRet)
	renderSuccess(b, m, hasResp, errRet)

	b.WriteString("}\n\n")
}

// schemaErrors returns the subset of m.errors that need a case in the
// response switch. A content-less error response needs none: expectSuccess
// turns it into an *APIError carrying its status, headers, and body, which is
// strictly more than an empty marker struct could.
func (m *clientMethodDef) schemaErrors() []clientErrorDef {
	var out []clientErrorDef
	for _, e := range m.errors {
		if e.hasSchema {
			out = append(out, e)
		}
	}

	return out
}

// renderMethodSignature writes the method's doc comment and its opening
// "func (c *Client) Name(...) ... {" line.
func (g *generator) renderMethodSignature(b *strings.Builder, m *clientMethodDef, hasResp bool) {
	renderDocComment(b, methodDocLines(m))

	fmt.Fprintf(b, "func (c *Client) %s(ctx context.Context", m.name)
	if m.paramsType != "" {
		fmt.Fprintf(b, ", params %s", m.paramsType)
	}
	if m.requestType != "" {
		fmt.Fprintf(b, ", req %s", m.requestType)
	}
	b.WriteString(", opts ...RequestOption")

	if hasResp {
		fmt.Fprintf(b, ") (*%s, error) {\n", m.responseType)
	} else {
		b.WriteString(") error {\n")
	}
}

// methodDocLines builds the content of a method's doc comment as one string
// per line, with "" standing for a blank separator line — the single source
// every doc line (generated boilerplate or spec-sourced paragraph) goes
// through before renderDocComment turns it into "// "-prefixed text.
func methodDocLines(m *clientMethodDef) []string {
	lines := []string{
		fmt.Sprintf("%s is generated for operationId %s.", m.name, m.operationID),
		fmt.Sprintf("It performs a %s request against paths[%q] of the OpenAPI spec.",
			httpMethodLabel(m.httpMethod), m.path),
	}

	if codes := contentlessErrorCodes(m); len(codes) > 0 {
		noun, verb := "responses", "return"
		if len(codes) == 1 {
			noun, verb = "response", "returns"
		}
		lines = append(lines, fmt.Sprintf("The spec documents error %s %s with no content; %s an *APIError.",
			noun, strings.Join(codes, ", "), verb))
	}

	lines = append(lines, docParagraphLines(m.summary)...)
	lines = append(lines, docParagraphLines(m.description)...)

	return lines
}

// docParagraphLines turns spec-sourced text (a summary or description) into
// doc-comment lines: a leading blank separator followed by one line per
// input line, so it reads as its own paragraph rather than running on from
// the boilerplate lines above it. Returns nil if text is "".
func docParagraphLines(text string) []string {
	text = ensureTrailingPeriod(text)
	if text == "" {
		return nil
	}

	return append([]string{""}, strings.Split(text, "\n")...)
}

// renderDocComment writes lines as a "// "-prefixed doc comment; an empty
// line renders as a bare "//" rather than "// " with trailing whitespace.
func renderDocComment(b *strings.Builder, lines []string) {
	for _, line := range lines {
		if line == "" {
			b.WriteString("//\n")

			continue
		}

		fmt.Fprintf(b, "// %s\n", line)
	}
}

// ensureTrailingPeriod appends "." to text if it doesn't already end with
// one, so a spec's summary/description reads as a complete sentence in the
// generated doc comment regardless of how the spec author punctuated it.
func ensureTrailingPeriod(text string) string {
	text = strings.TrimRight(text, " \t\n")
	if text == "" || strings.HasSuffix(text, ".") {
		return text
	}

	return text + "."
}

// contentlessErrorCodes lists the documented error codes that get no case of
// their own, so the doc comment can say where they went.
func contentlessErrorCodes(m *clientMethodDef) []string {
	var codes []string
	for _, e := range m.errors {
		if !e.hasSchema {
			codes = append(codes, e.code)
		}
	}

	return codes
}

// renderQueryValues emits the url.Values build-up for any "query"-located
// Params field. A slice field uses Add per element, so an array parameter
// becomes repeated keys rather than one "[a b c]".
func (g *generator) renderQueryValues(b *strings.Builder, m *clientMethodDef, imports importSet) {
	fields := paramsIn(g.paramsFields(m.paramsType), "query")
	if len(fields) == 0 {
		return
	}

	imports.add("net/url")
	b.WriteString("query := url.Values{}\n")
	for _, f := range fields {
		renderValueSet(b, "query.Set", "query.Add", f)
	}
	b.WriteString("\n")
}

// renderHeaderValues emits the http.Header build-up for any "header"-located
// Params field.
func (g *generator) renderHeaderValues(b *strings.Builder, m *clientMethodDef) {
	fields := paramsIn(g.paramsFields(m.paramsType), "header")
	if len(fields) == 0 {
		return
	}

	b.WriteString("header := http.Header{}\n")
	for _, f := range fields {
		renderValueSet(b, "header.Set", "header.Add", f)
	}
	b.WriteString("\n")
}

// renderCookieValues emits the cookie slice for any "cookie"-located Params
// field. These were previously resolved into the Params struct and then
// silently dropped.
func (g *generator) renderCookieValues(b *strings.Builder, m *clientMethodDef) {
	fields := paramsIn(g.paramsFields(m.paramsType), "cookie")
	if len(fields) == 0 {
		return
	}

	b.WriteString("var cookies []*http.Cookie\n")
	for _, f := range fields {
		expr, guard := paramValueExpr(f)
		if guard != "" {
			fmt.Fprintf(b, "if %s {\n", guard)
		}
		fmt.Fprintf(b, "cookies = append(cookies, &http.Cookie{Name: %q, Value: formatValue(%s)})\n", f.paramName, expr)
		if guard != "" {
			b.WriteString("}\n")
		}
	}
	b.WriteString("\n")
}

// paramsIn filters fields to those bound from the given parameter location.
func paramsIn(fields []fieldDef, in string) []fieldDef {
	var out []fieldDef
	for _, f := range fields {
		if f.in == in {
			out = append(out, f)
		}
	}

	return out
}

// paramValueExpr returns the expression yielding f's value, plus the
// condition guarding it when f is an optional (pointer) parameter.
func paramValueExpr(f fieldDef) (expr, guard string) {
	if strings.HasPrefix(f.typ, "*") {
		return "*params." + f.name, "params." + f.name + " != nil"
	}

	return "params." + f.name, ""
}

// renderValueSet emits the setter call for one query or header parameter:
// nil-guarded when the field is optional, and a loop using addExpr when it is
// a slice, so repeated keys are produced rather than a Go slice rendering.
func renderValueSet(b *strings.Builder, setExpr, addExpr string, f fieldDef) {
	if strings.HasPrefix(f.typ, "[]") {
		fmt.Fprintf(b, "for _, v := range params.%s {\n%s(%q, formatValue(v))\n}\n", f.name, addExpr, f.paramName)

		return
	}

	expr, guard := paramValueExpr(f)
	if guard != "" {
		fmt.Fprintf(b, "if %s {\n%s(%q, formatValue(%s))\n}\n", guard, setExpr, f.paramName, expr)

		return
	}

	fmt.Fprintf(b, "%s(%q, formatValue(%s))\n", setExpr, f.paramName, expr)
}

// renderDoCall emits the single call into the runtime that sends the request.
func (g *generator) renderDoCall(b *strings.Builder, m *clientMethodDef, errRet func(string) string) {
	b.WriteString("resp, err := c.do(ctx, request{\n")
	b.WriteString("op: op,\n")
	fmt.Fprintf(b, "method: %s,\n", m.httpMethod)
	fmt.Fprintf(b, "path: %s,\n", pathExpr(m.path, g.paramsFields(m.paramsType)))

	if len(paramsIn(g.paramsFields(m.paramsType), "query")) > 0 {
		b.WriteString("query: query,\n")
	}
	if len(paramsIn(g.paramsFields(m.paramsType), "header")) > 0 {
		b.WriteString("header: header,\n")
	}
	if len(paramsIn(g.paramsFields(m.paramsType), "cookie")) > 0 {
		b.WriteString("cookies: cookies,\n")
	}
	if m.requestType != "" {
		b.WriteString("body: req,\n")
	}

	b.WriteString("}, opts)\n")
	fmt.Fprintf(b, "if err != nil {\nreturn %s\n}\n\n", errRet("err"))
}

// renderErrorCases emits the dispatch for documented error responses that
// have a schema. Two or more get a switch; exactly one gets an if, which
// keeps gocritic's singleCaseSwitch quiet.
func (g *generator) renderErrorCases(b *strings.Builder, m *clientMethodDef, errRet func(string) string) {
	errs := m.schemaErrors()

	switch len(errs) {
	case 0:
		return
	case 1:
		fmt.Fprintf(b, "if resp.StatusCode == %s {\n", statusConst(errs[0].code))
		fmt.Fprintf(b, "return %s\n}\n\n", errRet(decodeErrorExpr(errs[0].typ)))
	default:
		b.WriteString("switch resp.StatusCode {\n")
		for _, e := range errs {
			fmt.Fprintf(b, "case %s:\n", statusConst(e.code))
			fmt.Fprintf(b, "return %s\n", errRet(decodeErrorExpr(e.typ)))
		}
		b.WriteString("}\n\n")
	}
}

// decodeErrorExpr is the runtime call decoding one documented error payload.
func decodeErrorExpr(typ string) string {
	return "decodeError[" + typ + "](resp, op)"
}

// renderSuccess emits the 2xx gate and the success decode. The gate is what
// makes an undocumented status — a 500, or a 4xx the spec never listed — an
// error rather than a body decoded into the success type. When the spec
// declares a "default" response, the gate decodes it instead of returning a
// bare envelope.
func renderSuccess(b *strings.Builder, m *clientMethodDef, hasResp bool, errRet func(string) string) {
	gate := "resp.expectSuccess(op)"
	if m.defaultErrorType != "" {
		gate = fmt.Sprintf("expectSuccessDefault[%s](resp, op)", m.defaultErrorType)
	}

	if !hasResp {
		fmt.Fprintf(b, "return %s\n", gate)

		return
	}

	fmt.Fprintf(b, "if err := %s; err != nil {\nreturn %s\n}\n\n", gate, errRet("err"))
	fmt.Fprintf(b, "return decodeJSON[%s](resp, op)\n", m.responseType)
}

// pathExpr converts an OpenAPI path template into a Go expression building
// the request path: literal runs become quoted string constants and each
// "{name}" token becomes pathParam(params.<Field>).
//
// Building the path by concatenation rather than fmt.Sprintf is what keeps a
// literal "%" in a spec path from being read as a format verb, and routing
// every substitution through pathParam is what escapes a value containing
// "/", "?", or "#" instead of letting it retarget the request.
func pathExpr(path string, fields []fieldDef) string {
	var parts []string
	var literal strings.Builder

	flush := func() {
		if literal.Len() > 0 {
			parts = append(parts, strconv.Quote(literal.String()))
			literal.Reset()
		}
	}

	for i := 0; i < len(path); {
		if path[i] != '{' {
			literal.WriteByte(path[i])
			i++

			continue
		}

		end := strings.IndexByte(path[i:], '}')
		if end == -1 {
			literal.WriteByte(path[i])
			i++

			continue
		}

		name := path[i+1 : i+end]
		if field, ok := pathFieldFor(fields, name); ok {
			flush()
			parts = append(parts, "pathParam(params."+field+")")
		} else {
			// No parameter declares this token, so there is no field to
			// reference. Emitting params.<Name> anyway would produce code
			// that does not compile; leaving the token literal at least
			// yields a request the server can reject.
			literal.WriteString(path[i : i+end+1])
		}

		i += end + 1
	}

	flush()

	if len(parts) == 0 {
		return `""`
	}

	return strings.Join(parts, " + ")
}

// pathFieldFor finds the Params field bound to a path parameter named name,
// matching on the original parameter name that buildParamField recorded
// rather than re-deriving the Go identifier, so the two cannot disagree.
func pathFieldFor(fields []fieldDef, name string) (string, bool) {
	for _, f := range fields {
		if f.in == "path" && f.paramName == name {
			return f.name, true
		}
	}

	return "", false
}

// statusConst renders a status code as its net/http constant when one exists,
// falling back to the numeric literal for codes the standard library does not
// name.
func statusConst(code string) string {
	status, err := strconv.Atoi(code)
	if err != nil {
		return code
	}

	if name, ok := statusConstNames[status]; ok {
		return name
	}

	return code
}

// statusConstNames maps each status code the generator emits a case for to
// its net/http constant, so the generated switch reads in status names rather
// than bare magic numbers. Codes absent here fall back to the literal.
//
//nolint:gochecknoglobals // a fixed lookup table, not mutable state
var statusConstNames = map[int]string{
	http.StatusBadRequest:          "http.StatusBadRequest",
	http.StatusUnauthorized:        "http.StatusUnauthorized",
	http.StatusPaymentRequired:     "http.StatusPaymentRequired",
	http.StatusForbidden:           "http.StatusForbidden",
	http.StatusNotFound:            "http.StatusNotFound",
	http.StatusMethodNotAllowed:    "http.StatusMethodNotAllowed",
	http.StatusNotAcceptable:       "http.StatusNotAcceptable",
	http.StatusRequestTimeout:      "http.StatusRequestTimeout",
	http.StatusConflict:            "http.StatusConflict",
	http.StatusGone:                "http.StatusGone",
	http.StatusPreconditionFailed:  "http.StatusPreconditionFailed",
	http.StatusUnprocessableEntity: "http.StatusUnprocessableEntity",
	http.StatusTooManyRequests:     "http.StatusTooManyRequests",
	http.StatusInternalServerError: "http.StatusInternalServerError",
	http.StatusNotImplemented:      "http.StatusNotImplemented",
	http.StatusBadGateway:          "http.StatusBadGateway",
	http.StatusServiceUnavailable:  "http.StatusServiceUnavailable",
	http.StatusGatewayTimeout:      "http.StatusGatewayTimeout",
}

// httpMethodLabel turns an "http.Method*" constant expression into its
// lowercase HTTP method name, for the method's doc comment.
func httpMethodLabel(httpMethod string) string {
	return strings.ToLower(strings.TrimPrefix(httpMethod, "http.Method"))
}
