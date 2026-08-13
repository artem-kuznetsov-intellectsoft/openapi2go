package generated

import "encoding/json"

// OneOf holds a value that is exactly one of two possible types. After
// successful unmarshaling, exactly one of A or B is non-nil.
type OneOf[A, B any] struct {
	A *A
	B *B
}

// MarshalJSON serializes whichever of A or B is set, or null if neither is.
func (o OneOf[A, B]) MarshalJSON() ([]byte, error) {
	switch {
	case o.A != nil:
		return json.Marshal(o.A)
	case o.B != nil:
		return json.Marshal(o.B)
	default:
		return []byte("null"), nil
	}
}

// UnmarshalJSON tries A first, then B, keeping whichever one decodes
// successfully.
func (o *OneOf[A, B]) UnmarshalJSON(data []byte) error {
	var a A
	if err := json.Unmarshal(data, &a); err == nil {
		o.A = &a
		return nil
	}

	var b B
	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}
	o.B = &b

	return nil
}
