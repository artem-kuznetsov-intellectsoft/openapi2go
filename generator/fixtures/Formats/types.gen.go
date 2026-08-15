package generated

// AllSpecFormats is generated from components.schemas.AllSpecFormats.
type AllSpecFormats struct {
	IntegerInt32   int32    `json:"integer_int32"`
	IntegerInt64   int64    `json:"integer_int64"`
	NumberDouble   float64  `json:"number_double"`
	NumberFloat    float32  `json:"number_float"`
	StringBinary   string   `json:"string_binary"`
	StringByte     []byte   `json:"string_byte"`
	StringDate     Date     `json:"string_date"`
	StringDateTime DateTime `json:"string_date-time"`
	StringEmail    string   `json:"string_email"`
	StringPassword string   `json:"string_password"`
	StringUri      string   `json:"string_uri"`
	StringUuid     string   `json:"string_uuid"`
}
