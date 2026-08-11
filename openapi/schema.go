package openapi

import (
	"encoding/json"
)

// OpenAPI represents the root OpenAPI 3.0.x Document (the OpenAPI Object).
type OpenAPI struct {
	OpenAPI      string                 `json:"openapi"`
	Info         Info                   `json:"info"`
	Servers      []Server               `json:"servers,omitempty"`
	Paths        Paths                  `json:"paths"`
	Components   *Components            `json:"components,omitempty"`
	Security     []SecurityRequirement  `json:"security,omitempty"`
	Tags         []Tag                  `json:"tags,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`
}

// Info represents metadata about the API (the Info Object).
type Info struct {
	Title          string   `json:"title"`
	Description    string   `json:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
	Version        string   `json:"version"`
}

// Contact represents contact information for the exposed API (the Contact Object).
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License represents license information for the exposed API (the License Object).
type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Server represents a target server environment (the Server Object).
type Server struct {
	URL         string                    `json:"url"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

// ServerVariable represents a variable for server URL template substitution (the Server Variable Object).
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

// Paths holds the relative paths to the individual endpoints and their operations (the Paths Object).
// Keys must begin with a forward slash (/).
type Paths map[string]*PathItem

// PathItem describes the operations available on a single path (the Path Item Object).
type PathItem struct {
	Ref         string               `json:"$ref,omitempty"`
	Summary     string               `json:"summary,omitempty"`
	Description string               `json:"description,omitempty"`
	Get         *Operation           `json:"get,omitempty"`
	Put         *Operation           `json:"put,omitempty"`
	Post        *Operation           `json:"post,omitempty"`
	Delete      *Operation           `json:"delete,omitempty"`
	Options     *Operation           `json:"options,omitempty"`
	Head        *Operation           `json:"head,omitempty"`
	Patch       *Operation           `json:"patch,omitempty"`
	Trace       *Operation           `json:"trace,omitempty"`
	Servers     []Server             `json:"servers,omitempty"`
	Parameters  []*RefOr[*Parameter] `json:"parameters,omitempty"`
}

// Operation describes a single API operation on a path (the Operation Object).
type Operation struct {
	Tags         []string                     `json:"tags,omitempty"`
	Summary      string                       `json:"summary,omitempty"`
	Description  string                       `json:"description,omitempty"`
	ExternalDocs *ExternalDocumentation       `json:"externalDocs,omitempty"`
	OperationID  string                       `json:"operationId,omitempty"`
	Parameters   []*RefOr[*Parameter]         `json:"parameters,omitempty"`
	RequestBody  *RefOr[*RequestBody]         `json:"requestBody,omitempty"`
	Responses    Responses                    `json:"responses"`
	Callbacks    map[string]*RefOr[*Callback] `json:"callbacks,omitempty"`
	Deprecated   bool                         `json:"deprecated,omitempty"`
	Security     []SecurityRequirement        `json:"security,omitempty"`
	Servers      []Server                     `json:"servers,omitempty"`
}

// Parameter describes a single operation parameter (the Parameter Object).
type Parameter struct {
	Name            string                      `json:"name"`
	In              string                      `json:"in"` // "query", "header", "path", "cookie"
	Description     string                      `json:"description,omitempty"`
	Required        bool                        `json:"required,omitempty"`
	Deprecated      bool                        `json:"deprecated,omitempty"`
	AllowEmptyValue bool                        `json:"allowEmptyValue,omitempty"`
	Style           string                      `json:"style,omitempty"`
	Explode         *bool                       `json:"explode,omitempty"`
	AllowReserved   bool                        `json:"allowReserved,omitempty"`
	Schema          *RefOr[*Schema]             `json:"schema,omitempty"`
	Example         any                         `json:"example,omitempty"`
	Examples        map[string]*RefOr[*Example] `json:"examples,omitempty"`
	Content         map[string]*MediaType       `json:"content,omitempty"`
}

// RequestBody describes a single request body (the Request Body Object).
type RequestBody struct {
	Description string                `json:"description,omitempty"`
	Content     map[string]*MediaType `json:"content"`
	Required    bool                  `json:"required,omitempty"`
}

// MediaType provides schema and examples for a media type (the Media Type Object).
type MediaType struct {
	Schema   *RefOr[*Schema]             `json:"schema,omitempty"`
	Example  any                         `json:"example,omitempty"`
	Examples map[string]*RefOr[*Example] `json:"examples,omitempty"`
	Encoding map[string]*Encoding        `json:"encoding,omitempty"`
}

// Encoding represents encoding definitions applied to schema properties (the Encoding Object).
type Encoding struct {
	ContentType   string                     `json:"contentType,omitempty"`
	Headers       map[string]*RefOr[*Header] `json:"headers,omitempty"`
	Style         string                     `json:"style,omitempty"`
	Explode       *bool                      `json:"explode,omitempty"`
	AllowReserved bool                       `json:"allowReserved,omitempty"`
}

// Responses maps HTTP response status codes to expected responses (the Responses Object).
// Key is either an HTTP status code string (e.g. "200") or "default".
type Responses map[string]*RefOr[*Response]

// Response describes a single response from an API operation (the Response Object).
type Response struct {
	Description string                     `json:"description"`
	Headers     map[string]*RefOr[*Header] `json:"headers,omitempty"`
	Content     map[string]*MediaType      `json:"content,omitempty"`
	Links       map[string]*RefOr[*Link]   `json:"links,omitempty"`
}

// Components holds a set of reusable objects for different aspects of the OAS (the Components Object).
type Components struct {
	Schemas         map[string]*RefOr[*Schema]         `json:"schemas,omitempty"`
	Responses       map[string]*RefOr[*Response]       `json:"responses,omitempty"`
	Parameters      map[string]*RefOr[*Parameter]      `json:"parameters,omitempty"`
	Examples        map[string]*RefOr[*Example]        `json:"examples,omitempty"`
	RequestBodies   map[string]*RefOr[*RequestBody]    `json:"requestBodies,omitempty"`
	Headers         map[string]*RefOr[*Header]         `json:"headers,omitempty"`
	SecuritySchemes map[string]*RefOr[*SecurityScheme] `json:"securitySchemes,omitempty"`
	Links           map[string]*RefOr[*Link]           `json:"links,omitempty"`
	Callbacks       map[string]*RefOr[*Callback]       `json:"callbacks,omitempty"`
}

// Schema represents input and output data types (the Schema Object, extended JSON Schema).
type Schema struct {
	// JSON Schema properties
	Title                string                     `json:"title,omitempty"`
	MultipleOf           *float64                   `json:"multipleOf,omitempty"`
	Maximum              *float64                   `json:"maximum,omitempty"`
	ExclusiveMaximum     bool                       `json:"exclusiveMaximum,omitempty"`
	Minimum              *float64                   `json:"minimum,omitempty"`
	ExclusiveMinimum     bool                       `json:"exclusiveMinimum,omitempty"`
	MaxLength            *uint64                    `json:"maxLength,omitempty"`
	MinLength            *uint64                    `json:"minLength,omitempty"`
	Pattern              string                     `json:"pattern,omitempty"`
	MaxItems             *uint64                    `json:"maxItems,omitempty"`
	MinItems             *uint64                    `json:"minItems,omitempty"`
	UniqueItems          bool                       `json:"uniqueItems,omitempty"`
	MaxProperties        *uint64                    `json:"maxProperties,omitempty"`
	MinProperties        *uint64                    `json:"minProperties,omitempty"`
	Required             []string                   `json:"required,omitempty"`
	Enum                 []any                      `json:"enum,omitempty"`
	Type                 string                     `json:"type,omitempty"`
	AllOf                []*RefOr[*Schema]          `json:"allOf,omitempty"`
	OneOf                []*RefOr[*Schema]          `json:"oneOf,omitempty"`
	AnyOf                []*RefOr[*Schema]          `json:"anyOf,omitempty"`
	Not                  *RefOr[*Schema]            `json:"not,omitempty"`
	Items                *RefOr[*Schema]            `json:"items,omitempty"`
	Properties           map[string]*RefOr[*Schema] `json:"properties,omitempty"`
	AdditionalProperties any                        `json:"additionalProperties,omitempty"` // boolean or *RefOr[*Schema]
	Description          string                     `json:"description,omitempty"`
	Format               string                     `json:"format,omitempty"`
	Default              any                        `json:"default,omitempty"`

	// OpenAPI Extensions
	Nullable      bool                   `json:"nullable,omitempty"`
	Discriminator *Discriminator         `json:"discriminator,omitempty"`
	ReadOnly      bool                   `json:"readOnly,omitempty"`
	WriteOnly     bool                   `json:"writeOnly,omitempty"`
	XML           *XML                   `json:"xml,omitempty"`
	ExternalDocs  *ExternalDocumentation `json:"externalDocs,omitempty"`
	Example       any                    `json:"example,omitempty"`
	Deprecated    bool                   `json:"deprecated,omitempty"`
}

// SecurityRequirement lists the required security schemes (the Security Requirement Object).
// Each key corresponds to a Security Scheme declared in Components.
type SecurityRequirement map[string][]string

// Tag adds metadata to a single tag used by the Operation Object (the Tag Object).
type Tag struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`
}

// ExternalDocumentation allows referencing an external resource for extended documentation (the External Documentation Object).
type ExternalDocumentation struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
}

// Example represents a parameter or media type example (the Example Object).
type Example struct {
	Summary       string `json:"summary,omitempty"`
	Description   string `json:"description,omitempty"`
	Value         any    `json:"value,omitempty"`
	ExternalValue string `json:"externalValue,omitempty"`
}

// Header describes a single response header (the Header Object).
type Header struct {
	Description string                      `json:"description,omitempty"`
	Required    bool                        `json:"required,omitempty"`
	Deprecated  bool                        `json:"deprecated,omitempty"`
	Style       string                      `json:"style,omitempty"`
	Explode     *bool                       `json:"explode,omitempty"`
	Schema      *RefOr[*Schema]             `json:"schema,omitempty"`
	Example     any                         `json:"example,omitempty"`
	Examples    map[string]*RefOr[*Example] `json:"examples,omitempty"`
	Content     map[string]*MediaType       `json:"content,omitempty"`
}

// Discriminator adds support for polymorphism (the Discriminator Object).
type Discriminator struct {
	PropertyName string            `json:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty"`
}

// XML adds additional metadata for fine-tuning XML model definitions (the XML Object).
type XML struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Prefix    string `json:"prefix,omitempty"`
	Attribute bool   `json:"attribute,omitempty"`
	Wrapped   bool   `json:"wrapped,omitempty"`
}

// Link represents a design-time relationship between a response and another operation (the Link Object).
type Link struct {
	OperationRef string         `json:"operationRef,omitempty"`
	OperationID  string         `json:"operationId,omitempty"`
	Parameters   map[string]any `json:"parameters,omitempty"`
	RequestBody  any            `json:"requestBody,omitempty"`
	Description  string         `json:"description,omitempty"`
	Server       *Server        `json:"server,omitempty"`
}

// SecurityScheme defines a security mechanism (the Security Scheme Object).
type SecurityScheme struct {
	Type             string      `json:"type"` // "apiKey", "http", "oauth2", "openIdConnect"
	Description      string      `json:"description,omitempty"`
	Name             string      `json:"name,omitempty"`             // Required for apiKey
	In               string      `json:"in,omitempty"`               // Required for apiKey ("query", "header", "cookie")
	InScheme         string      `json:"scheme,omitempty"`           // Required for http (e.g. "basic", "bearer")
	BearerFormat     string      `json:"bearerFormat,omitempty"`     // e.g. "JWT"
	Flows            *OAuthFlows `json:"flows,omitempty"`            // Required for oauth2
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitempty"` // Required for openIdConnect
}

// OAuthFlows allows configuration of the supported OAuth Flows (the OAuth Flows Object).
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}

// OAuthFlow details configuration details for a supported OAuth Flow (the OAuth Flow Object).
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"` // Required for implicit / authorizationCode
	TokenURL         string            `json:"tokenUrl,omitempty"`         // Required for password / clientCredentials / authorizationCode
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"` // Required
}

// Callback represents a map of out-of-band callbacks (the Callback Object).
// Keys are runtime expressions evaluated at runtime.
type Callback map[string]*PathItem

// RefOr is a generic wrapper supporting either a direct inline Object or an external reference string ($ref).
// This is used for all OpenAPI component structures that can be referenced.
type RefOr[T any] struct {
	Ref   string `json:"$ref,omitempty"`
	Value T      `json:"-"`
}

// MarshalJSON customizes JSON serialization for RefOr to output either the reference or the inline value.
func (r *RefOr[T]) MarshalJSON() ([]byte, error) {
	if r.Ref != "" {
		return json.Marshal(&struct {
			Ref string `json:"$ref"`
		}{
			Ref: r.Ref,
		})
	}

	return json.Marshal(r.Value)
}

// UnmarshalJSON customizes JSON deserialization for RefOr to capture $ref or populate the inline value.
func (r *RefOr[T]) UnmarshalJSON(data []byte) error {
	var refOnly struct {
		Ref string `json:"$ref"`
	}
	if err := json.Unmarshal(data, &refOnly); err == nil && refOnly.Ref != "" {
		r.Ref = refOnly.Ref

		return nil
	}

	return json.Unmarshal(data, &r.Value)
}
