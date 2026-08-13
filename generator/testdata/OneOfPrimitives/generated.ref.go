package generated

// OneOfPrimitives is generated from components.schemas.OneOfPrimitives.
type OneOfPrimitives struct {
	IdOrCode OneOf[string, int64] `json:"id_or_code,omitempty"`
}
