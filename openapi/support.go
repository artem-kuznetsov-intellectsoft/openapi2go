package openapi

import _ "embed"

//go:embed date.go
var dateSource string

//go:embed oneof.go
var oneofSource string

//go:embed discriminated.go
var discriminatedSource string

//go:embed clientruntime/client_runtime.go
var clientRuntimeSource string

// SupportFiles returns each support-type source file embedded from this
// package, keyed by filename, with its verbatim content. Tools that generate
// Go code from an OpenAPI spec (see the generator package) copy the entries
// their output needs into the generated code's own output directory, so the
// generated package can define DateTime, Date, OneOf, Discriminated, and the
// Client/APIError HTTP plumbing itself instead of importing this module.
//
// client_runtime.go comes from the clientruntime subpackage rather than this
// one, so its Client/HTTPResponse/APIError names do not collide with the
// OpenAPI object model declared here.
func SupportFiles() map[string]string {
	return map[string]string{
		"date.go":           dateSource,
		"oneof.go":          oneofSource,
		"discriminated.go":  discriminatedSource,
		"client_runtime.go": clientRuntimeSource,
	}
}
