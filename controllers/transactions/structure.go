package transactions

func strPtr(s string) *string { return &s }

// Matches transaction-detail metadata output (camelCase keys)
type FormMetadataDto struct {
	Name              string  `json:"name"`
	Type              string  `json:"type"`
	CustomType        *string `json:"customType"`
	Label             string  `json:"label"`
	BRequired         bool    `json:"bRequired"`
	BRemoteDataSource bool    `json:"bRemoteDataSource"`
	DataSource        *string `json:"dataSource"`
	Header            *string `json:"header"`
	OrderNumber       int     `json:"orderNumber"`
	BVisible          bool    `json:"bVisible"`
	BSortable         bool    `json:"bSortable"`
	Editable          bool    `json:"editable"`
	BFilterable       bool    `json:"bFilterable"`
}

// This is the metadata that .NET generates via attributes on TransactionDetailsViewDto.
// (Only properties decorated with [FormMetadata] appear.)
var TransactionDetailsMetadata = []FormMetadataDto{
	{
		Name:     "TransactionReference",
		Type:     "String",
		Label:    "Transaction Reference",
		BVisible: true,
	},
	{
		Name:     "CustomerAccountDescription",
		Type:     "String",
		Label:    "Customer Account",
		BVisible: false,
	},
	{
		Name:     "Date",
		Type:     "DateTime",
		Label:    "Transaction Date",
		BVisible: false,
	},
	{
		Name:       "PaymentReleaseDate",
		Type:       "Custom",
		Label:      "Payment Release Date",
		CustomType: strPtr("PaymentReleaseDate"),
		BVisible:   false,
	},
	{
		Name:     "ExternalId",
		Type:     "String",
		Label:    "External ID",
		BVisible: true,
	},
	{
		Name:     "Amount",
		Type:     "Decimal",
		Label:    "Amount",
		BVisible: true,
	},
	{
		Name:     "Description",
		Type:     "String",
		Label:    "Description",
		BVisible: true,
	},
	{
		Name:     "AssetCode",
		Type:     "String",
		Label:    "Asset Code",
		BVisible: true,
	},
	{
		Name:     "AssetClass",
		Type:     "String",
		Label:    "Asset Class",
		BVisible: false,
	},
	{
		Name:     "TransactionType",
		Type:     "String",
		Label:    "Transaction Type",
		BVisible: true,
	},
	{
		Name:     "Product",
		Type:     "String",
		Label:    "Product",
		BVisible: true,
	},
	{
		Name:     "Direction",
		Type:     "String",
		Label:    "Direction",
		BVisible: true,
	},
	{
		Name:     "PendingComplianceReview",
		Type:     "String",
		Label:    "Pending Compliance Review",
		BVisible: true,
	},
	{
		Name:     "HasFee",
		Type:     "String",
		Label:    "Has Fee",
		BVisible: true,
	},
	{
		Name:     "Rejected",
		Type:     "String",
		Label:    "Rejected",
		BVisible: true,
	},
	{
		Name:     "Reversed",
		Type:     "String",
		Label:    "Reversed",
		BVisible: true,
	},
	{
		Name:     "PaymentType",
		Type:     "String",
		Label:    "Payment Type",
		BVisible: true,
	},
	{
		Name:     "ScreeningAlertsAvailable",
		Type:     "String",
		Label:    "Screening Alerts Available",
		BVisible: true,
	},
	{
		Name:     "FeeIncurringAction",
		Type:     "String",
		Label:    "Fee Incurring Action",
		BVisible: true,
	},
	{
		Name:     "Status",
		Type:     "String",
		Label:    "Status",
		BVisible: true,
	},
	{
		Name:     "OriginatorName",
		Type:     "String",
		Label:    "Originator Name",
		BVisible: true,
	},
	{
		Name:     "BeneficiaryName",
		Type:     "String",
		Label:    "Beneficiary Name",
		BVisible: true,
	},
	{
		Name:     "International",
		Type:     "String",
		Label:    "International",
		BVisible: true,
	},
	{
		Name:     "ImadNumber",
		Type:     "String",
		Label:    "Imad Number",
		BVisible: true,
	},
	{
		Name:     "BankReference",
		Type:     "String",
		Label:    "Bank ID",
		BVisible: true,
	},
	{
		Name:     "Memo",
		Type:     "String",
		Label:    "Memo",
		BVisible: true,
	},
	{
		Name:     "BeneficiaryReference",
		Type:     "String",
		Label:    "Beneficiary Reference",
		BVisible: true,
	},
	{
		Name:       "PayeeDetails",
		Type:       "Custom",
		Label:      "Payee Details",
		CustomType: strPtr("PayeeDetails"),
		BVisible:   true,
	},
	{
		Name:       "TransmitterDetails",
		Type:       "Custom",
		Label:      "Transmitter Details",
		CustomType: strPtr("TransmitterDetails"),
		BVisible:   true,
	},
	{
		Name:       "CryptoTransferDetails",
		Type:       "Custom",
		Label:      "Crypto Transfer Details",
		CustomType: strPtr("CryptoTransferDetails"),
		BVisible:   true,
	},
	{
		Name:  "ApprovalDetails",
		Type:  "Custom",
		Label: "Approval Details",
		// NOTE: .NET me yahan CustomType = "CryptoTransferDetails" likha hua hai (as-is)
		CustomType: strPtr("CryptoTransferDetails"),
		BVisible:   true,
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

type pendingProductsResponse struct {
	Id      interface{}          `json:"id"`
	Details []pendingProductsOut `json:"details"`
	Status  interface{}          `json:"status"`
	Errors  []interface{}        `json:"errors"`
}

type pendingProductsRaw struct {
	ProductId               int     `json:"productId"`
	ProductName             string  `json:"productName"`
	Asset                   string  `json:"asset"`
	DisplayName             string  `json:"displayName"`
	Amount                  float64 `json:"amount"`
	PendingTransactionCount int     `json:"pendingTransactionCount"`
}

type pendingProductsOut struct {
	ProductId               int     `json:"productId"`
	ProductName             string  `json:"productName"`
	Asset                   string  `json:"asset"`
	DisplayName             string  `json:"displayName"`
	Amount                  float64 `json:"amount"`
	PendingTransactionCount int     `json:"pendingTransactionCount"`
}

type spNotesEnvelope struct {
	Notes []spNoteItem `json:"Notes"`
}

type spNoteItem struct {
	Id        int    `json:"Id"`
	Text      string `json:"Text"`
	AddDate   string `json:"AddDate"`
	AddedBy   string `json:"AddedBy"`
	BPinned   bool   `json:"bPinned"`
	BEditable bool   `json:"bEditable"`
}

type apiNoteItem struct {
	Id        string `json:"id"`
	Text      string `json:"text"`
	AddDate   string `json:"addDate"`
	AddedBy   string `json:"addedBy"`
	BEditable bool   `json:"bEditable"`
	BPinned   bool   `json:"bPinned"`
}

type transactionJSONResponse struct {
	Id      interface{}   `json:"id"`
	Details interface{}   `json:"details"` // string OR nil
	Status  interface{}   `json:"status"`
	Errors  []interface{} `json:"errors"`
}

type transactionNote struct {
	Id        string `json:"id"`
	Text      string `json:"text"`
	AddDate   string `json:"addDate"`
	AddedBy   string `json:"addedBy"`
	BEditable bool   `json:"bEditable"`
	BPinned   bool   `json:"bPinned"`
}

type addNoteResponse struct {
	Id      interface{}   `json:"id"`
	Details interface{}   `json:"details"`
	Status  interface{}   `json:"status"`
	Errors  []interface{} `json:"errors"`
}

type addNoteDetails struct {
	TransactionsId int64  `json:"transactionsId"`
	Text           string `json:"text"`
}

type editNoteResponse struct {
	Id      interface{}   `json:"id"`
	Details interface{}   `json:"details"`
	Status  interface{}   `json:"status"`
	Errors  []interface{} `json:"errors"`
}

type editNoteDetails struct {
	Id        string `json:"id"`
	Text      string `json:"text"`
	BEditable bool   `json:"bEditable"`
}

type pinNoteRequest struct {
	Id      string `json:"id"`
	BPinned bool   `json:"bPinned"`
}

type pinNoteDetails struct {
	Id      string `json:"id"`
	BPinned bool   `json:"bPinned"`
}
