package generated

// MasterAccount represents the master account to which the virtual IBAN is linked.
type MasterAccount string

const (
	MasterAccountCDA MasterAccount = "CDA"
	MasterAccountFDA MasterAccount = "FDA"
)

// Type represents the usage model: COBO = collect on behalf of; POBO = payment on behalf of.
type Type string

const (
	TypeCOBO Type = "COBO"
	TypePOBO Type = "POBO"
)

// ExpectedYearlyTxCount represents the expected yearly transaction count. Allowed values: LT_50 = < 50 transactions / year; B_50_500 = 50 - 500 transactions / year; B_500_5000 = 500 - 5,000 transactions / year; B_5000_50000 = 5,000 - 50,000 transactions / year; GT_50000 = > 50,000 transactions / year.
type ExpectedYearlyTxCount string

const (
	ExpectedYearlyTxCountLT_50        ExpectedYearlyTxCount = "LT_50"
	ExpectedYearlyTxCountB_50_500     ExpectedYearlyTxCount = "B_50_500"
	ExpectedYearlyTxCountB_500_5000   ExpectedYearlyTxCount = "B_500_5000"
	ExpectedYearlyTxCountB_5000_50000 ExpectedYearlyTxCount = "B_5000_50000"
	ExpectedYearlyTxCountGT_50000     ExpectedYearlyTxCount = "GT_50000"
)

// ExpectedYearlyVolume represents the expected yearly transaction volume. Allowed values: LT_100K = < €100,000; B_100K_1M = €100,000 - €1,000,000; B_1M_10M = €1,000,000 - €10,000,000; B_10M_50M = €10,000,000 - €50,000,000; B_50M_250M = €50,000,000 - €250,000,000; GT_250M = > €250,000,000.
type ExpectedYearlyVolume string

const (
	ExpectedYearlyVolumeLT_100K    ExpectedYearlyVolume = "LT_100K"
	ExpectedYearlyVolumeB_100K_1M  ExpectedYearlyVolume = "B_100K_1M"
	ExpectedYearlyVolumeB_1M_10M   ExpectedYearlyVolume = "B_1M_10M"
	ExpectedYearlyVolumeB_10M_50M  ExpectedYearlyVolume = "B_10M_50M"
	ExpectedYearlyVolumeB_50M_250M ExpectedYearlyVolume = "B_50M_250M"
	ExpectedYearlyVolumeGT_250M    ExpectedYearlyVolume = "GT_250M"
)

// FinancialAssets represents the total financial assets. Allowed values: LT_100K = < €100,000; B_100K_1M = €100,000 - €1,000,000; B_1M_10M = €1,000,000 - €10,000,000; B_10M_100M = €10,000,000 - €100,000,000; GT_100M = > €100,000,000; OTHER = Other.
type FinancialAssets string

const (
	FinancialAssetsLT_100K    FinancialAssets = "LT_100K"
	FinancialAssetsB_100K_1M  FinancialAssets = "B_100K_1M"
	FinancialAssetsB_1M_10M   FinancialAssets = "B_1M_10M"
	FinancialAssetsB_10M_100M FinancialAssets = "B_10M_100M"
	FinancialAssetsGT_100M    FinancialAssets = "GT_100M"
	FinancialAssetsOTHER      FinancialAssets = "OTHER"
)

type ProductsSubscribed string

const (
	ProductsSubscribedDIRECT_MINT_BURN ProductsSubscribed = "DIRECT_MINT_BURN"
	ProductsSubscribedEUR_ACCOUNT      ProductsSubscribed = "EUR_ACCOUNT"
	ProductsSubscribedFX_SWAP          ProductsSubscribed = "FX_SWAP"
)

// Purpose represents the purpose of the relationship. Allowed values: OPERATIONAL_ACCOUNT = Operational account; SALARY = Salary; INVESTMENT = Investment; TREASURY_MANAGEMENT = Treasury management; ON_OFF_RAMP = On/Off-ramp (fiat <-> crypto); INTERNATIONAL_PAYMENTS = International payments / remittance; OTHER = Other (not listed - please specify).
type Purpose string

const (
	PurposeOPERATIONAL_ACCOUNT    Purpose = "OPERATIONAL_ACCOUNT"
	PurposeSALARY                 Purpose = "SALARY"
	PurposeINVESTMENT             Purpose = "INVESTMENT"
	PurposeTREASURY_MANAGEMENT    Purpose = "TREASURY_MANAGEMENT"
	PurposeON_OFF_RAMP            Purpose = "ON_OFF_RAMP"
	PurposeINTERNATIONAL_PAYMENTS Purpose = "INTERNATIONAL_PAYMENTS"
	PurposeOTHER                  Purpose = "OTHER"
)

// SourceOfFunds represents the source of funds. Allowed values: BUSINESS_TRADING_REVENUE = Business / trading revenue; SHAREHOLDER_CAPITAL_EQUITY_CONTRIBUTION = Shareholder capital / equity contribution; INTRA_GROUP_FINANCING_PARENT_FUNDING = Intra-group financing / parent funding; BANK_LOAN_CREDIT_FACILITY = Bank loan / credit facility; INVESTOR_FUNDRAISING_PROCEEDS = Investor / fundraising proceeds (VC, PE); SALE_OF_ASSETS_BUSINESS = Sale of assets / business; CRYPTO_ASSET_ACTIVITY_PROCEEDS = Proceeds from crypto-asset activity; THIRD_PARTY_CLIENT_FUNDS = Third-party / client funds (not own funds); OTHER = Other (not listed).
type SourceOfFunds string

const (
	SourceOfFundsBUSINESS_TRADING_REVENUE                SourceOfFunds = "BUSINESS_TRADING_REVENUE"
	SourceOfFundsSHAREHOLDER_CAPITAL_EQUITY_CONTRIBUTION SourceOfFunds = "SHAREHOLDER_CAPITAL_EQUITY_CONTRIBUTION"
	SourceOfFundsINTRA_GROUP_FINANCING_PARENT_FUNDING    SourceOfFunds = "INTRA_GROUP_FINANCING_PARENT_FUNDING"
	SourceOfFundsBANK_LOAN_CREDIT_FACILITY               SourceOfFunds = "BANK_LOAN_CREDIT_FACILITY"
	SourceOfFundsINVESTOR_FUNDRAISING_PROCEEDS           SourceOfFunds = "INVESTOR_FUNDRAISING_PROCEEDS"
	SourceOfFundsSALE_OF_ASSETS_BUSINESS                 SourceOfFunds = "SALE_OF_ASSETS_BUSINESS"
	SourceOfFundsCRYPTO_ASSET_ACTIVITY_PROCEEDS          SourceOfFunds = "CRYPTO_ASSET_ACTIVITY_PROCEEDS"
	SourceOfFundsTHIRD_PARTY_CLIENT_FUNDS                SourceOfFunds = "THIRD_PARTY_CLIENT_FUNDS"
	SourceOfFundsOTHER                                   SourceOfFunds = "OTHER"
)

// TurnoverBracket represents the annual turnover. Allowed values: LT_100K = < €100,000; B_100K_1M = €100,000 - €1,000,000; B_1M_10M = €1,000,000 - €10,000,000; B_10M_50M = €10,000,000 - €50,000,000; B_50M_250M = €50,000,000 - €250,000,000; GT_250M = > €250,000,000; OTHER = Other.
type TurnoverBracket string

const (
	TurnoverBracketLT_100K    TurnoverBracket = "LT_100K"
	TurnoverBracketB_100K_1M  TurnoverBracket = "B_100K_1M"
	TurnoverBracketB_1M_10M   TurnoverBracket = "B_1M_10M"
	TurnoverBracketB_10M_50M  TurnoverBracket = "B_10M_50M"
	TurnoverBracketB_50M_250M TurnoverBracket = "B_50M_250M"
	TurnoverBracketGT_250M    TurnoverBracket = "GT_250M"
	TurnoverBracketOTHER      TurnoverBracket = "OTHER"
)

// Direction represents the sEND = outbound; RECEIVE = inbound; BOTH = inbound and outbound.
type Direction string

const (
	DirectionSEND    Direction = "SEND"
	DirectionRECEIVE Direction = "RECEIVE"
	DirectionBOTH    Direction = "BOTH"
)

type RequestBody struct {
	ExternalId string `json:"externalId,omitempty"`
	Name       string `json:"name,omitempty"`
}

type CustomerDetailCompanyResponseDto struct{}

type Response201 struct {
	CustomerDetailCompanyResponseDto
	Company Company       `json:"company,omitempty"`
	Vibans  []VibansEntry `json:"vibans,omitempty"`
}

type CompanyDetailResponseDto struct{}

type Company struct {
	CompanyDetailResponseDto
	Compliance *Compliance `json:"compliance,omitempty"`
}

type VibansEntry struct {
	Alias         *string       `json:"alias"`
	BcOrderId     *string       `json:"bcOrderId"`
	CreatedAt     DateTime      `json:"createdAt"`
	CustomerId    string        `json:"customerId"`
	DeletedAt     *DateTime     `json:"deletedAt"`
	Iban          *string       `json:"iban"`
	Id            string        `json:"id"`
	IsActive      bool          `json:"isActive"`
	MasterAccount MasterAccount `json:"masterAccount"`
	Type          Type          `json:"type"`
	UpdatedAt     DateTime      `json:"updatedAt"`
}

type Compliance struct {
	ExpectedYearlyTxCount ExpectedYearlyTxCount  `json:"expectedYearlyTxCount,omitempty"`
	ExpectedYearlyVolume  ExpectedYearlyVolume   `json:"expectedYearlyVolume,omitempty"`
	FinancialAssets       FinancialAssets        `json:"financialAssets,omitempty"`
	FinancialAssetsOther  string                 `json:"financialAssetsOther,omitempty"`
	GeographicScope       []GeographicScopeEntry `json:"geographicScope,omitempty"`
	ProductsSubscribed    []ProductsSubscribed   `json:"productsSubscribed,omitempty"`
	Purpose               Purpose                `json:"purpose,omitempty"`
	PurposeDetails        string                 `json:"purposeDetails,omitempty"`
	SourceOfFunds         SourceOfFunds          `json:"sourceOfFunds,omitempty"`
	SourceOfFundsDetails  string                 `json:"sourceOfFundsDetails,omitempty"`
	TurnoverBracket       TurnoverBracket        `json:"turnoverBracket,omitempty"`
	TurnoverBracketOther  string                 `json:"turnoverBracketOther,omitempty"`
	Website               string                 `json:"website,omitempty"`
}

type GeographicScopeEntry struct {
	CountryCode string    `json:"countryCode"`
	Direction   Direction `json:"direction"`
}

type Response400 struct{}

func (r Response400) Error() string {
	panic("TODO: define the output")
}

type Response401 struct{}

func (r Response401) Error() string {
	panic("TODO: define the output")
}
