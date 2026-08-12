package generated

// SimpleUser is generated from components.schemas.SimpleUser.
type SimpleUser struct {
	Username string `json:"username"`
}

// ArraysAndCollections is generated from components.schemas.ArraysAndCollections.
type ArraysAndCollections struct {
	IntegerList []int64      `json:"integer_list,omitempty"`
	ObjectList  []SimpleUser `json:"object_list,omitempty"`
	StringList  []string     `json:"string_list,omitempty"`
}
