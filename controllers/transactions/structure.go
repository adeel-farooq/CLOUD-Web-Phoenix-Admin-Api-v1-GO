package transactions

// Matches .NET FormMetadataDto output (PascalCase keys)
type FormMetadataDto struct {
	Name              string `json:"Name"`
	Type              string `json:"Type"`
	CustomType        string `json:"CustomType,omitempty"`
	Label             string `json:"Label"`
	BRequired         bool   `json:"bRequired"`
	BRemoteDataSource *bool  `json:"bRemoteDataSource,omitempty"`
	DataSource        string `json:"DataSource,omitempty"`
	Header            string `json:"Header,omitempty"`
	OrderNumber       int    `json:"OrderNumber"`
	BVisible          bool   `json:"bVisible"`
	BSortable         bool   `json:"bSortable"`
	Editable          bool   `json:"Editable"`
	BFilterable       bool   `json:"bFilterable"`
}

// This is the metadata that .NET generates via attributes on TransactionDetailsViewDto.
// (Only properties decorated with [FormMetadata] appear.)
var TransactionDetailsMetadata = []FormMetadataDto{
	{
		Name:       "TransactionRail",
		Type:       "String",
		Label:      "Transaction Rail",
		CustomType: "TransactionRail",
		BVisible:   true,
	},
	{
		Name:       "Direction",
		Type:       "String",
		Label:      "Direction",
		CustomType: "TransactionDirection",
		BVisible:   true,
	},
	{
		Name:       "Status",
		Type:       "String",
		Label:      "Status",
		CustomType: "TransactionStatus",
		BVisible:   true,
	},
	{
		Name:     "TransactionType",
		Type:     "String",
		Label:    "Transaction Type",
		BVisible: true,
	},
	{
		Name:     "Amount",
		Type:     "Decimal",
		Label:    "Amount",
		BVisible: true,
	},
	{
		Name:     "Balance",
		Type:     "Decimal",
		Label:    "Balance",
		BVisible: true,
	},
	{
		Name:     "Asset",
		Type:     "String",
		Label:    "Asset",
		BVisible: true,
	},
	{
		Name:     "Product",
		Type:     "String",
		Label:    "Product",
		BVisible: true,
	},
	{
		Name:     "CustomerName",
		Type:     "String",
		Label:    "Customer Name",
		BVisible: true,
	},
	{
		Name:     "Reference",
		Type:     "String",
		Label:    "Reference",
		BVisible: true,
	},
	{
		Name:     "BankReference",
		Type:     "String",
		Label:    "Bank Reference",
		BVisible: true,
	},
	{
		Name:     "ImadNumber",
		Type:     "String",
		Label:    "Imad Number",
		BVisible: true,
	},
	{
		Name:     "Date",
		Type:     "DateTime",
		Label:    "Date",
		BVisible: true,
	},
	{
		Name:       "Rejected",
		Type:       "String",
		Label:      "Rejected",
		CustomType: "RejectionStatus",
		BVisible:   true,
	},
	{
		Name:     "AlertsAvailable",
		Type:     "Boolean",
		Label:    "Alerts Available",
		BVisible: true,
	},
	{
		Name:     "TransactionJson",
		Type:     "String",
		Label:    "Transaction Json",
		BVisible: true,
	},
}

var ViewAlertsMetadata = []FormMetadataDto{
	{
		Name:        "TransactionAlertSummary",
		Type:        "Custom",
		Label:       "Risk",
		OrderNumber: 1,
		BVisible:    true,
		BSortable:   false,
		BFilterable: false,
		BRequired:   false,
		Editable:    false,
	},
	{
		Name:        "TransactionAlerts",
		Type:        "Custom",
		Label:       "Transaction Alerts",
		OrderNumber: 2,
		BVisible:    true,
		BSortable:   false,
		BFilterable: false,
		BRequired:   false,
		Editable:    false,
	},
}
