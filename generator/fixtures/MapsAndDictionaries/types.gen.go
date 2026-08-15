package generated

// SimpleUser is generated from components.schemas.SimpleUser.
type SimpleUser struct {
	Username string `json:"username"`
}

// MapsAndDictionaries is generated from components.schemas.MapsAndDictionaries.
type MapsAndDictionaries struct {
	ObjectMap map[string]SimpleUser `json:"object_map,omitempty"`
	StringMap map[string]string     `json:"string_map,omitempty"`
}
