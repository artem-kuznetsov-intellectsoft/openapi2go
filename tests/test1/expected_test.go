package test1_test

// CustomerDetailCompanyResponseDto represetns 
// a schema components.schemas.CustomerDetailCompanyResponseDto 
// from components.schemas.CustomerDetailCompanyResponseDto.json.
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

type CompanyDetailResponseDto struct {
	TaxId                map[string]string              `json:"jurisdictionCountryCode"`
	LegalRepresentatives []CompanyIndividualResponseDto `json:"legalRepresentatives"`
}

type CompanyIndividualResponseDto struct{}
