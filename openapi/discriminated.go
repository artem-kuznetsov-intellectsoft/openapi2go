package openapi

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Discriminated holds a value that is exactly one of two possible types,
// selected using an OpenAPI discriminator: the concrete type is the one
// with an exported string field whose value equals its own Go type name —
// the OpenAPI 3.0 default discriminator behavior when no explicit
// `mapping` is declared.
type Discriminated[A, B any] struct {
	A *A
	B *B
}

// MarshalJSON serializes whichever of A or B is set, or null if neither is.
func (d Discriminated[A, B]) MarshalJSON() ([]byte, error) {
	switch {
	case d.A != nil:
		return json.Marshal(d.A)
	case d.B != nil:
		return json.Marshal(d.B)
	default:
		return []byte("null"), nil
	}
}

// UnmarshalJSON decodes data into both A and B, then keeps whichever one's
// discriminator field value matches its own type name.
func (d *Discriminated[A, B]) UnmarshalJSON(data []byte) error {
	var a A
	aMatch := json.Unmarshal(data, &a) == nil && hasMatchingDiscriminator(a)

	var b B
	bMatch := json.Unmarshal(data, &b) == nil && hasMatchingDiscriminator(b)

	switch {
	case aMatch:
		d.A = &a
	case bMatch:
		d.B = &b
	default:
		return fmt.Errorf("openapi: no discriminator field on %T or %T matches its own type name", a, b)
	}

	return nil
}

// hasMatchingDiscriminator reports whether v has an exported string field
// whose value equals v's own Go type name.
func hasMatchingDiscriminator(v any) bool {
	rv := reflect.ValueOf(v)
	typeName := rv.Type().Name()

	for _, f := range rv.Fields() {
		if f.Kind() == reflect.String && f.String() == typeName {
			return true
		}
	}

	return false
}
