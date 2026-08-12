package generated

import "time"

// BaseModel is generated from components.schemas.BaseModel.
type BaseModel struct {
	CreatedAt *time.Time `json:"created_at,omitempty"`
	Id        string     `json:"id"`
}

// CompositionAllOf is generated from components.schemas.CompositionAllOf.
type CompositionAllOf struct {
	BaseModel
	ExtendedProperty string `json:"extended_property"`
}
