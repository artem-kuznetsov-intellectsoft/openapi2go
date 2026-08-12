package test1

import "time"

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
	Id                        string                   `json:"id"`
	ExternalId                map[string]any           `json:"externalId,omitempty"`
	Name                      string                   `json:"name"`
	PhysicalAddress           map[string]any           `json:"physicalAddress"`
	CompanyTaxId              map[string]any           `json:"companyTaxId"`
	CompanyRegistrationNumber map[string]any           `json:"companyRegistrationNumber"`
	DefaultChain              map[string]any           `json:"defaultChain"`
	DefaultChainAddress       map[string]any           `json:"defaultChainAddress"`
	DefaultBankAddress        map[string]any           `json:"defaultBankAddress"`
	CustomerType              CustomerType             `json:"customerType"`
	CreatedAt                 time.Time                `json:"createdAt"`
	UpdatedAt                 time.Time                `json:"updatedAt"`
	VerificationStatus        VerificationStatus       `json:"verificationStatus"`
	LastScreenedAt            any                      `json:"lastScreenedAt"`
	Vibans                    []any                    `json:"vibans"`
	Addresses                 []any                    `json:"addresses"`
	Company                   CompanyDetailResponseDto `json:"company"`
}
