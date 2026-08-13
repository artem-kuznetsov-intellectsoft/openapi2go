### **AI Guide: Converting OpenAPI Specifications to Go Code**

When converting OpenAPI YAML/JSON schema definitions into Go data transfer objects (DTOs), follow these strict structural mapping rules to maintain compatibility, support nullability, and enforce typings:

---

#### **1. Field Naming & Struct Tagging**
* **Case Conversion**: Convert JSON property names written in `camelCase` (e.g., `externalId`, `companyTaxId`) into `PascalCase` to declare exported fields in Go (e.g., `ExternalId`, `CompanyTaxId`).
* **JSON Tagging**: Annotate every struct field with a `json` struct tag matching the exact name and case of the property from the JSON schema (e.g., `json:"id"`, `json:"createdAt"`).

---

#### **2. Type Mapping & Representation**
* **Basic Strings**: Map default OpenAPI properties defined as `"type": "string"` (including special formats like `uuid`) to primitive Go `string` fields (e.g., `"id"` maps to `Id string`).
* **Temporal Strings (Date-Time and Date)**: Map date-time properties (such as `"createdAt"` which has `"format": "date-time"`) to a `DateTime` type, not the standard library's `time.Time` (e.g., `"createdAt"` maps to `CreatedAt DateTime`). Map date-only properties (`"format": "date"`) to `Date`. Both types embed `time.Time` internally but define their own RFC3339 `MarshalJSON`/`UnmarshalJSON`. `DateTime`/`Date` (along with `OneOf`/`Discriminated`, below) are never imported from this module: their canonical source lives in `openapi/date.go`, `openapi/oneof.go`, and `openapi/discriminated.go`, embedded verbatim via `openapi.SupportFiles()` and copied by the generator into the *output* package (with only their `package` clause rewritten) whenever a field mapping needs them — so generated code never depends on `github.com/artem-kuznetsov-intellectsoft/openapi2go`, and never imports `"time"` directly for these fields.
* **Integers**: Map `"type": "integer"` to Go `int64` by default. If the property also declares `"format": "int32"`, map it to Go `int32` instead (e.g., `"seatingCapacity"` with `"format": "int32"` maps to `SeatingCapacity int32`).
* **Required and Nullability**: `required` and `nullable` are independent axes. `nullable:true` means the field must be a pointer (e.g. `*string`) to natively support null values; `required:false` (or a missing `required` entry) only adds `,omitempty` to the JSON tag and does not by itself make the field a pointer. This pointer rule applies to scalar types (strings, numbers, booleans, enum types) and `$ref`-resolved or generated struct types only — it does not apply to `map[string]any`, slice, or `any` fields, since those are already nil-able.
* **Nullable `$ref` (OpenAPI 3.0 `allOf` idiom)**: OpenAPI 3.0.x has no native way to mark a `$ref` itself `nullable`, so specs commonly wrap it in a single-entry `allOf` alongside the sibling keyword: `"allOf": [{"$ref": "#/components/schemas/Foo"}], "nullable": true`. Treat this the same as a direct `"$ref": "#/components/schemas/Foo"` — type the field with the resolved struct name — except apply the pointer rule above if the wrapper schema sets `nullable: true`. This idiom is unrelated to multi-member `allOf` composition (below), which is used for combining schemas rather than annotating one.
* **Field Ordering**: Struct fields are emitted in alphabetical order of their JSON property name, not the declaration order from the source spec. This is because Go's `encoding/json` does not preserve object key order when unmarshaling into a map, so alphabetical order is used as the deterministic fallback.
* **Inline Objects to Named Structs**: Properties defined as an inline `"type": "object"` (without a `$ref`) that declare their own `properties` generate a named struct, named after the PascalCase property (e.g., `"compliance"` maps to `Compliance *Compliance`, generating a `Compliance` struct from those properties). This generated struct is emitted after the struct whose field references it. If the object has no `properties` of its own (a free-form object), the property instead maps to a map with string keys and any values: `map[string]any` (e.g., `"externalId"` maps to `ExternalId map[string]any`).
* **Arrays of Objects**: Properties defined as `"type": "array"` containing `"items": {"type": "object"}` map to a slice of a generated `<ArrayTypeName>Entry` struct (e.g., `"taxResidences"` with object items maps to `TaxResidences []TaxResidencesEntry`, generating a `TaxResidencesEntry` struct from those items' properties). This generated struct is emitted after the struct whose field references it. If the items schema has no `properties` of its own (a free-form object), the property instead maps to a slice of string-to-any maps: `[]map[string]any`.
* **Component References (`$ref`)**: When a property references another component schema (e.g., `"company": {"$ref": "#/components/schemas/CompanyDetailResponseDto"}`), type the field directly with the resolved PascalCase struct name (e.g., `Company CompanyDetailResponseDto`).
* **Composition (`allOf`)**: A bare `$ref` member of an `allOf` list becomes an embedded (anonymous) field of the referenced struct type — Go's idiom for composition. A non-`$ref` (inline) `allOf` member instead has its own `properties`/`required` merged directly into the enclosing schema's field set, as if declared there.
* **Composition (`oneOf`)**: A schema with exactly two `oneOf` members maps to `OneOf[A, B]`, a generic union type that marshals/unmarshals whichever of the two is set. If the schema also has a `discriminator` and no `properties`/`allOf` of its own (a pure discriminated union of two `$ref`s), it instead generates as a type alias to `Discriminated[A, B]`, which resolves the concrete type by matching a field's value to its own Go type name. Like `DateTime`/`Date` above, both types are copied into the output package rather than imported (see above).

---

#### **3. Enforcing Enums with Custom Types**
* **Type Naming**: The enum type is named after the *property*, not the schema it's declared on or referenced through — even if the enum is defined via a shared `$ref`, use the consuming property's own PascalCase name for the type. If the same property name is used with the same enum shape on more than one schema, only one type/const declaration is emitted (it is not duplicated per usage site).
* **Type Declaration**: For any string field containing an `"enum"` definition (such as `"customerType"` with values `"INDIVIDUAL"`, `"COMPANY"`), do not use raw strings. Instead, define a dedicated custom type in Go:

  ```go
  type CustomerType string
  ```

* **Constant Definitions**: Create a `const` block containing typed constants for each enum option. The naming convention for these constants is `<CustomTypeName><ENUMVALUE>` (e.g., `CustomerTypeINDIVIDUAL` and `CustomerTypeCOMPANY`):

  ```go
  const (
      CustomerTypeINDIVIDUAL CustomerType = "INDIVIDUAL"
      CustomerTypeCOMPANY    CustomerType = "COMPANY"
  )
  ```

* **Field Integration**: Define the corresponding field in the parent struct using this custom enum type rather than a standard string (e.g., `CustomerType CustomerType`).

---

---

#### **4. Paths & Error Responses**
* **Scope**: Only `paths.*.*.requestBody` and `paths.*.*.responses` are walked beyond `components.schemas`. A `requestBody` schema generates exactly like any other component reference (plain struct, no special treatment). Responses in the 4xx/5xx range are treated as errors (below); 1xx/2xx/3xx responses and the `default` response key are not — they generate as plain structs, same as `requestBody`.
* **Declaration order**: `requestBody` and 4xx/5xx response schemas are generated in path → operation → status-code order, ahead of the alphabetical `components.schemas` pass — so their declaration order in the output follows the spec's operation order rather than alphabetical schema-name order. Anything an operation doesn't touch (2xx responses, schemas unused by any operation) still falls through to, and is generated by, the alphabetical `components.schemas` pass afterward.
* **Schema-backed error response**: A 4xx/5xx response with a JSON schema (e.g. `$ref`'d to a component) generates that schema exactly like any other reference, then additionally gets an `Error() error` method (not the standard library's `Error() string` — this is a project-specific interface) on the resolved type.
* **Content-less error response**: A 4xx/5xx response with no `content` (just a `description`, e.g. a bare `"404": {"description": "Not Found"}`) synthesizes an empty struct named `Response<code>` (e.g. `Response404`), with no doc comment, and the same `Error() error` method as above. Reused/deduplicated by status code the same way a repeated enum shape is (see above): declared once even if the same code recurs across operations.
* **Method body**: Every `Error() error` method — schema-backed or content-less — always has the same body: `panic("TODO: define the output")`. The generator has no way to know what the correct output is (which field of an arbitrary schema holds the error message, or what a content-less response should report), so it stubs the body out for the developer to fill in by hand rather than guessing.

---

#### **5. Unmapped Keywords**
* **`readOnly` / `writeOnly`**: Not currently mapped to anything — no pointer, field-visibility, or separate-request/response-type behavior is generated for these keywords. A field marked `readOnly` or `writeOnly` is emitted exactly as it would be without that keyword.
