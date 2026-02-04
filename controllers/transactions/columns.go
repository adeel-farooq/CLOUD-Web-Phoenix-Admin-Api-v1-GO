package transactions

// =============================================================================
// REFACTORED COLUMN DEFINITIONS
// This file replaces columns.go with a builder pattern to eliminate duplication
// =============================================================================

// ColumnDef represents a single column definition
type ColumnDef struct {
	ColumnKey      string
	LabelKey       string
	LabelValue     string
	OrderNumber    int
	BSortable      bool
	BFilterable    bool
	BVisible       bool
	BLocked        bool
	Type           string
	FilterMetadata interface{}
	Tooltip        interface{}
}

// ToMap converts ColumnDef to the map format expected by the API
func (c ColumnDef) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"columnKey":      c.ColumnKey,
		"labelKey":       c.LabelKey,
		"labelValue":     c.LabelValue,
		"orderNumber":    c.OrderNumber,
		"bSortable":      c.BSortable,
		"bFilterable":    c.BFilterable,
		"bVisible":       c.BVisible,
		"bLocked":        c.BLocked,
		"type":           c.Type,
		"filterMetadata": c.FilterMetadata,
		"tooltip":        c.Tooltip,
	}
}

// ColumnsToMaps converts a slice of ColumnDef to slice of maps
func ColumnsToMaps(cols []ColumnDef) []map[string]interface{} {
	result := make([]map[string]interface{}, len(cols))
	for i, col := range cols {
		col.OrderNumber = i + 1 // Auto-assign order numbers
		result[i] = col.ToMap()
	}
	return result
}

// =============================================================================
// FILTER METADATA HELPERS
// =============================================================================

func FilterTextContains() map[string]interface{} {
	return map[string]interface{}{"filterType": "TextContains", "details": nil}
}

func FilterTextContainsWithDetails() map[string]interface{} {
	return map[string]interface{}{"filterType": "TextContains", "details": map[string]interface{}{}}
}

func FilterAmount() map[string]interface{} {
	return map[string]interface{}{"filterType": "Amount", "details": nil}
}

func FilterAmountWithDetails() map[string]interface{} {
	return map[string]interface{}{"filterType": "Amount", "details": map[string]interface{}{}}
}

func FilterDateTimeRange() map[string]interface{} {
	return map[string]interface{}{"filterType": "DateTime:Range", "details": map[string]interface{}{}}
}

func FilterSingleChoiceYesNo() map[string]interface{} {
	return map[string]interface{}{
		"filterType": "SingleChoice",
		"details": map[string]interface{}{
			"PossibleValues": []map[string]interface{}{
				{"value": "Yes", "label": "Yes"},
				{"value": "No", "label": "No"},
			},
		},
	}
}

func FilterSingleChoiceTrueFalse() map[string]interface{} {
	return map[string]interface{}{
		"filterType": "SingleChoice",
		"details": map[string]interface{}{
			"PossibleValues": []map[string]interface{}{
				{"value": "True", "label": "True"},
				{"value": "False", "label": "False"},
			},
		},
	}
}

func FilterMultipleChoiceDirection() map[string]interface{} {
	return map[string]interface{}{
		"filterType": "MultipleChoice",
		"details": map[string]interface{}{
			"PossibleValues": []map[string]interface{}{
				{"value": "Inbound", "label": "Inbound"},
				{"value": "Outbound", "label": "Outbound"},
			},
		},
	}
}

func FilterMultipleChoiceAssetClass() map[string]interface{} {
	return map[string]interface{}{
		"filterType": "MultipleChoice",
		"details": map[string]interface{}{
			"PossibleValues": []map[string]interface{}{
				{"value": "Currency", "label": "Currency"},
				{"value": "Cryptocurrency", "label": "Cryptocurrency"},
			},
		},
	}
}

func FilterMultipleChoiceStatus() map[string]interface{} {
	return map[string]interface{}{
		"filterType": "MultipleChoice",
		"details": map[string]interface{}{
			"PossibleValues": []map[string]interface{}{
				{"value": "Complete", "label": "Complete"},
				{"value": "Pending", "label": "Pending"},
				{"value": "Reserved", "label": "Reserved"},
				{"value": "Frozen", "label": "Frozen"},
			},
		},
	}
}

// =============================================================================
// SHARED COLUMN DEFINITIONS (reusable across list types)
// =============================================================================

var (
	ColTransactionId = ColumnDef{
		ColumnKey: "CustomerAssetAccountsTransactions__Id", LabelKey: "Id", LabelValue: "Id",
		BSortable: true, BFilterable: true, BVisible: false, Type: "Integer",
		FilterMetadata: FilterAmount(),
	}

	ColTransactionDate = ColumnDef{
		ColumnKey: "CustomerAssetAccountsTransactions__Date", LabelKey: "Ticker", LabelValue: "Date",
		BSortable: true, BFilterable: true, BVisible: true, Type: "DateTime",
		FilterMetadata: FilterDateTimeRange(),
	}

	ColDescription = ColumnDef{
		ColumnKey: "CustomerAssetAccountsTransactions__Description", LabelKey: "Description", LabelValue: "Description",
		BSortable: true, BFilterable: false, BVisible: true, Type: "String",
	}

	ColAmount = ColumnDef{
		ColumnKey: "CustomerAssetAccountsTransactions__Amount", LabelKey: "Amount", LabelValue: "Amount",
		BSortable: true, BFilterable: false, BVisible: true, Type: "Decimal",
	}

	ColAmountFilterable = ColumnDef{
		ColumnKey: "CustomerAssetAccountsTransactions__Amount", LabelKey: "Amount", LabelValue: "Amount",
		BSortable: true, BFilterable: true, BVisible: true, Type: "Decimal",
		FilterMetadata: FilterAmountWithDetails(),
	}

	ColFee = ColumnDef{
		ColumnKey: "TransactionAmounts__Fee", LabelKey: "Fee", LabelValue: "Fee",
		BSortable: false, BFilterable: false, BVisible: true, Type: "Decimal",
	}

	ColBalance = ColumnDef{
		ColumnKey: "TransactionAmounts__Balance", LabelKey: "Balance", LabelValue: "Balance",
		BSortable: false, BFilterable: false, BVisible: true, Type: "Decimal",
	}

	ColInfoRequest = ColumnDef{
		ColumnKey: "TransactionAmounts__InfoRequest", LabelKey: "InfoRequest", LabelValue: "Info Request",
		BSortable: true, BFilterable: true, BVisible: true, Type: "String",
		FilterMetadata: FilterTextContains(),
	}

	ColAlertsAvailable = ColumnDef{
		ColumnKey: "CustomerAssetAccountsTransactionScreenings__bAlertsAvailable", LabelKey: "bAlertsAvailable", LabelValue: "Alerts Available",
		BSortable: false, BFilterable: false, BVisible: false, Type: "Boolean",
	}

	ColInsufficientFunds = ColumnDef{
		ColumnKey: "CustomerAssetAccountsTransactions__bInsufficientOpAccountBalance", LabelKey: "bInsufficientOpAccountBalance", LabelValue: "Insufficient Funds",
		BSortable: false, BFilterable: false, BVisible: false, Type: "Boolean",
	}

	ColAssetCode = ColumnDef{
		ColumnKey: "Assets__Code", LabelKey: "Asset", LabelValue: "Asset",
		BSortable: true, BFilterable: true, BVisible: true, Type: "String",
		FilterMetadata: FilterTextContains(),
	}

	ColCustomersId = ColumnDef{
		ColumnKey: "Customers__Id", LabelKey: "CustomersID", LabelValue: "Cus. ID",
		BSortable: false, BFilterable: false, BVisible: false, Type: "Integer",
	}

	ColCustomersCode = ColumnDef{
		ColumnKey: "Customers__CustomersCode", LabelKey: "CustomersCode", LabelValue: "Cst. Code",
		BSortable: true, BFilterable: true, BVisible: true, Type: "String",
		FilterMetadata: FilterTextContains(),
	}

	ColCustomerName = ColumnDef{
		ColumnKey: "CustomerUsers__FullName", LabelKey: "CustomersName", LabelValue: "Cst. Name",
		BSortable: true, BFilterable: true, BVisible: true, Type: "String",
		FilterMetadata: FilterTextContains(),
	}

	ColCompanyName = ColumnDef{
		ColumnKey: "Customers__CompanyName", LabelKey: "BusinessName", LabelValue: "Business",
		BSortable: true, BFilterable: true, BVisible: true, Type: "String",
		FilterMetadata: FilterTextContains(),
	}

	ColBrand = ColumnDef{
		ColumnKey: "LicenseesBrands__StatementDescriptor", LabelKey: "LicenseesBrand", LabelValue: "Brand",
		BSortable: true, BFilterable: true, BVisible: true, Type: "String",
		FilterMetadata: FilterTextContains(),
	}

	ColWalletAddress = ColumnDef{
		ColumnKey: "ExternalAssetWallets__WalletAddress", LabelKey: "WalletAddress", LabelValue: "Wallet Address",
		BSortable: false, BFilterable: false, BVisible: true, Type: "String",
	}

	ColTransactionHash = ColumnDef{
		ColumnKey: "CustomerAssetAccountsTransactions__TransactionHash", LabelKey: "TransactionHash", LabelValue: "Transaction Hash",
		BSortable: false, BFilterable: false, BVisible: true, Type: "String",
	}

	ColWidgetCode = ColumnDef{
		ColumnKey: "WidgetClientUsers__WidgetClientUsersCode", LabelKey: "WidgetClientUsersCode", LabelValue: "Widget Code",
		BSortable: true, BFilterable: true, BVisible: true, Type: "String",
		FilterMetadata: FilterTextContains(),
	}

	ColTransactionType = ColumnDef{
		ColumnKey: "FeeIncurringActions__DisplayName", LabelKey: "TransactionType", LabelValue: "Type",
		BSortable: true, BFilterable: true, BVisible: true, Type: "String",
		FilterMetadata: FilterTextContains(),
	}

	ColExternalId = ColumnDef{
		ColumnKey: "CustomerAssetAccountsTransactions__ExternalId", LabelKey: "ExternalID", LabelValue: "External ID",
		BSortable: false, BFilterable: false, BVisible: false, Type: "String",
	}

	ColMaxDecimalPrecision = ColumnDef{
		ColumnKey: "Assets__MaxDecimalPrecision", LabelKey: "MaxDecimalPrecision", LabelValue: "MaxDecimalPrecision",
		BSortable: false, BFilterable: false, BVisible: false, Type: "Integer",
	}

	ColProductName = ColumnDef{
		ColumnKey: "Products__ProductName", LabelKey: "ProductName", LabelValue: "Product",
		BSortable: false, BFilterable: true, BVisible: true, Type: "String",
		FilterMetadata: FilterTextContains(),
	}

	ColCancelAvailable = ColumnDef{
		ColumnKey: "bCancelAvailable", LabelKey: "bCancelAvailable", LabelValue: "Cancel Available",
		BSortable: false, BFilterable: false, BVisible: false, Type: "Boolean",
	}

	ColMarkAsPendingAvailable = ColumnDef{
		ColumnKey: "bMarkAsPendingAvailable", LabelKey: "bMarkAsPendingAvailable", LabelValue: "Mark As Pending Available",
		BSortable: false, BFilterable: false, BVisible: false, Type: "Boolean",
	}

	ColCompleteAvailable = ColumnDef{
		ColumnKey: "bCompleteAvailable", LabelKey: "bCompleteAvailable", LabelValue: "Complete Available",
		BSortable: false, BFilterable: false, BVisible: false, Type: "Boolean",
	}

	ColAssignAvailable = ColumnDef{
		ColumnKey: "bAssignAvailable", LabelKey: "bAssignAvailable", LabelValue: "Assign Available",
		BSortable: false, BFilterable: false, BVisible: false, Type: "Boolean",
	}

	ColReference = ColumnDef{
		ColumnKey: "CustomerAssetAccountsTransactions__Refference", LabelKey: "Refference", LabelValue: "Refference",
		BSortable: true, BFilterable: false, BVisible: true, Type: "String",
	}
)

// =============================================================================
// LIST-SPECIFIC COLUMN BUILDERS
// =============================================================================

// ColumnsTransactionsListRefactored returns columns for the basic transactions list
func ColumnsTransactionsListRefactored() []map[string]interface{} {
	cols := []ColumnDef{
		ColTransactionId,
		ColTransactionDate,
		ColDescription,
		ColAmount,
		ColFee,
		ColBalance,
		ColInfoRequest,
		ColAlertsAvailable,
		ColInsufficientFunds,
		ColAssetCode,
		ColCustomersId,
		ColCustomersCode,
		ColCustomerName,
		ColCompanyName,
		ColBrand,
		ColWalletAddress,
		ColTransactionHash,
		ColWidgetCode,
		ColTransactionType,
		ColExternalId,
		ColMaxDecimalPrecision,
		ColProductName,
		ColCancelAvailable,
		ColMarkAsPendingAvailable,
		ColCompleteAvailable,
		ColAssignAvailable,
	}
	return ColumnsToMaps(cols)
}

// ColumnsPendingTransactionsListRefactored returns columns for pending transactions
// (identical to ColumnsTransactionsList)
func ColumnsPendingTransactionsListRefactored() []map[string]interface{} {
	return ColumnsTransactionsListRefactored()
}

// ColumnsPendingTransactionsTreasuryListRefactored returns columns for treasury pending transactions
// (same as pending but with Reference column inserted before Description)
func ColumnsPendingTransactionsTreasuryListRefactored() []map[string]interface{} {
	cols := []ColumnDef{
		ColTransactionId,
		ColTransactionDate,
		ColReference, // Extra column for treasury
		ColDescription,
		ColAmount,
		ColFee,
		ColBalance,
		ColInfoRequest,
		ColAlertsAvailable,
		ColInsufficientFunds,
		ColAssetCode,
		ColCustomersId,
		ColCustomersCode,
		ColCustomerName,
		ColCompanyName,
		ColBrand,
		ColWalletAddress,
		ColTransactionHash,
		ColWidgetCode,
		ColTransactionType,
		ColExternalId,
		ColMaxDecimalPrecision,
		ColProductName,
		ColCancelAvailable,
		ColMarkAsPendingAvailable,
		ColCompleteAvailable,
		ColAssignAvailable,
	}
	return ColumnsToMaps(cols)
}

// ColumnsTransactionsListAllRefactored returns columns for the "all transactions" list
func ColumnsTransactionsListAllRefactored() []map[string]interface{} {
	// This list has different columns and more filter options
	cols := []ColumnDef{
		ColTransactionId,
		{
			ColumnKey: "CustomerAssetAccountsTransactions__CustomerAssetAccountsTransactionsCode", LabelKey: "CustomerAssetAccountsTransactionsCode", LabelValue: "Code",
			BSortable: true, BFilterable: true, BVisible: true, Type: "String",
			FilterMetadata: FilterTextContains(),
		},
		ColDescription,
		ColAmountFilterable,
		{
			ColumnKey: "CustomerAssetAccountsTransactions__Date", LabelKey: "Ticker", LabelValue: "Date",
			BSortable: true, BFilterable: true, BVisible: true, Type: "DateTime",
			FilterMetadata: FilterDateTimeRange(),
		},
		{
			ColumnKey: "CustomerAssetAccountsTransactions__BankReference", LabelKey: "BankReference", LabelValue: "Bank ID",
			BSortable: true, BFilterable: true, BVisible: true, Type: "String",
			FilterMetadata: FilterTextContainsWithDetails(),
		},
		{
			ColumnKey: "CustomerAssetAccountsTransactions__ImadNumber", LabelKey: "ImadNumber", LabelValue: "IMAD",
			BSortable: true, BFilterable: true, BVisible: true, Type: "String",
			FilterMetadata: FilterTextContainsWithDetails(),
		},
		{
			ColumnKey: "CustomerAssetAccounts__Id", LabelKey: "CustomerAssetAccountsID", LabelValue: "Customer Asset Accounts ID",
			BSortable: true, BFilterable: false, BVisible: false, Type: "Integer",
		},
		ColCustomersId,
		{
			ColumnKey: "Customers__CustomersCode", LabelKey: "CustomersCode", LabelValue: "Cst. Code",
			BSortable: true, BFilterable: true, BVisible: true, Type: "String",
			FilterMetadata: FilterTextContainsWithDetails(),
		},
		{
			ColumnKey: "CustomerDetails__CustomerName", LabelKey: "CustomerName", LabelValue: "Cst. Name",
			BSortable: true, BFilterable: true, BVisible: true, Type: "String",
			FilterMetadata: FilterTextContainsWithDetails(),
		},
		{
			ColumnKey: "Licensees__LicenseeName", LabelKey: "LicenseeName", LabelValue: "Licensee",
			BSortable: true, BFilterable: true, BVisible: true, Type: "String",
			FilterMetadata: FilterTextContainsWithDetails(),
		},
		{
			ColumnKey: "LicenseesBrands__InternalName", LabelKey: "LicenseesBrand", LabelValue: "Brand",
			BSortable: true, BFilterable: true, BVisible: true, Type: "String",
			FilterMetadata: FilterTextContains(),
		},
		{
			ColumnKey: "Assets__Code", LabelKey: "Asset", LabelValue: "Asset",
			BSortable: true, BFilterable: true, BVisible: true, Type: "String",
			FilterMetadata: FilterTextContainsWithDetails(),
		},
		ColMaxDecimalPrecision,
		{
			ColumnKey: "AssetClasses__Name", LabelKey: "AssetClass", LabelValue: "Asset Type",
			BSortable: false, BFilterable: true, BVisible: false, Type: "String",
			FilterMetadata: FilterMultipleChoiceAssetClass(),
		},
		{
			ColumnKey: "Products__DisplayName", LabelKey: "ProductName", LabelValue: "Product",
			BSortable: true, BFilterable: true, BVisible: true, Type: "String",
			FilterMetadata: FilterTextContainsWithDetails(),
		},
		{
			ColumnKey: "Transaction__Rail", LabelKey: "TransactionRail", LabelValue: "Transaction Rail",
			BSortable: true, BFilterable: true, BVisible: true, Type: "String",
			FilterMetadata: FilterTextContainsWithDetails(),
		},
		{
			ColumnKey: "TransactionDirection__Direction", LabelKey: "Direction", LabelValue: "Direction",
			BSortable: false, BFilterable: true, BVisible: false, Type: "String",
			FilterMetadata: FilterMultipleChoiceDirection(),
		},
		{
			ColumnKey: "TransactionStatus__Status", LabelKey: "Status", LabelValue: "Status",
			BSortable: true, BFilterable: true, BVisible: false, Type: "String",
			FilterMetadata: FilterMultipleChoiceStatus(),
		},
		{
			ColumnKey: "Approvals__Status", LabelKey: "CustomerApprovalStatus", LabelValue: "CA Status",
			BSortable: false, BFilterable: false, BVisible: true, Type: "String",
		},
		{
			ColumnKey: "TransactionReversed__Reversed", LabelKey: "Reversed", LabelValue: "Reversed",
			BSortable: true, BFilterable: true, BVisible: false, Type: "String",
			FilterMetadata: FilterSingleChoiceYesNo(),
		},
		{
			ColumnKey: "TransactionScreening__Rejected", LabelKey: "Rejected", LabelValue: "Rejected",
			BSortable: true, BFilterable: true, BVisible: false, Type: "String",
			FilterMetadata: FilterSingleChoiceYesNo(),
		},
		{
			ColumnKey: "CustomerAssetAccountsTransactions__bFeeWaivable", LabelKey: "bFeeWaivable", LabelValue: "Fee Waivable",
			BSortable: false, BFilterable: false, BVisible: false, Type: "Boolean",
		},
		{
			ColumnKey: "InformationRequestStatus__InformationRequestStatus", LabelKey: "InfoRequest", LabelValue: "RFI Status",
			BSortable: true, BFilterable: true, BVisible: true, Type: "String",
			FilterMetadata: FilterTextContainsWithDetails(),
		},
		{
			ColumnKey: "TransactionScreening__OriginatorName", LabelKey: "OriginatorName", LabelValue: "Originator Name",
			BSortable: true, BFilterable: true, BVisible: false, Type: "String",
			FilterMetadata: FilterTextContainsWithDetails(),
		},
		{
			ColumnKey: "TransactionScreening__BeneficiaryName", LabelKey: "BeneficiaryName", LabelValue: "Beneficiary Name",
			BSortable: true, BFilterable: true, BVisible: false, Type: "String",
			FilterMetadata: FilterTextContainsWithDetails(),
		},
		{
			ColumnKey: "TransactionScreening__TransmitterName", LabelKey: "TransmitterName", LabelValue: "Transmitter Name",
			BSortable: true, BFilterable: true, BVisible: false, Type: "String",
			FilterMetadata: FilterTextContainsWithDetails(),
		},
		{
			ColumnKey: "TransactionScreening__SenderJurisdiction", LabelKey: "SenderJurisdiction", LabelValue: "Sender Jurisdiction",
			BSortable: true, BFilterable: true, BVisible: false, Type: "String",
			FilterMetadata: FilterTextContainsWithDetails(),
		},
		{
			ColumnKey: "TransactionScreening__TransmitterJurisdiction", LabelKey: "TransmitterJurisdiction", LabelValue: "Transmitter Jurisdiction",
			BSortable: true, BFilterable: true, BVisible: false, Type: "String",
			FilterMetadata: FilterTextContainsWithDetails(),
		},
		{
			ColumnKey: "TransactionScreening__BeneficiaryJurisdiction", LabelKey: "BeneficiaryJurisdiction", LabelValue: "Beneficiary Jurisdiction",
			BSortable: true, BFilterable: true, BVisible: false, Type: "String",
			FilterMetadata: FilterTextContainsWithDetails(),
		},
		{
			ColumnKey: "TransactionScreening__PaymentType", LabelKey: "PaymentType", LabelValue: "Payment Type",
			BSortable: true, BFilterable: true, BVisible: false, Type: "String",
			FilterMetadata: FilterTextContainsWithDetails(),
		},
		{
			ColumnKey: "TransactionScreening__International", LabelKey: "International", LabelValue: "International",
			BSortable: true, BFilterable: true, BVisible: false, Type: "String",
			FilterMetadata: FilterSingleChoiceYesNo(),
		},
		{
			ColumnKey: "TransactionAmounts__Balance", LabelKey: "Balance", LabelValue: "Balance",
			BSortable: false, BFilterable: false, BVisible: true, Type: "Decimal",
		},
		ColAlertsAvailable,
		{
			ColumnKey: "bCancellable", LabelKey: "bCancellable", LabelValue: "Cancellable",
			BSortable: true, BFilterable: true, BVisible: false, Type: "Boolean",
			FilterMetadata: FilterSingleChoiceTrueFalse(),
		},
	}
	return ColumnsToMaps(cols)
}

// ColumnsFrozenTransactionsListRefactored returns columns for frozen transactions
func ColumnsFrozenTransactionsListRefactored() []map[string]interface{} {
	cols := []ColumnDef{
		{
			ColumnKey: "Assets__Name", LabelKey: "name", LabelValue: "Name",
			BSortable: true, BFilterable: true, BVisible: true, Type: "String",
			FilterMetadata: FilterTextContains(),
		},
		{
			ColumnKey: "Assets__Code", LabelKey: "ticker", LabelValue: "Ticker",
			BSortable: true, BFilterable: true, BVisible: true, Type: "String",
			FilterMetadata: FilterTextContains(),
		},
		{
			ColumnKey: "CustomerAssetAccountsTransactions__CustomerAssetAccountsTransactionsCode", LabelKey: "reference", LabelValue: "Reference",
			BSortable: true, BFilterable: true, BVisible: true, Type: "String",
			FilterMetadata: FilterTextContains(),
		},
		{
			ColumnKey: "CustomerAssetAccountsTransactions__Date", LabelKey: "date", LabelValue: "Date",
			BSortable: true, BFilterable: true, BVisible: true, Type: "DateTime",
			FilterMetadata: FilterDateTimeRange(),
		},
		{
			ColumnKey: "TransactionTypes__Type", LabelKey: "ordertype", LabelValue: "Order Type",
			BSortable: true, BFilterable: true, BVisible: true, Type: "String",
			FilterMetadata: FilterTextContains(),
		},
		{
			ColumnKey: "CustomerAssetAccountsTransactions__Amount", LabelKey: "amount", LabelValue: "Amount",
			BSortable: true, BFilterable: true, BVisible: true, Type: "Decimal",
			FilterMetadata: FilterAmount(),
		},
		{
			ColumnKey: "FeeTransactions__Amount", LabelKey: "fee", LabelValue: "Fee",
			BSortable: true, BFilterable: true, BVisible: true, Type: "Decimal",
			FilterMetadata: FilterAmount(),
		},
		{
			ColumnKey: "TRMLabsHelper__bSupportedCurrency", LabelKey: "SupportedCurrency", LabelValue: "Supported Currency",
			BSortable: true, BFilterable: true, BVisible: true, Type: "Boolean",
			FilterMetadata: map[string]interface{}{
				"filterType": "SingleChoice",
				"details": map[string]interface{}{
					"PossibleValues": []map[string]string{
						{"value": "True", "label": "True"},
						{"value": "False", "label": "False"},
					},
				},
			},
		},
		{
			ColumnKey: "CustomerAssetAccountsTransactions__bAwaitingUnfreeze", LabelKey: "Awaiting Unfreeze", LabelValue: "Awaiting Unfreeze",
			BSortable: true, BFilterable: true, BVisible: true, Type: "Boolean",
			FilterMetadata: map[string]interface{}{
				"filterType": "SingleChoice",
				"details": map[string]interface{}{
					"PossibleValues": []map[string]string{
						{"value": "True", "label": "True"},
						{"value": "False", "label": "False"},
					},
				},
			},
		},
	}
	return ColumnsToMaps(cols)
}
