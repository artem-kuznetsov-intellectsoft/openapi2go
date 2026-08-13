package generated

// CustomerControllerGetAllCustomersParams is generated for operationId CustomerController_getAllCustomers.
type CustomerControllerGetAllCustomersParams struct {
	Limit float64
	Page  *float64
}

type CustomerListCompanyResponseDto struct{}

type CustomerListIndividualResponseDto struct{}

type Response200 struct {
	Customers  []OneOf[CustomerListCompanyResponseDto, CustomerListIndividualResponseDto] `json:"customers,omitempty"`
	Pagination Pagination                                                                 `json:"pagination,omitempty"`
}

type Pagination struct {
	Limit      float64 `json:"limit,omitempty"`
	Page       float64 `json:"page,omitempty"`
	Total      float64 `json:"total,omitempty"`
	TotalPages float64 `json:"totalPages,omitempty"`
}

type Response401 struct{}

func (r Response401) Error() error {
	panic("TODO: define the output")
}
