package generated

import "github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"

// PrimitivesAndFormats is generated from components.schemas.PrimitivesAndFormats.
type PrimitivesAndFormats struct {
	BinaryField   string           `json:"binary_field,omitempty"`
	BooleanField  bool             `json:"boolean_field,omitempty"`
	ByteField     []byte           `json:"byte_field,omitempty"`
	DateField     openapi.Date     `json:"date_field,omitempty"`
	DateTimeField openapi.DateTime `json:"date_time_field,omitempty"`
	DoubleField   float64          `json:"double_field,omitempty"`
	FloatField    float32          `json:"float_field,omitempty"`
	Int32Field    int32            `json:"int32_field,omitempty"`
	Int64Field    int64            `json:"int64_field,omitempty"`
	PasswordField string           `json:"password_field,omitempty"`
	StringField   string           `json:"string_field,omitempty"`
}
