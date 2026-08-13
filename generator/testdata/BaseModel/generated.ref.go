package generated

import (
	"github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"
	"time"
)

// BaseModel is generated from components.schemas.BaseModel.
type BaseModel struct {
	CreatedAt time.Time    `json:"created_at,omitempty"`
	Id        string       `json:"id"`
	UpdatedAt openapi.Date `json:"updated_at,omitempty"`
}
