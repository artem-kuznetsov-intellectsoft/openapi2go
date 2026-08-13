package generated

// CustomerType represents the customer type.
type CustomerType string

const (
	CustomerTypeINDIVIDUAL CustomerType = "INDIVIDUAL"
	CustomerTypeCOMPANY    CustomerType = "COMPANY"
)

// VerificationStatus represents the verification status.
type VerificationStatus string

const (
	VerificationStatusPENDING_VERIFICATION        VerificationStatus = "PENDING_VERIFICATION"
	VerificationStatusAPPROVED                    VerificationStatus = "APPROVED"
	VerificationStatusPENDING_MANUAL_VERIFICATION VerificationStatus = "PENDING_MANUAL_VERIFICATION"
	VerificationStatusREJECTED                    VerificationStatus = "REJECTED"
)

type CompanyDetailResponseDto struct{}

// CustomerDetailCompanyResponseDto is generated from components.schemas.CustomerDetailCompanyResponseDto.
type CustomerDetailCompanyResponseDto struct {
	Addresses                 []map[string]any         `json:"addresses"`
	Company                   CompanyDetailResponseDto `json:"company"`
	CompanyRegistrationNumber map[string]any           `json:"companyRegistrationNumber"`
	CompanyTaxId              map[string]any           `json:"companyTaxId"`
	CreatedAt                 DateTime                 `json:"createdAt"`
	CustomerType              CustomerType             `json:"customerType"`
	DefaultBankAddress        map[string]any           `json:"defaultBankAddress"`
	DefaultChain              map[string]any           `json:"defaultChain"`
	DefaultChainAddress       map[string]any           `json:"defaultChainAddress"`
	ExternalId                map[string]any           `json:"externalId,omitempty"`
	Id                        string                   `json:"id"`
	LastScreenedAt            map[string]any           `json:"lastScreenedAt"`
	Name                      string                   `json:"name"`
	PhysicalAddress           map[string]any           `json:"physicalAddress"`
	UpdatedAt                 DateTime                 `json:"updatedAt"`
	VerificationStatus        VerificationStatus       `json:"verificationStatus"`
	Vibans                    []map[string]any         `json:"vibans"`
}
