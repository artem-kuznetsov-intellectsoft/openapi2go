package generated

// SimpleUser is generated from components.schemas.SimpleUser.
type SimpleUser struct {
	Username string `json:"username"`
}

// OneOfObjectsWithoutDiscriminator is generated from components.schemas.OneOfObjectsWithoutDiscriminator.
type OneOfObjectsWithoutDiscriminator struct {
	AssignedTo OneOf[string, SimpleUser] `json:"assigned_to,omitzero"`
}
