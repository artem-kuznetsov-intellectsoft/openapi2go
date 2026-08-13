### **AI Guide: Converting OpenAPI Specifications to Go Code**

When converting OpenAPI YAML/JSON schema definitions into Go data transfer objects (DTOs), follow these strict structural mapping rules to maintain compatibility, support nullability, and enforce typings:

---

#### **1. Field Naming & Struct Tagging**
* **Case Conversion**: Convert JSON property names written in `camelCase` (e.g., `externalId`, `companyTaxId`) into `PascalCase` to declare exported fields in Go (e.g., `ExternalId`, `CompanyTaxId`).
* **JSON Tagging**: Annotate every struct field with a `json` struct tag matching the exact name and case of the property from the JSON schema (e.g., `json:"id"`, `json:"createdAt"`).

---

#### **2. Type Mapping & Representation**
* **Basic Strings**: Map default OpenAPI properties defined as `"type": "string"` (including special formats like `uuid`) to primitive Go `string` fields (e.g., `"id"` maps to `Id string`).
* **Temporal Strings (Date-Time)**: Map date-time properties (such as `"createdAt"` which has `"format": "date-time"`) to Go's `time.Time` (e.g., `"createdAt"` maps to `CreatedAt time.Time`). Any file that generates such a field must import `"time"`.
* **Integers**: Map `"type": "integer"` to Go `int64` by default. If the property also declares `"format": "int32"`, map it to Go `int32` instead (e.g., `"seatingCapacity"` with `"format": "int32"` maps to `SeatingCapacity int32`).
* **Required and Nullability**: `required` and `nullable` are independent axes. `nullable:true` means the field must be a pointer (e.g. `*string`) to natively support null values; `required:false` (or a missing `required` entry) only adds `,omitempty` to the JSON tag and does not by itself make the field a pointer. This pointer rule applies to scalar types (strings, numbers, booleans, enum types) and `$ref`-resolved or generated struct types only — it does not apply to `map[string]any`, slice, or `any` fields, since those are already nil-able.
* **Field Ordering**: Struct fields are emitted in alphabetical order of their JSON property name, not the declaration order from the source spec. This is because Go's `encoding/json` does not preserve object key order when unmarshaling into a map, so alphabetical order is used as the deterministic fallback.
* **Inline Objects to Named Structs**: Properties defined as an inline `"type": "object"` (without a `$ref`) that declare their own `properties` generate a named struct, named after the PascalCase property (e.g., `"compliance"` maps to `Compliance *Compliance`, generating a `Compliance` struct from those properties). This generated struct is emitted after the struct whose field references it. If the object has no `properties` of its own (a free-form object), the property instead maps to a map with string keys and any values: `map[string]any` (e.g., `"externalId"` maps to `ExternalId map[string]any`).
* **Arrays of Objects**: Properties defined as `"type": "array"` containing `"items": {"type": "object"}` map to a slice of a generated `<ArrayTypeName>Entry` struct (e.g., `"taxResidences"` with object items maps to `TaxResidences []TaxResidencesEntry`, generating a `TaxResidencesEntry` struct from those items' properties). This generated struct is emitted after the struct whose field references it. If the items schema has no `properties` of its own (a free-form object), the property instead maps to a slice of string-to-any maps: `[]map[string]any`.
* **Component References (`$ref`)**: When a property references another component schema (e.g., `"company": {"$ref": "#/components/schemas/CompanyDetailResponseDto"}`), type the field directly with the resolved PascalCase struct name (e.g., `Company CompanyDetailResponseDto`).

---

#### **3. Enforcing Enums with Custom Types**
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
