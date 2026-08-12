package test1

import "time"

// CustomerType represents the customer type.
type CustomerType string

const (
	CustomerTypeIndividual CustomerType = "INDIVIDUAL"
	CustomerTypeCompany    CustomerType = "COMPANY"
)

// VerificationStatus represents the verification status.
type VerificationStatus string

const (
	VerificationStatusPendingVerification       VerificationStatus = "PENDING_VERIFICATION"
	VerificationStatusApproved                  VerificationStatus = "APPROVED"
	VerificationStatusPendingManualVerification VerificationStatus = "PENDING_MANUAL_VERIFICATION"
	VerificationStatusRejected                  VerificationStatus = "REJECTED"
)

type CompanyDetailResponseDto struct{}

// CustomerDetailCompanyResponseDto is generated from components.schemas.CustomerDetailCompanyResponseDto.
type CustomerDetailCompanyResponseDto struct {
	Id                        string                   `json:"id"`
	ExternalId                any                      `json:"externalId,omitempty"`
	Name                      string                   `json:"name"`
	PhysicalAddress           any                      `json:"physicalAddress"`
	CompanyTaxId              any                      `json:"companyTaxId"`
	CompanyRegistrationNumber any                      `json:"companyRegistrationNumber"`
	DefaultChain              any                      `json:"defaultChain"`
	DefaultChainAddress       any                      `json:"defaultChainAddress"`
	DefaultBankAddress        any                      `json:"defaultBankAddress"`
	CustomerType              CustomerType             `json:"customerType"`
	CreatedAt                 time.Time                `json:"createdAt"`
	UpdatedAt                 time.Time                `json:"updatedAt"`
	VerificationStatus        VerificationStatus       `json:"verificationStatus"`
	LastScreenedAt            any                      `json:"lastScreenedAt"`
	Vibans                    []any                    `json:"vibans"`
	Addresses                 []any                    `json:"addresses"`
	Company                   CompanyDetailResponseDto `json:"company"`
}
