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
	Id      interface{}      `json:"id"`
	Details []PendingProduct `json:"details"`
	Status  interface{}      `json:"status"`
	Errors  []interface{}    `json:"errors"`
}

// PendingProduct represents a pending product item
// (Consolidated from duplicate pendingProductsRaw and pendingProductsOut types)
type PendingProduct struct {
	ProductId               int     `json:"productId"`
	ProductName             string  `json:"productName"`
	Asset                   string  `json:"asset"`
	DisplayName             string  `json:"displayName"`
	Amount                  float64 `json:"amount"`
	PendingTransactionCount int     `json:"pendingTransactionCount"`
}

// Type aliases for backward compatibility
type pendingProductsRaw = PendingProduct
type pendingProductsOut = PendingProduct

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

type ReleaseRequest struct {
	CustomerAssetAccountsTransactionsId int64  `json:"customerAssetAccountsTransactionsId"`
	TransactionDetailsJson              string `json:"transactionDetailsJson,omitempty"`
}

type ReleaseDetails struct {
	CustomerAssetAccountsTransactionsId int64 `json:"customerAssetAccountsTransactionsId"`
}

// .NET-style validation error (fieldName + messageCode)
type FieldError struct {
	FieldName   string `json:"fieldName"`
	MessageCode string `json:"messageCode"`
}

// =====================================================
// TransactionsModule - Cancellation DTOs
// =====================================================

// GetCancellationEligibilityRequest - GET /transactionsmodule/cancel
type GetCancellationEligibilityRequest struct {
	CustomerAssetAccountsTransactionsId int `form:"customerAssetAccountsTransactionsId" json:"customerAssetAccountsTransactionsId"`
}

// CancellationEligibilityResponse - response for GET /transactionsmodule/cancel
type CancellationEligibilityResponse struct {
	BCanBeCancelled bool   `json:"bCanBeCancelled"`
	Message         string `json:"message,omitempty"`
}

// CancelTransactionRequest - POST /transactionsmodule/cancel
type CancelTransactionRequest struct {
	CustomerAssetAccountsTransactionsId int `json:"customerAssetAccountsTransactionsId"`
}

// CancelTransactionResponse - response for POST /transactionsmodule/cancel
type CancelTransactionResponse struct {
	CustomerTransactionCancellationRequestsId int `json:"customerTransactionCancellationRequestsId"`
	CustomerAssetAccountsTransactionsId       int `json:"customerAssetAccountsTransactionsId"`
}

// =====================================================
// TransactionsModule - WaiveFee DTOs
// =====================================================

// WaiveFeeRequest - POST /transactionsmodule/waive-fee
type WaiveFeeRequest struct {
	CustomerAssetAccountsTransactionsId int      `json:"customerAssetAccountsTransactionsId"`
	Amount                              float64  `json:"amount"`
	AssetSymbol                         string   `json:"assetSymbol,omitempty"`
	PercentageAmount                    *float64 `json:"percentageAmount,omitempty"`
	FixedAmount                         *float64 `json:"fixedAmount,omitempty"`
	BFullRefund                         bool     `json:"bFullRefund"`
}

// WaiveFeeResponse - response for waive-fee endpoints
type WaiveFeeResponse struct {
	CustomerAssetAccountsTransactionsId int     `json:"customerAssetAccountsTransactionsId"`
	Amount                              float64 `json:"amount"`
	AssetSymbol                         string  `json:"assetSymbol,omitempty"`
	AmountWaived                        float64 `json:"amountWaived,omitempty"`
}

// =====================================================
// TransactionsModule - Document DTOs
// =====================================================

// DocumentListItem represents a document in the list
type DocumentListItem struct {
	Id          int    `json:"id"`
	Text        string `json:"text"`
	FileName    string `json:"fileName"`
	AddDate     string `json:"addDate"`
	AddedBy     string `json:"addedBy"`
	BPinned     bool   `json:"bPinned"`
	BEditable   bool   `json:"bEditable"`
	ContentType string `json:"contentType,omitempty"`
}

// DocumentListResponse - response for GET /transactionsmodule/list-documents
type DocumentListResponse struct {
	Documents []DocumentListItem `json:"documents"`
}

// EditDocumentRequest - POST /transactionsmodule/edit-document
type EditDocumentRequest struct {
	Id   int    `json:"id" form:"id"`
	Text string `json:"text" form:"text"`
}

// PinDocumentRequest - POST /transactionsmodule/pin-document
type PinDocumentRequest struct {
	Id      int  `json:"id"`
	BPinned bool `json:"bPinned"`
}

// GetEditNoteResponse - response for GET /transactionsmodule/edit-note
type GetEditNoteResponse struct {
	Id        string `json:"id"`
	Text      string `json:"text"`
	BEditable bool   `json:"bEditable"`
}

// =====================================================
// PendingTransactionsModule - DTOs
// =====================================================

// AssignFormResponse - response for GET /pendingtransactionsmodule/assign
type AssignFormResponse struct {
	CustomerAssetAccountsTransactionsId int64          `json:"customerAssetAccountsTransactionsId"`
	AssignedUsers                       []int          `json:"assignedUsers"`
	ListAdminUsers                      []DropDownItem `json:"listAdminUsers"`
}

// DropDownItem - generic dropdown item
type DropDownItem struct {
	Value int    `json:"value"`
	Label string `json:"label"`
}

// AssignFormRequest - POST /pendingtransactionsmodule/assign
type AssignFormRequest struct {
	CustomerAssetAccountsTransactionsId int64 `json:"customerAssetAccountsTransactionsId"`
	AssignedUsers                       []int `json:"assignedUsers"`
}

// CompleteRequest - POST /pendingtransactionsmodule/complete
type CompleteRequest struct {
	CustomerAssetAccountsTransactionsId int64  `json:"customerAssetAccountsTransactionsId"`
	TransactionDetailsJson              string `json:"transactionDetailsJson,omitempty"`
}

// CompleteResponse - response for POST /pendingtransactionsmodule/complete
type CompleteResponse struct {
	CustomerAssetAccountsTransactionsId int64 `json:"customerAssetAccountsTransactionsId"`
}

// PendingCancelRequest - POST /pendingtransactionsmodule/cancel
type PendingCancelRequest struct {
	CustomerAssetAccountsTransactionsId int64  `json:"customerAssetAccountsTransactionsId"`
	TransactionDetailsJson              string `json:"transactionDetailsJson,omitempty"`
	BRefundFee                          bool   `json:"bRefundFee"`
	BCancelled                          bool   `json:"bCancelled"`
}

// PendingCancelResponse - response for POST /pendingtransactionsmodule/cancel
type PendingCancelResponse struct {
	CustomerTransactionCancellationRequestsId int   `json:"customerTransactionCancellationRequestsId"`
	CustomerAssetAccountsTransactionsId       int64 `json:"customerAssetAccountsTransactionsId"`
}

// MarkAsPendingRequest - POST /pendingtransactionsmodule/markaspending
type MarkAsPendingRequest struct {
	CustomerAssetAccountsTransactionsId int64  `json:"customerAssetAccountsTransactionsId"`
	ExternalID                          string `json:"externalID,omitempty"`
	TransactionDetailsJson              string `json:"transactionDetailsJson,omitempty"`
}

// =====================================================
// FrozenTransactionsModule - DTOs
// =====================================================

// FrozenTransactionView - response for GET /frozentransactionsmodule/view
type FrozenTransactionView struct {
	TRMLabsHelperId          int     `json:"trmLabsHelperId"`
	Risk                     float64 `json:"risk,omitempty"`
	RiskLabel                string  `json:"riskLabel,omitempty"`
	ScreenStatus             string  `json:"screenStatus,omitempty"`
	ScreenStatusFailedReason string  `json:"screenStatusFailedReason,omitempty"`
	TransactionHash          string  `json:"transactionHash,omitempty"`
	TransactionUUID          string  `json:"transactionUUID,omitempty"`
	Type                     string  `json:"type,omitempty"`
	AddDate                  string  `json:"addDate,omitempty"`
}

// FrozenViewMetadata - metadata for frozen transaction view form
var FrozenViewMetadata = []FormMetadataDto{
	{Name: "risk", Type: "Integer", Label: "Risk", BVisible: true},
	{Name: "riskLabel", Type: "String", Label: "Risk Label", BVisible: true},
	{Name: "screenStatus", Type: "String", Label: "Screen Status", BVisible: true},
	{Name: "screenStatusFailedReason", Type: "String", Label: "Screen Status Failed Reason", BVisible: true},
	{Name: "transactionHash", Type: "String", Label: "Transaction Hash", BVisible: true},
	{Name: "transactionUUID", Type: "String", Label: "Transaction UUID", BVisible: true},
	{Name: "type", Type: "String", Label: "Type", BVisible: true},
	{Name: "addDate", Type: "DateTime", Label: "Date of Query", BVisible: true},
}

// UnfreezeRequest - POST /frozentransactionsmodule/unfreeze
type UnfreezeRequest struct {
	CustomerTransactionID int `json:"customerTransactionID"`
}

// =====================================================
// PendingTransactionsTreasuryModule - DTOs
// =====================================================

// TreasuryCancelRequest - POST /pendingtransactionstreasurymodule/cancel
type TreasuryCancelRequest struct {
	CustomerAssetAccountsTransactionsId int64  `json:"customerAssetAccountsTransactionsId"`
	TransactionDetailsJson              string `json:"transactionDetailsJson,omitempty"`
	BRefundFee                          bool   `json:"bRefundFee"`
	BCancelled                          bool   `json:"bCancelled"`
}

// TreasuryCancelResponse - response for POST /pendingtransactionstreasurymodule/cancel
type TreasuryCancelResponse struct {
	CustomerTransactionCancellationRequestsId int   `json:"customerTransactionCancellationRequestsId"`
	CustomerAssetAccountsTransactionsId       int64 `json:"customerAssetAccountsTransactionsId"`
}
