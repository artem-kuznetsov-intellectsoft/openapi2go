package generated

import "github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"

// OneOfPrimitives is generated from components.schemas.OneOfPrimitives.
type OneOfPrimitives struct {
	IdOrCode openapi.OneOf[string, int64] `json:"id_or_code,omitempty"`
}
