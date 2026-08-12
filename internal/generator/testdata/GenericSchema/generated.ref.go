package generated

import "time"

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

// BaseModel is generated from components.schemas.BaseModel.
type BaseModel struct {
	CreatedAt *time.Time `json:"created_at,omitempty"`
	Id        string     `json:"id"`
}

// CompositionAllOf is generated from components.schemas.CompositionAllOf.
type CompositionAllOf struct{}

// MapsAndDictionaries is generated from components.schemas.MapsAndDictionaries.
type MapsAndDictionaries struct {
	ObjectMap map[string]any `json:"object_map,omitempty"`
	StringMap map[string]any `json:"string_map,omitempty"`
}

// NullableFields is generated from components.schemas.NullableFields.
type NullableFields struct {
	NonNullableString *string `json:"non_nullable_string,omitempty"`
	NullableInt       *int64  `json:"nullable_int,omitempty"`
	NullableString    *string `json:"nullable_string,omitempty"`
}

// PolymorphicCat is generated from components.schemas.PolymorphicCat.
type PolymorphicCat struct{}

// PolymorphicPet is generated from components.schemas.PolymorphicPet.
type PolymorphicPet struct {
	Name    string `json:"name"`
	PetType string `json:"pet_type"`
}

// PrimitivesAndFormats is generated from components.schemas.PrimitivesAndFormats.
type PrimitivesAndFormats struct {
	BinaryField   *string    `json:"binary_field,omitempty"`
	BooleanField  *bool      `json:"boolean_field,omitempty"`
	ByteField     *string    `json:"byte_field,omitempty"`
	DateField     *string    `json:"date_field,omitempty"`
	DateTimeField *time.Time `json:"date_time_field,omitempty"`
	DoubleField   *float64   `json:"double_field,omitempty"`
	FloatField    *float64   `json:"float_field,omitempty"`
	Int32Field    *int64     `json:"int32_field,omitempty"`
	Int64Field    *int64     `json:"int64_field,omitempty"`
	PasswordField *string    `json:"password_field,omitempty"`
	StringField   *string    `json:"string_field,omitempty"`
}

// ReadWriteOnly is generated from components.schemas.ReadWriteOnly.
type ReadWriteOnly struct {
	Id          *string `json:"id,omitempty"`
	NormalField *string `json:"normal_field,omitempty"`
	Password    *string `json:"password,omitempty"`
}

// RequiredAndOptional is generated from components.schemas.RequiredAndOptional.
type RequiredAndOptional struct {
	OptionalInteger *int64  `json:"optional_integer,omitempty"`
	OptionalString  *string `json:"optional_string,omitempty"`
	RequiredInteger int64   `json:"required_integer"`
	RequiredString  string  `json:"required_string"`
}
