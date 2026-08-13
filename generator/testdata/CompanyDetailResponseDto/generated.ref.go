package generated

import "github.com/artem-kuznetsov-intellectsoft/openapi2go/openapi"

type Capacity string

const (
	CapacityOWN_ACCOUNT       Capacity = "OWN_ACCOUNT"
	CapacityTHIRD_PARTY_FUNDS Capacity = "THIRD_PARTY_FUNDS"
)

// CrsClassification represents the cRS classification (ACTIVE_NFE / PASSIVE_NFE / FINANCIAL_INSTITUTION).
type CrsClassification string

const (
	CrsClassificationACTIVE_NFE            CrsClassification = "ACTIVE_NFE"
	CrsClassificationPASSIVE_NFE           CrsClassification = "PASSIVE_NFE"
	CrsClassificationFINANCIAL_INSTITUTION CrsClassification = "FINANCIAL_INSTITUTION"
)

// LegalForm represents the legal form code; OTHER falls back to `legalFormOther`.
type LegalForm string

const (
	LegalFormLISTED_PUBLIC_COMPANY           LegalForm = "LISTED_PUBLIC_COMPANY"
	LegalFormREGULATED_FINANCIAL_INSTITUTION LegalForm = "REGULATED_FINANCIAL_INSTITUTION"
	LegalFormREGULATED_INVESTMENT_FUND       LegalForm = "REGULATED_INVESTMENT_FUND"
	LegalFormPRIVATE_LIMITED_COMPANY         LegalForm = "PRIVATE_LIMITED_COMPANY"
	LegalFormPUBLIC_LIMITED_COMPANY          LegalForm = "PUBLIC_LIMITED_COMPANY"
	LegalFormCOOPERATIVE_MUTUAL_ENTITY       LegalForm = "COOPERATIVE_MUTUAL_ENTITY"
	LegalFormPARTNERSHIP_CIVIL_COMPANY       LegalForm = "PARTNERSHIP_CIVIL_COMPANY"
	LegalFormFOREIGN_BRANCH                  LegalForm = "FOREIGN_BRANCH"
	LegalFormHOLDING_SPV                     LegalForm = "HOLDING_SPV"
	LegalFormTRUST_OR_SIMILAR_ARRANGEMENT    LegalForm = "TRUST_OR_SIMILAR_ARRANGEMENT"
	LegalFormFOUNDATION_OR_ENDOWMENT         LegalForm = "FOUNDATION_OR_ENDOWMENT"
	LegalFormNON_PROFIT_ASSOCIATION_NGO      LegalForm = "NON_PROFIT_ASSOCIATION_NGO"
	LegalFormGOVERNMENT_PUBLIC_AUTHORITY     LegalForm = "GOVERNMENT_PUBLIC_AUTHORITY"
	LegalFormSTATE_OWNED_ENTERPRISE          LegalForm = "STATE_OWNED_ENTERPRISE"
	LegalFormOTHER                           LegalForm = "OTHER"
	LegalFormSOLE_PROPRIETORSHIP             LegalForm = "SOLE_PROPRIETORSHIP"
)

// PreferredWalletChain represents the preferred wallet chain for incoming settlements.
type PreferredWalletChain string

const (
	PreferredWalletChainRIPPLE    PreferredWalletChain = "RIPPLE"
	PreferredWalletChainETHEREUM  PreferredWalletChain = "ETHEREUM"
	PreferredWalletChainPOLYGON   PreferredWalletChain = "POLYGON"
	PreferredWalletChainAVALANCHE PreferredWalletChain = "AVALANCHE"
	PreferredWalletChainPLASMA    PreferredWalletChain = "PLASMA"
	PreferredWalletChainSOLANA    PreferredWalletChain = "SOLANA"
	PreferredWalletChainP_LAYER   PreferredWalletChain = "P_LAYER"
)

// Sector represents the nACE level-1 sector; OTHER falls back to `sectorOther`.
type Sector string

const (
	SectorAGRICULTURE_FORESTRY_FISHING            Sector = "AGRICULTURE_FORESTRY_FISHING"
	SectorMINING_QUARRYING_EXTRACTIVES            Sector = "MINING_QUARRYING_EXTRACTIVES"
	SectorMANUFACTURING_GENERAL                   Sector = "MANUFACTURING_GENERAL"
	SectorTEXTILES_CLOTHING_LEATHER_MANUFACTURING Sector = "TEXTILES_CLOTHING_LEATHER_MANUFACTURING"
	SectorENERGY_UTILITIES                        Sector = "ENERGY_UTILITIES"
	SectorCONSTRUCTION_PUBLIC_WORKS               Sector = "CONSTRUCTION_PUBLIC_WORKS"
	SectorWHOLESALE_RETAIL_GENERAL                Sector = "WHOLESALE_RETAIL_GENERAL"
	SectorPRECIOUS_METALS_JEWELLERY_ART_ANTIQUES  Sector = "PRECIOUS_METALS_JEWELLERY_ART_ANTIQUES"
	SectorMOTOR_VEHICLE_BOAT_LUXURY_GOODS         Sector = "MOTOR_VEHICLE_BOAT_LUXURY_GOODS"
	SectorTRANSPORT_LOGISTICS                     Sector = "TRANSPORT_LOGISTICS"
	SectorPOSTAL_COURIER_VALUE_TRANSFER           Sector = "POSTAL_COURIER_VALUE_TRANSFER"
	SectorACCOMMODATION_FOOD_SERVICE              Sector = "ACCOMMODATION_FOOD_SERVICE"
	SectorINFORMATION_IT_COMMUNICATION            Sector = "INFORMATION_IT_COMMUNICATION"
	SectorFINANCIAL_INSURANCE_SERVICES            Sector = "FINANCIAL_INSURANCE_SERVICES"
	SectorMONEY_SERVICES_CURRENCY_EXCHANGE        Sector = "MONEY_SERVICES_CURRENCY_EXCHANGE"
	SectorCRYPTO_ASSET_SERVICES                   Sector = "CRYPTO_ASSET_SERVICES"
	SectorREAL_ESTATE                             Sector = "REAL_ESTATE"
	SectorLEGAL_ACCOUNTING_AUDIT_ADVISORY         Sector = "LEGAL_ACCOUNTING_AUDIT_ADVISORY"
	SectorOTHER_PROFESSIONAL_SCIENTIFIC_TECHNICAL Sector = "OTHER_PROFESSIONAL_SCIENTIFIC_TECHNICAL"
	SectorADMINISTRATIVE_SUPPORT_SERVICES         Sector = "ADMINISTRATIVE_SUPPORT_SERVICES"
	SectorGAMBLING_BETTING_CASINOS                Sector = "GAMBLING_BETTING_CASINOS"
	SectorARTS_ENTERTAINMENT_RECREATION_OTHER     Sector = "ARTS_ENTERTAINMENT_RECREATION_OTHER"
	SectorARMS_DEFENCE_DUAL_USE_GOODS             Sector = "ARMS_DEFENCE_DUAL_USE_GOODS"
	SectorPUBLIC_ADMIN_EDUCATION_HEALTH_SOCIAL    Sector = "PUBLIC_ADMIN_EDUCATION_HEALTH_SOCIAL"
	SectorPERSONAL_HOUSEHOLD_SERVICES             Sector = "PERSONAL_HOUSEHOLD_SERVICES"
	SectorOTHER                                   Sector = "OTHER"
)

type CompanyComplianceResponseDto struct{}

type DocumentResponseDto struct{}

type CompanyIndividualResponseDto struct{}

type PhysicalAddressResponseDto struct{}

type TcAcceptanceResponseDto struct{}

type CompanyUboResponseDto struct{}

// CompanyDetailResponseDto is generated from components.schemas.CompanyDetailResponseDto.
type CompanyDetailResponseDto struct {
	ActivityDescription     map[string]any                 `json:"activityDescription"`
	BrandName               map[string]any                 `json:"brandName"`
	Capacity                *Capacity                      `json:"capacity"`
	Compliance              *CompanyComplianceResponseDto  `json:"compliance"`
	CreatedAt               openapi.DateTime               `json:"createdAt"`
	CrsClassification       *CrsClassification             `json:"crsClassification"`
	Documents               []DocumentResponseDto          `json:"documents"`
	Email                   map[string]any                 `json:"email"`
	Id                      string                         `json:"id"`
	JurisdictionCountryCode map[string]any                 `json:"jurisdictionCountryCode"`
	KybVerificationLink     string                         `json:"kybVerificationLink"`
	LegalForm               *LegalForm                     `json:"legalForm"`
	LegalFormOther          map[string]any                 `json:"legalFormOther"`
	LegalRepresentatives    []CompanyIndividualResponseDto `json:"legalRepresentatives"`
	Name                    map[string]any                 `json:"name"`
	Phone                   map[string]any                 `json:"phone"`
	PhysicalAddress         *PhysicalAddressResponseDto    `json:"physicalAddress"`
	PreferredBankAccount    map[string]any                 `json:"preferredBankAccount"`
	PreferredWalletAddress  map[string]any                 `json:"preferredWalletAddress"`
	PreferredWalletChain    *PreferredWalletChain          `json:"preferredWalletChain"`
	RegistrationDate        map[string]any                 `json:"registrationDate"`
	RegistrationNumber      map[string]any                 `json:"registrationNumber"`
	Sector                  *Sector                        `json:"sector"`
	SectorOther             map[string]any                 `json:"sectorOther"`
	TaxId                   map[string]any                 `json:"taxId"`
	TaxResidenceCountryCode map[string]any                 `json:"taxResidenceCountryCode"`
	TcAcceptances           []TcAcceptanceResponseDto      `json:"tcAcceptances"`
	Tin                     map[string]any                 `json:"tin"`
	Ubos                    []CompanyUboResponseDto        `json:"ubos"`
	UpdatedAt               openapi.DateTime               `json:"updatedAt"`
	Website                 map[string]any                 `json:"website"`
}
