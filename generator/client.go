package generator

import (
	"fmt"
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
}

// clientErrorDef is one 4xx/5xx case in a client method's response switch.
type clientErrorDef struct {
	code      string
	typ       string
	hasSchema bool // false for a content-less error response (no unmarshal).
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

// renderClient renders client.go: a Client type plus one method per
// operation recorded by registerClientMethod. Returns "" if the spec had no
// such operation, so callers know not to emit a client.go file at all.
func (g *generator) renderClient(pkgName string) string {
	if len(g.clientMethods) == 0 {
		return ""
	}

	usesJSON, usesBytes, usesURL, usesIO := g.clientImports()

	var b strings.Builder

	fmt.Fprintf(&b, "package %s\n\n", pkgName)

	b.WriteString("import (\n")
	if usesBytes {
		b.WriteString("\"bytes\"\n")
	}
	b.WriteString("\"context\"\n")
	if usesJSON {
		b.WriteString("\"encoding/json\"\n")
	}
	b.WriteString("\"fmt\"\n")
	if usesIO {
		b.WriteString("\"io\"\n")
	}
	b.WriteString("\"net/http\"\n")
	if usesURL {
		b.WriteString("\"net/url\"\n")
	}
	b.WriteString(")\n\n")

	b.WriteString("type Client struct {\n")
	b.WriteString("baseURL string\n")
	b.WriteString("apiKey  string\n")
	b.WriteString("http    *http.Client\n")
	b.WriteString("}\n\n")

	b.WriteString("func NewClient(baseURL, apiKey string, httpClient *http.Client) *Client {\n")
	b.WriteString("return &Client{baseURL: baseURL, apiKey: apiKey, http: httpClient}\n")
	b.WriteString("}\n\n")

	for _, m := range g.clientMethods {
		g.renderClientMethod(&b, m)
	}

	return b.String()
}

// clientImports reports which conditionally-needed packages client.go
// requires: encoding/json (any request body, success response, or
// schema-backed error needs marshal/unmarshal), bytes (a request body is
// sent), net/url (a query parameter is built), and io (respBody is read at
// all — a success response or schema-backed error needs it).
func (g *generator) clientImports() (usesJSON, usesBytes, usesURL, usesIO bool) {
	for _, m := range g.clientMethods {
		if m.requestType != "" {
			usesBytes = true
			usesJSON = true
		}

		if m.responseType != "" {
			usesJSON = true
			usesIO = true
		}

		if m.hasSchemaError() {
			usesJSON = true
			usesIO = true
		}

		for _, f := range g.paramsFields(m.paramsType) {
			usesURL = usesURL || f.in == "query"
		}
	}

	return usesJSON, usesBytes, usesURL, usesIO
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

// renderClientMethod emits one Client method for m: build the request URL
// (substituting path parameters, appending query parameters), marshal a
// request body if there is one, send it, then switch on the status code to
// return a typed error or the unmarshaled success response.
func (g *generator) renderClientMethod(b *strings.Builder, m *clientMethodDef) {
	hasResp := m.responseType != ""
	errRet := clientErrRet(hasResp)

	g.renderMethodSignature(b, m, hasResp)
	g.renderRequestBuildAndSend(b, m, errRet, hasResp || m.hasSchemaError())
	g.renderResponseHandling(b, m, hasResp, errRet)

	b.WriteString("}\n\n")
}

// hasSchemaError reports whether any error case unmarshals respBody, i.e.
// whether respBody is read at all.
func (m *clientMethodDef) hasSchemaError() bool {
	for _, e := range m.errors {
		if e.hasSchema {
			return true
		}
	}

	return false
}

// renderMethodSignature writes the method's doc comment and its opening
// "func (c *Client) Name(...) ... {" line.
func (g *generator) renderMethodSignature(b *strings.Builder, m *clientMethodDef, hasResp bool) {
	fmt.Fprintf(b, "// %s is generated for operationId %s.\n", m.name, m.operationID)
	fmt.Fprintf(b, "// It performs a %s request against paths[%q] of the OpenAPI spec.\n",
		httpMethodLabel(m.httpMethod), m.path)

	fmt.Fprintf(b, "func (c *Client) %s(ctx context.Context", m.name)
	if m.paramsType != "" {
		fmt.Fprintf(b, ", params %s", m.paramsType)
	}
	if m.requestType != "" {
		fmt.Fprintf(b, ", req %s", m.requestType)
	}

	if hasResp {
		fmt.Fprintf(b, ") (*%s, error) {\n", m.responseType)
	} else {
		b.WriteString(") error {\n")
	}
}

// renderRequestBuildAndSend writes everything from marshaling an optional
// request body through sending the request. needsRespBody controls whether
// the response body is also read into a respBody local — skipped when
// nothing downstream would reference it, which would otherwise leave an
// unused variable.
func (g *generator) renderRequestBuildAndSend(b *strings.Builder, m *clientMethodDef, errRet func(string) string, needsRespBody bool) {
	if m.requestType != "" {
		b.WriteString("body, err := json.Marshal(req)\n")
		fmt.Fprintf(b, "if err != nil {\nreturn %s\n}\n\n", errRet("err"))
	}

	g.renderRequestURL(b, m)

	bodyExpr := "nil"
	if m.requestType != "" {
		bodyExpr = "bytes.NewReader(body)"
	}
	fmt.Fprintf(b, "httpReq, err := http.NewRequestWithContext(ctx, %s, reqURL, %s)\n", m.httpMethod, bodyExpr)
	fmt.Fprintf(b, "if err != nil {\nreturn %s\n}\n\n", errRet("err"))

	g.renderHeaderParams(b, m)
	if m.requestType != "" {
		b.WriteString("httpReq.Header.Set(\"Content-Type\", \"application/json\")\n\n")
	}

	b.WriteString("httpResp, err := c.http.Do(httpReq)\n")
	fmt.Fprintf(b, "if err != nil {\nreturn %s\n}\ndefer httpResp.Body.Close()\n\n", errRet("err"))

	if !needsRespBody {
		return
	}

	b.WriteString("respBody, err := io.ReadAll(httpResp.Body)\n")
	fmt.Fprintf(b, "if err != nil {\nreturn %s\n}\n\n", errRet("err"))
}

// renderResponseHandling writes the 4xx/5xx status-code switch, then the
// success path: unmarshal respBody into responseType and return it, or just
// return nil when the operation has no success schema.
func (g *generator) renderResponseHandling(
	b *strings.Builder,
	m *clientMethodDef,
	hasResp bool,
	errRet func(string) string,
) {
	if len(m.errors) > 0 {
		b.WriteString("switch httpResp.StatusCode {\n")
		for _, e := range m.errors {
			g.renderErrorCase(b, e, errRet)
		}
		b.WriteString("}\n\n")
	}

	if !hasResp {
		b.WriteString("return nil\n")
		return
	}

	fmt.Fprintf(b, "var resp %s\n", m.responseType)
	b.WriteString("if err := json.Unmarshal(respBody, &resp); err != nil {\nreturn nil, err\n}\n\n")
	b.WriteString("return &resp, nil\n")
}

// renderErrorCase writes one "case <code>:" arm of the response switch: a
// schema-backed error unmarshals respBody into its type first; a
// content-less one just returns its zero value.
func (g *generator) renderErrorCase(b *strings.Builder, e clientErrorDef, errRet func(string) string) {
	fmt.Fprintf(b, "case %s:\n", e.code)

	if !e.hasSchema {
		fmt.Fprintf(b, "return %s\n", errRet(e.typ+"{}"))
		return
	}

	fmt.Fprintf(b, "var errResp %s\n", e.typ)
	b.WriteString("if err := json.Unmarshal(respBody, &errResp); err != nil {\n")
	fmt.Fprintf(b, "return %s\n}\n", errRet("err"))
	fmt.Fprintf(b, "return %s\n", errRet("errResp"))
}

// renderRequestURL emits the reqURL local variable: m.path with each
// "{name}" path-parameter token substituted via fmt.Sprintf, followed by a
// query string built from any "query"-located Params field. Path parameters
// are always required (see buildParamField), so their field access is never
// nil-guarded; a "query" field may be a pointer (optional), so those are.
func (g *generator) renderRequestURL(b *strings.Builder, m *clientMethodDef) {
	format, args := formatPathTemplate(m.path)

	fmt.Fprintf(b, "reqURL := fmt.Sprintf(%q, c.baseURL", format)
	for _, a := range args {
		fmt.Fprintf(b, ", %s", a)
	}
	b.WriteString(")\n")

	var queryFields []fieldDef
	for _, f := range g.paramsFields(m.paramsType) {
		if f.in == "query" {
			queryFields = append(queryFields, f)
		}
	}

	if len(queryFields) > 0 {
		b.WriteString("q := url.Values{}\n")
		for _, f := range queryFields {
			renderOptionalValueSet(b, "q.Set", f)
		}
		b.WriteString("if len(q) > 0 {\nreqURL += \"?\" + q.Encode()\n}\n")
	}

	b.WriteString("\n")
}

// renderHeaderParams emits httpReq.Header.Set calls for any "header"-located
// Params field, after httpReq has been built.
func (g *generator) renderHeaderParams(b *strings.Builder, m *clientMethodDef) {
	for _, f := range g.paramsFields(m.paramsType) {
		if f.in == "header" {
			renderOptionalValueSet(b, "httpReq.Header.Set", f)
		}
	}
}

// renderOptionalValueSet emits a call to setterExpr(f.paramName, <value>) for
// Params field f, nil-guarding it first when f is a pointer (an optional
// query/header parameter).
func renderOptionalValueSet(b *strings.Builder, setterExpr string, f fieldDef) {
	if strings.HasPrefix(f.typ, "*") {
		fmt.Fprintf(b, "if params.%s != nil {\n%s(%q, fmt.Sprint(*params.%s))\n}\n",
			f.name, setterExpr, f.paramName, f.name)

		return
	}

	fmt.Fprintf(b, "%s(%q, fmt.Sprint(params.%s))\n", setterExpr, f.paramName, f.name)
}

// formatPathTemplate converts an OpenAPI path template into a fmt.Sprintf
// format string (with a leading "%s" for c.baseURL) plus the "params.<Field>"
// argument expressions for each "{name}" token, in the order they appear.
func formatPathTemplate(path string) (string, []string) {
	var out strings.Builder
	var args []string

	out.WriteString("%s")

	for i := 0; i < len(path); {
		if path[i] != '{' {
			out.WriteByte(path[i])
			i++

			continue
		}

		end := strings.IndexByte(path[i:], '}')
		if end == -1 {
			out.WriteByte(path[i])
			i++

			continue
		}

		name := path[i+1 : i+end]
		out.WriteString("%v")
		args = append(args, "params."+toPascalCase(name))
		i += end + 1
	}

	return out.String(), args
}

// httpMethodLabel turns an "http.Method*" constant expression into its
// lowercase HTTP method name, for the method's doc comment.
func httpMethodLabel(httpMethod string) string {
	return strings.ToLower(strings.TrimPrefix(httpMethod, "http.Method"))
}
