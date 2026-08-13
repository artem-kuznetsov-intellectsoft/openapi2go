package generated

import "github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"

// BaseModel is generated from components.schemas.BaseModel.
type BaseModel struct {
	CreatedAt openapi.DateTime `json:"created_at,omitempty"`
	Id        string           `json:"id"`
}

// CompositionAllOf is generated from components.schemas.CompositionAllOf.
type CompositionAllOf struct {
	BaseModel
	ExtendedProperty string `json:"extended_property"`
}
