### **AI Guide: `omitzero` vs. `omitempty` for Optional Struct-Typed Fields**

This rule refines type-mapping.md §2's required/nullable → JSON-tag mapping. It does not
change any *type* decision (pointer vs. non-pointer is still decided exactly as type-mapping.md
§2 describes) — it only corrects which tag option is emitted for one specific combination,
because that combination's currently-documented `,omitempty` tag is a silent no-op for struct
Kind values.

---

#### **1. The problem**

`encoding/json`'s `omitempty` only omits `false`, `0`, `""`, `nil` (pointer/interface/slice/
map/chan/func), and empty arrays/slices/maps/strings. It has no special case for struct Kind
values — a zero-valued struct is not "empty" by this definition, `omitempty` or not. So a
non-pointer struct field is emitted into every payload regardless of whether it's the zero
value, silently defeating the optionality `required: false` was meant to express.

Go 1.24 added `omitzero` (see the `encoding/json` entry in the Go 1.24 release notes,
go.dev/doc/go1.24) to close exactly this gap: a field tagged `omitzero` is omitted if its value
`IsZero() bool` reports true, or, absent that method, if it equals its type's zero value —
checked recursively field-by-field, per the Go language specification's definition of a type's
zero value (go.dev/ref/spec#The_zero_value). That recursive check is exactly what's needed for
`DateTime`/`Date`/`OneOf[A, B]`/
`Discriminated[A, B]`-alias/`$ref`-resolved/inline-generated struct fields, none of which
implement `IsZero() bool` themselves — their "unset" state is their ordinary Go zero value
(e.g. `DateTime{}`, or `OneOf[A, B]{A: nil, B: nil}`).

---

#### **2. The rule**

Per the required/nullable matrix (type-mapping.md §2), only the `required: false` +
`nullable: false` row leaves a field non-pointer while still needing an optionality tag. For
that row:

* **Struct-Kind field → `omitzero`, not `omitempty`.** If the field's resolved Go type is a
  struct value (not `*T`) — this covers `DateTime`, `Date`, `OneOf[A, B]`, a `Discriminated[A,
  B]` type-alias, a `$ref`-resolved component struct, and an inline-generated named struct
  (property-object or array-item-object) — emit `,omitzero` in the JSON tag instead of
  `,omitempty`.
* **Everything else in that row is unaffected.** A scalar (`string`, `int64`, `int32`, `bool`)
  or an enum type (a defined `string`/numeric type, per type-mapping.md §3) keeps `,omitempty`
  exactly as type-mapping.md §2 already documents — `encoding/json` resolves emptiness for
  those by underlying Kind (`""`, `0`), which `omitempty` already handles correctly. A slice
  or `map[string]any` field also keeps `,omitempty` — `nil`/empty-length is exactly what
  `omitempty` already checks for those, and the pointer rule never applies to them (they're
  already nil-able).
* **Other rows are unaffected.** `required: true` rows carry no `omitempty`/`omitzero` tag at
  all today (the value must always be present), and the two `nullable: true` rows are pointer
  types, where `omitempty`'s existing nil-pointer check is already correct — none of that
  changes.
* **Never emit both.** Don't tag a field `,omitempty,omitzero` — for every type this rule
  touches, `omitzero`'s zero-value check already subsumes what `omitempty` would additionally
  catch, so pairing them is redundant, not more correct.

---

#### **3. Worked example**

```yaml
CompanyDetailResponseDto:
  type: object
  properties:
    createdAt:
      type: string
      format: date-time
  # createdAt is not in `required`, and has no nullable/allOf wrapper
```

```go
// Before this rule (type-mapping.md §2 alone): never omitted — DateTime{} is not "empty".
CreatedAt DateTime `json:"createdAt,omitempty"`

// After this rule: omitted when CreatedAt is the zero-value DateTime{}.
CreatedAt DateTime `json:"createdAt,omitzero"`
```

---

#### **4. Implementation**

`generator.buildField` (`generator/generator.go`) picks the tag suffix from two signals:
`nullable` and a third `isStruct` return value threaded through the whole type-resolution
chain — `resolveRefOrType` → `resolveSchemaType` → `resolveScalarSchemaType` →
`resolveObjectSchemaType`/`resolveStringSchemaType`/`resolveNumericSchemaType`. Each of those
reports `isStruct` alongside the Go type name it resolves: `true` for `DateTime`, `Date`,
`OneOf[A, B]`, an inline-generated named struct, or a `$ref`-resolved component struct (always
`true` there, via `unwrapRef`'s branch in `resolveRefOrType`); `false` for every scalar, enum,
`map[...]`, `[]...`, and `any` case. `buildField` then does:

```go
switch {
case required:
    // no suffix
case !nullable && isStruct:
    tag += ",omitzero"
default:
    tag += ",omitempty"
}
```

The `fixtures/OmitZero/` golden-file case exercises the full matrix in one schema: a required
scalar (no suffix), an optional `DateTime` and an optional `$ref` struct (both `,omitzero`), a
nullable `$ref` struct (pointer, `,omitempty`), an optional enum/map/slice (`,omitempty`,
unaffected), and an optional `OneOf[...]` plus an optional inline-object field (both
`,omitzero`). Regenerating any fixture with `UPDATE_GOLDEN=1 go test ./generator/...` after a
type-mapping change will reflect this rule automatically — no separate omitzero-specific
regeneration step is needed.
