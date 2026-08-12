package generated

import "time"

type IndividualComplianceResponseDto struct{}

type PhysicalAddressResponseDto struct{}

// IndividualResponseDto is generated from components.schemas.IndividualResponseDto.
type IndividualResponseDto struct {
	Compliance              *IndividualComplianceResponseDto `json:"compliance"`
	CreatedAt               time.Time                        `json:"createdAt"`
	DateOfBirth             *string                          `json:"dateOfBirth"`
	DocumentCountryCode     *string                          `json:"documentCountryCode"`
	DocumentExpiry          *string                          `json:"documentExpiry"`
	DocumentNumber          *string                          `json:"documentNumber"`
	DocumentType            *string                          `json:"documentType"`
	Email                   *string                          `json:"email"`
	FirstName               *string                          `json:"firstName"`
	Id                      string                           `json:"id"`
	IsSoleTrader            bool                             `json:"isSoleTrader"`
	KycVerificationLink     string                           `json:"kycVerificationLink"`
	LastName                *string                          `json:"lastName"`
	LivenessLink            string                           `json:"livenessLink"`
	MiddleName              *string                          `json:"middleName"`
	Nationalities           []string                         `json:"nationalities"`
	Phone                   *string                          `json:"phone"`
	PhysicalAddress         *PhysicalAddressResponseDto      `json:"physicalAddress"`
	PlaceOfBirthCity        *string                          `json:"placeOfBirthCity"`
	PlaceOfBirthCountryCode *string                          `json:"placeOfBirthCountryCode"`
	ResidenceCountryCode    *string                          `json:"residenceCountryCode"`
	TaxResidences           []TaxResidence                   `json:"taxResidences"`
	UpdatedAt               time.Time                        `json:"updatedAt"`
}

type TaxResidence struct {
	CountryCode string `json:"countryCode"`
	Tin         string `json:"tin"`
}
