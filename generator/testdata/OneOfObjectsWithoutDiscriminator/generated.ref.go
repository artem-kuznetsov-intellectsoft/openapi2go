package generated

import "github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"

// SimpleUser is generated from components.schemas.SimpleUser.
type SimpleUser struct {
	Username string `json:"username"`
}

// OneOfObjectsWithoutDiscriminator is generated from components.schemas.OneOfObjectsWithoutDiscriminator.
type OneOfObjectsWithoutDiscriminator struct {
	AssignedTo openapi.OneOf[string, SimpleUser] `json:"assigned_to,omitempty"`
}
