package generated

type Response200 struct {
	Ok bool `json:"ok,omitempty"`
}

// GetResourceParams is generated for operationId GetResource.
type GetResourceParams struct {
	XApiKey  string
	XTraceId *string
	Id       string
}

type GetResourceResponse200 struct {
	Id string `json:"id,omitempty"`
}

type Response404 struct{}

func (r Response404) Error() string {
	panic("TODO: define the output")
}

// ReplaceResourceParams is generated for operationId ReplaceResource.
type ReplaceResourceParams struct {
	XApiKey string
	Id      string
}

// DeleteResourceParams is generated for operationId DeleteResource.
type DeleteResourceParams struct {
	XApiKey string
	Id      string
}

// OptionsResourceParams is generated for operationId OptionsResource.
type OptionsResourceParams struct {
	XApiKey string
	Id      string
}

// HeadResourceParams is generated for operationId HeadResource.
type HeadResourceParams struct {
	XApiKey string
	Id      string
}

// PatchResourceParams is generated for operationId PatchResource.
type PatchResourceParams struct {
	XApiKey string
	Id      string
}

// TraceResourceParams is generated for operationId TraceResource.
type TraceResourceParams struct {
	XApiKey string
	Id      string
}
