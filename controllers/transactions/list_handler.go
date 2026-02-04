package transactions

import (
	"net/http"
	"strconv"

	"cloud-web-phoenix-customer-v1-go/controllers/admin"
	"cloud-web-phoenix-customer-v1-go/controllers/auth"
	"cloud-web-phoenix-customer-v1-go/db"

	"github.com/gin-gonic/gin"
)

// =============================================================================
// GENERIC LIST HANDLER - Eliminates ~80% of duplication in index.go
// =============================================================================

// ListConfig defines the configuration for a list endpoint
type ListConfig struct {
	// ListKey is the key used for list selections (e.g., "Admin_CustomerAssetAccounts")
	ListKey string

	// SPName is the stored procedure name to execute
	SPName string

	// ColumnMap maps frontend column keys to SP column names for filtering/sorting
	ColumnMap map[string]string

	// SearchFields lists the columns that can be searched
	SearchFields []string

	// ColumnsFunc returns the column definitions for the list
	ColumnsFunc func() []map[string]interface{}

	// RequiredParams maps query param names to SP param names (validation required)
	RequiredParams map[string]string

	// OptionalParams maps query param names to SP param names (no validation)
	OptionalParams map[string]string

	// ExtraParams are static params to add to every SP call
	ExtraParams map[string]interface{}

	// FormatMoney indicates whether to format money fields (default true)
	FormatMoney bool

	// MaxPageSize is the maximum allowed page size (default 200)
	MaxPageSize int

	// CustomResponseBuilder allows custom response building (optional)
	CustomResponseBuilder func(c *gin.Context, listData []map[string]interface{}, total int, q admin.QueryRecordList, siteUsersId int)
}

// HandleList is a generic handler for list endpoints
// It handles: auth, param validation, query parsing, list selections, SP execution, and response building
func HandleList(c *gin.Context, cfg ListConfig) {
	// 1. Auth check
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	siteUsersId := user["id"].(int)

	// 2. Validate and collect required params
	spExtraParams := make(map[string]interface{})

	for queryParam, spParam := range cfg.RequiredParams {
		valStr := c.Query(queryParam)
		if valStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "0",
				"errors": []gin.H{
					{"fieldName": queryParam, "messageCode": "Required"},
				},
			})
			return
		}

		val, err := strconv.Atoi(valStr)
		if err != nil || val <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "0",
				"errors": []gin.H{
					{"fieldName": queryParam, "messageCode": "Invalid"},
				},
			})
			return
		}
		spExtraParams[spParam] = val
	}

	// 3. Collect optional params
	for queryParam, spParam := range cfg.OptionalParams {
		valStr := c.Query(queryParam)
		if valStr != "" {
			if val, err := strconv.Atoi(valStr); err == nil {
				spExtraParams[spParam] = val
			} else {
				// If not a number, pass as string
				spExtraParams[spParam] = valStr
			}
		}
	}

	// 4. Add static extra params
	for k, v := range cfg.ExtraParams {
		spExtraParams[k] = v
	}

	// 5. Parse query params
	q := admin.ParseQueryRecordList(c.Request.URL.Query())

	// 6. Apply page size limit
	maxPageSize := cfg.MaxPageSize
	if maxPageSize == 0 {
		maxPageSize = 200
	}
	if q.PageSize > maxPageSize {
		q.PageSize = maxPageSize
	}

	// 7. Load and apply list selections
	if cfg.ListKey != "" {
		ex := loadListSelections(siteUsersId, cfg.ListKey)
		admin.OverrideWithSelections(&q, ex)
	}

	// 8. Build SP config
	spCfg := admin.ListSPConfig{
		ListKey:      cfg.ListKey,
		TrackingID:   "DefaultTrackingID",
		ColumnMap:    cfg.ColumnMap,
		SearchFields: cfg.SearchFields,
	}

	// 9. Build SP params
	spParams := admin.BuildListSPParams(q, siteUsersId, spCfg)

	// 10. Add extra params
	for k, v := range spExtraParams {
		spParams[k] = v
	}

	// 11. Execute SP
	res, err := auth.ExecSP(db.DB, cfg.SPName, spParams, 2)

	// 12. Build response function
	respond := func(listData []map[string]interface{}, total int) {
		if cfg.CustomResponseBuilder != nil {
			cfg.CustomResponseBuilder(c, listData, total, q, siteUsersId)
			return
		}

		columns := cfg.ColumnsFunc()
		details := admin.BuildListDetails(columns, listData, q, total)

		c.JSON(http.StatusOK, gin.H{
			"status":  "1",
			"id":      siteUsersId,
			"errors":  []interface{}{},
			"details": details,
		})
	}

	// 13. Handle errors
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			respond([]map[string]interface{}{}, 0)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "0",
			"error":  err.Error(),
		})
		return
	}

	// 14. Parse rows
	rows, ok := res.([]map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "0",
			"error":  "Invalid SP response",
		})
		return
	}

	if len(rows) == 0 {
		respond([]map[string]interface{}{}, 0)
		return
	}

	// 15. Process rows
	var listData []map[string]interface{}
	var total int

	formatMoney := cfg.FormatMoney
	if cfg.ColumnsFunc != nil && !formatMoney {
		// Default to true if not explicitly set
		formatMoney = true
	}

	if formatMoney {
		listData, total = ProcessListRowsAdmin(rows)
	} else {
		listData, total = ProcessListRowsAdminNoMoney(rows)
	}

	respond(listData, total)
}

// loadListSelections loads user's saved list selections
func loadListSelections(siteUsersId int, listKey string) *admin.ListSelections {
	sp := "v1_General_ListSelectionsModule_GetSiteUsersListSelections"
	params := map[string]interface{}{
		"SiteUsersId": siteUsersId,
		"TrackingId":  "DefaultTrackingID",
		"ListKey":     listKey,
	}

	res, err := auth.ExecSP(db.DB, sp, params, 1)
	if err != nil {
		return nil
	}
	row, ok := auth.AsSingleRow(res)
	if !ok {
		return nil
	}
	return admin.SelectionsFromRow(row)
}

// =============================================================================
// PRE-CONFIGURED LIST CONFIGS
// These can be used directly with HandleList
// =============================================================================

// ListConfigTransactions returns config for the basic transactions list
func ListConfigTransactions() ListConfig {
	return ListConfig{
		ListKey: "Admin_CustomerAssetAccounts",
		SPName:  "v1_AdminRole_TransactionsModule_GetTransactionList",
		ColumnMap: map[string]string{
			"customerAssetAccountsTransactions__Id":                                    "CustomerAssetAccountsTransactions.Id",
			"customerAssetAccountsTransactions__Date":                                  "CustomerAssetAccountsTransactions.Date",
			"customerAssetAccountsTransactions__Description":                           "CustomerAssetAccountsTransactions.Description",
			"transactionAmounts__MoneyIn":                                              "TransactionAmounts.MoneyIn",
			"transactionAmounts__MoneyOut":                                             "TransactionAmounts.MoneyOut",
			"transactionAmounts__Balance":                                              "TransactionAmounts.Balance",
			"customerAssetAccountsTransactions__CustomerAssetAccountsTransactionsCode": "CustomerAssetAccountsTransactions.CustomerAssetAccountsTransactionsCode",
			"transactionAmounts__Status":                                               "TransactionAmounts.Status",
			"approvals__Status":                                                        "Approvals.Status",
		},
		SearchFields: []string{
			"CustomerAssetAccountsTransactions.Description",
			"CustomerAssetAccountsTransactions.CustomerAssetAccountsTransactionsCode",
			"TransactionAmounts.Status",
			"Approvals.Status",
		},
		ColumnsFunc: ColumnsTransactionsListRefactored,
		RequiredParams: map[string]string{
			"customerAssetAccountsId": "CustomerAssetAccountsID",
		},
		FormatMoney: true,
	}
}

// ListConfigTransactionsAll returns config for the "all transactions" list
func ListConfigTransactionsAll() ListConfig {
	return ListConfig{
		ListKey:      "Admin_GetAllTransactions",
		SPName:       "v1_AdminRole_TransactionsModule_GetAllTransactionsList",
		ColumnMap:    map[string]string{},
		SearchFields: []string{},
		ColumnsFunc:  ColumnsTransactionsListAllRefactored,
		FormatMoney:  true,
	}
}

// ListConfigPendingTransactions returns config for pending transactions list
func ListConfigPendingTransactions(bFilterToOwnAssignments bool) ListConfig {
	return ListConfig{
		ListKey: "Admin_GetPendingTransactions",
		SPName:  "v1_AdminRole_PendingTransactionsModule_GetTransactionList",
		ColumnMap: map[string]string{
			"customerAssetAccountsTransactions__Id":          "CustomerAssetAccountsTransactions.Id",
			"customerAssetAccountsTransactions__Date":        "CustomerAssetAccountsTransactions.Date",
			"customerAssetAccountsTransactions__Description": "CustomerAssetAccountsTransactions.Description",
			"customerAssetAccountsTransactions__Amount":      "CustomerAssetAccountsTransactions.Amount",
			"transactionAmounts__Fee":                        "TransactionAmounts.Fee",
			"transactionAmounts__Balance":                    "TransactionAmounts.Balance",
			"transactionAmounts__InfoRequest":                "TransactionAmounts.InfoRequest",
			"assets__Code":                                   "Assets.Code",
			"customers__CustomersCode":                       "Customers.CustomersCode",
			"customerUsers__FullName":                        "CustomerUsers.FullName",
			"customers__CompanyName":                         "Customers.CompanyName",
			"licenseesBrands__StatementDescriptor":           "LicenseesBrands.StatementDescriptor",
			"widgetClientUsers__WidgetClientUsersCode":       "WidgetClientUsers.WidgetClientUsersCode",
			"feeIncurringActions__DisplayName":               "FeeIncurringActions.DisplayName",
			"products__ProductName":                          "Products.ProductName",
		},
		SearchFields: []string{
			"CustomerAssetAccountsTransactions.Description",
			"Assets.Code",
			"Customers.CustomersCode",
			"CustomerUsers.FullName",
			"Customers.CompanyName",
			"LicenseesBrands.StatementDescriptor",
			"WidgetClientUsers.WidgetClientUsersCode",
			"FeeIncurringActions.DisplayName",
		},
		ColumnsFunc: ColumnsPendingTransactionsListRefactored,
		ExtraParams: map[string]interface{}{
			"bFilterToOwnAssignments": bFilterToOwnAssignments,
		},
		FormatMoney: true,
	}
}

// ListConfigPendingTransactionsTreasury returns config for treasury pending transactions
func ListConfigPendingTransactionsTreasury() ListConfig {
	return ListConfig{
		ListKey: "Admin_GetPendingTransactionsTreasury",
		SPName:  "v1_AdminRole_PendingTransactionsTreasuryModule_GetTransactionList",
		ColumnMap: map[string]string{
			"customerAssetAccountsTransactions__Id":          "CustomerAssetAccountsTransactions.Id",
			"customerAssetAccountsTransactions__Date":        "CustomerAssetAccountsTransactions.Date",
			"customerAssetAccountsTransactions__Description": "CustomerAssetAccountsTransactions.Description",
			"customerAssetAccountsTransactions__Amount":      "CustomerAssetAccountsTransactions.Amount",
			"transactionAmounts__Fee":                        "TransactionAmounts.Fee",
			"transactionAmounts__Balance":                    "TransactionAmounts.Balance",
			"transactionAmounts__InfoRequest":                "TransactionAmounts.InfoRequest",
			"assets__Code":                                   "Assets.Code",
			"customers__CustomersCode":                       "Customers.CustomersCode",
			"customerUsers__FullName":                        "CustomerUsers.FullName",
			"customers__CompanyName":                         "Customers.CompanyName",
			"licenseesBrands__StatementDescriptor":           "LicenseesBrands.StatementDescriptor",
			"widgetClientUsers__WidgetClientUsersCode":       "WidgetClientUsers.WidgetClientUsersCode",
			"feeIncurringActions__DisplayName":               "FeeIncurringActions.DisplayName",
			"products__ProductName":                          "Products.ProductName",
		},
		SearchFields: []string{
			"CustomerAssetAccountsTransactions.Description",
			"Assets.Code",
			"Customers.CustomersCode",
			"CustomerUsers.FullName",
			"Customers.CompanyName",
			"LicenseesBrands.StatementDescriptor",
			"WidgetClientUsers.WidgetClientUsersCode",
			"FeeIncurringActions.DisplayName",
		},
		ColumnsFunc: ColumnsPendingTransactionsTreasuryListRefactored,
		FormatMoney: true,
	}
}

// ListConfigFrozenTransactions returns config for frozen transactions
func ListConfigFrozenTransactions() ListConfig {
	return ListConfig{
		ListKey:      "Admin_GetFrozenTransactions",
		SPName:       "v1_AdminRole_FrozenTransactionsModule_GetList",
		ColumnMap:    map[string]string{},
		SearchFields: []string{},
		ColumnsFunc:  ColumnsFrozenTransactionsListRefactored,
		FormatMoney:  false, // Frozen list doesn't format money
	}
}
