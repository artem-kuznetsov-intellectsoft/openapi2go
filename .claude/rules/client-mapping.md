### **AI Guide: Generating a Go Client from OpenAPI Paths**

These rules govern `client.go` — a `Client` type with one method per operation, generated
from `spec.Paths` alongside (but separately from) the struct/enum generation described in
type-mapping.md. Every type name referenced here (Params struct, request/response/error
types) is reused verbatim from the resolution that generator.go's `walkOperation` already
performs for the main generated file — client generation never re-derives a name
independently, so `client.go` cannot drift out of sync with what was actually declared.

---

#### **1. Scope**
* **One method per operationId**: A `Client` method is generated for an operation only if it
  has a non-empty `operationId` — same skip rule as the `<OperationId>Params` struct
  (type-mapping.md §6). An operation with no `operationId` gets no method at all.
* **No client.go when there's nothing to generate**: If no operation in the entire spec has an
  `operationId`, `client.go` is not generated at all — not even an empty `Client` struct.
* **Method naming**: `PascalCase(operationId)` (e.g. `get-user-by-id` → `GetUserById`).

---

#### **2. Method Signature**
* **Fixed leading parameter**: every method's first parameter is always `ctx context.Context`.
* **`params <OperationId>Params`**: present only if that operation actually has a Params
  struct (type-mapping.md §6). The type name is exactly `<OperationId>Params` — no
  independent resolution happens here.
* **`req <RequestType>`**: present only if the operation has a `requestBody` with a JSON
  schema. `<RequestType>` is exactly the name type-mapping.md §4 already generated for it — a
  `$ref`'d schema's own name, or `RequestBody` for an inline one.
* **Return type**: `(*<ResponseType>, error)` when the first 2xx response (in status-code
  order) has a JSON schema; `<ResponseType>` is again reused verbatim from type-mapping.md §4
  (`$ref`'d name, or `Response<code>` for inline). If no 2xx response has a schema, the method
  returns a bare `error` instead, and every internal error return drops the `nil, ` prefix to
  match.

---

#### **3. Request Building**
* **Path parameters**: every `{name}` token in the OpenAPI path template is substituted with
  `params.<PascalCase(name)>` via `fmt.Sprintf`, in the order the tokens appear in the path — not
  in the Params struct's alphabetical field order. A path parameter is always required
  (type-mapping.md §6), so this access is never nil-guarded.
* **Query parameters**: a Params field whose original parameter had `"in": "query"` is set on
  the request via `net/url.Values` (`q.Set(paramName, fmt.Sprint(value))`), then appended to
  the URL as `?`+`q.Encode()` if any were set.
* **Header parameters**: a Params field with `"in": "header"` is set via
  `httpReq.Header.Set(paramName, fmt.Sprint(value))`, after the request is constructed.
* **Pointer (optional) params are nil-guarded**: a query/header field typed as a pointer (the
  non-required case, per type-mapping.md §6's pointer rule) is wrapped in
  `if params.X != nil { ... }` before use; a non-pointer (required) field is used directly, no
  guard.
* **Cookie parameters are not wired up.** A `"in": "cookie"` parameter still gets its Params
  struct field (type-mapping.md §6 doesn't distinguish cookie from the other locations), but
  no client method currently sets it on the outgoing request. This is a deliberate, currently
  untested gap — no fixture defines a cookie parameter yet — not an oversight to silently
  "fix" without a fixture to validate against.
* **Request body**: when the operation has a requestBody, `req` is JSON-marshaled and sent as
  the request body, and `Content-Type: application/json` is set. With no requestBody, the
  request body is `nil` and no Content-Type header is set.

---

#### **4. Response Handling**
* **Error switch**: every 4xx/5xx response the operation declares becomes one
  `case <code>:` arm of a `switch httpResp.StatusCode`, in status-code order, reusing the exact
  type name and `Error()` wiring type-mapping.md §4 already generated. A schema-backed error
  unmarshals the response body into that type and returns it as the method's `error` value; a
  content-less error just returns its zero value (`Response<code>{}`) with no unmarshal. If the
  operation has no 4xx/5xx responses at all, no switch statement is emitted.
* **Success response**: the body is unmarshaled into `<ResponseType>` (§2 above) and returned
  as `&resp, nil`. If the operation has no typed success response, the method returns `nil` (or
  the marshal/request/transport error, on failure) with no data value.

---

#### **5. Auth is intentionally unimplemented**
* `Client` carries an `apiKey` field (matching `openapi/client_example.go`'s shape), but no
  method currently injects it into the request — no fixture yet defines
  `components.securitySchemes`, so there is no reliable way to resolve an `apiKey` security
  scheme to a concrete header/query parameter name. Treat the unused `apiKey` field as a
  deliberate placeholder pending such a fixture, not a bug to silently wire up with a guessed
  header name.
