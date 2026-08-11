package examples

type CustomerDetailCompanyResponseDto struct {
	Id           string                   `json:"id"`
	ExternalId   map[string]string        `json:"externalId"`
	CustomerType CustomerType             `json:"customerType"`
	CreatedAt    *string                  `json:"createdAt"`
	Vibans       []map[string]string      `json:"vibans"`
	Company      CompanyDetailResponseDto `json:"company"`
}

type CustomerType string

const (
	CustomerTypeINDIVIDUAL CustomerType = "INDIVIDUAL"
	CustomerTypeCOMPANY    CustomerType = "COMPANY"
)

type CompanyDetailResponseDto struct{}
