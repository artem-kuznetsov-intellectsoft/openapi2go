package generated

// Request is generated from components.schemas.Request.
type Request struct {
	Age  int64  `json:"age,omitempty"`
	Name string `json:"name,omitempty"`
}

// ResponseBadRequest is generated from components.schemas.ResponseBadRequest.
type ResponseBadRequest struct {
	Message string `json:"message,omitempty"`
}

func (r ResponseBadRequest) Error() error {
	panic("TODO: define the output")
}

type Response404 struct{}

func (r Response404) Error() error {
	panic("TODO: define the output")
}

// ResponseOK is generated from components.schemas.ResponseOK.
type ResponseOK struct {
	Id string `json:"id,omitempty"`
}
