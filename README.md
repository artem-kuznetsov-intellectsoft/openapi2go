### **AI Guide: Converting OpenAPI Specifications to Go Code**

When converting OpenAPI YAML/JSON schema definitions into Go data transfer objects (DTOs), follow these strict structural mapping rules to maintain compatibility, support nullability, and enforce typings:

---

#### **1. Field Naming & Struct Tagging**
* **Case Conversion**: Convert JSON property names written in `camelCase` (e.g., `externalId`, `companyTaxId`) into `PascalCase` to declare exported fields in Go (e.g., `ExternalId`, `CompanyTaxId`).
* **JSON Tagging**: Annotate every struct field with a `json` struct tag matching the exact name and case of the property from the JSON schema (e.g., `json:"id"`, `json:"createdAt"`).

---

#### **2. Type Mapping & Representation**
* **Basic Strings**: Map default OpenAPI properties defined as `"type": "string"` (including special formats like `uuid`) to primitive Go `string` fields (e.g., `"id"` maps to `Id string`).
* **Temporal Strings (Date-Time) & Nullability**: Map date-time properties (such as `"createdAt"` which has `"format": "date-time"`) to a string (`string`).
* **Required and Nullability**: `required:false` (or just missing `required` property) or `nullable:true` means the field must be a pointer (e.g. `*string`) to natively support optionality or null values.
* **Untyped Objects to Maps**: Properties defined as a generic `"type": "object"` (without a `$ref`) must be represented as a map with string keys and any values: `map[string]any` (e.g., `"externalId"` maps to `ExternalId map[string]any`).
* **Arrays of Objects**: Properties defined as `"type": "array"` containing `"items": {"type": "object"}` must map to a slice of string-to-any maps: `[]map[string]any` (e.g., `"vibans"` maps to `Vibans []map[string]any`).
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
      CustomerTypeIndividual CustomerType = "INDIVIDUAL"
      CustomerTypeCOMPANY    CustomerType = "COMPANY"
  )
  ```

* **Field Integration**: Define the corresponding field in the parent struct using this custom enum type rather than a standard string (e.g., `CustomerType CustomerType`).
