package transactions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"cloud-web-phoenix-customer-v1-go/controllers/admin"
	"cloud-web-phoenix-customer-v1-go/controllers/auth"
	"cloud-web-phoenix-customer-v1-go/db"

	"github.com/gin-gonic/gin"
)

func List(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	// ✅ required query param in .NET: customerAssetAccountsId
	customerAssetAccountsIdStr := c.Query("customerAssetAccountsId")
	if customerAssetAccountsIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"errors": []gin.H{
				{"fieldName": "customerAssetAccountsId", "messageCode": "Required"},
			},
		})
		return
	}

	customerAssetAccountsId, err := strconv.Atoi(customerAssetAccountsIdStr)
	if err != nil || customerAssetAccountsId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"errors": []gin.H{
				{"fieldName": "customerAssetAccountsId", "messageCode": "Invalid"},
			},
		})
		return
	}

	// .NET: QueryRecordListDto bind
	q := admin.ParseQueryRecordList(c.Request.URL.Query())

	// ✅ .NET: PageSize max 200
	if q.PageSize > 200 {
		q.PageSize = 200
	}

	// .NET list selections key for this endpoint:
	// "Admin_CustomerAssetAccounts"
	loadSelections := func() *admin.ListSelections {
		sp := "v1_General_ListSelectionsModule_GetSiteUsersListSelections"
		params := map[string]interface{}{
			"SiteUsersId": siteUsersId,
			"TrackingId":  "DefaultTrackingID",
			"ListKey":     "Admin_CustomerAssetAccounts",
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

	ex := loadSelections()
	admin.OverrideWithSelections(&q, ex)

	cfg := admin.ListSPConfig{
		ListKey:    "Admin_CustomerAssetAccounts",
		TrackingID: "DefaultTrackingID",
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
	}

	// Build SP params (pagination + filters + sort + search)
	spParams := admin.BuildListSPParams(q, siteUsersId, cfg)

	// ✅ additional param required by this list:
	spParams["CustomerAssetAccountsID"] = customerAssetAccountsId

	spName := "v1_AdminRole_TransactionsModule_GetTransactionList"
	res, err := auth.ExecSP(db.DB, spName, spParams, 2)

	respond := func(listData []map[string]interface{}, total int) {
		// ✅ Path B: static columns EXACT like .NET
		columns := ColumnsTransactionsList()

		details := admin.BuildListDetails(columns, listData, q, total)
		c.JSON(http.StatusOK, gin.H{
			"status":  "1",
			"id":      siteUsersId,
			"errors":  []interface{}{},
			"details": details,
		})
	}

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

	// Normalize row keys to match .NET listData keys (lowerCamel)
	listData := make([]map[string]interface{}, 0, len(rows))
	total := 0

	for _, row := range rows {
		item := admin.NormalizeRowKeys(row)
		FormatMoneyFields2dp(item)
		listData = append(listData, item)

		if v, ok := row["HowManyResults"]; ok && v != nil {
			switch t := v.(type) {
			case int:
				total = t
			case int64:
				total = int(t)
			case float64:
				total = int(t)
			default:
				i, _ := strconv.Atoi(fmt.Sprint(t))
				if i > 0 {
					total = i
				}
			}
		}
	}

	respond(listData, total)
}

func ListAll(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	q := admin.ParseQueryRecordList(c.Request.URL.Query())
	if q.PageSize > 200 {
		q.PageSize = 200
	}

	// .NET list selections key for listall:
	// "Admin_GetAllTransactions"
	loadSelections := func() *admin.ListSelections {
		sp := "v1_General_ListSelectionsModule_GetSiteUsersListSelections"
		params := map[string]interface{}{
			"SiteUsersId": siteUsersId,
			"TrackingId":  "DefaultTrackingID",
			"ListKey":     "Admin_GetAllTransactions",
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

	ex := loadSelections()
	admin.OverrideWithSelections(&q, ex)

	cfg := admin.ListSPConfig{
		ListKey:      "Admin_GetAllTransactions",
		TrackingID:   "DefaultTrackingID",
		ColumnMap:    map[string]string{},
		SearchFields: []string{},
	}

	spParams := admin.BuildListSPParams(q, siteUsersId, cfg)

	spName := "v1_AdminRole_TransactionsModule_GetAllTransactionsList"
	res, err := auth.ExecSP(db.DB, spName, spParams, 2)

	respond := func(listData []map[string]interface{}, total int) {
		// ✅ Path B: static columns EXACT like .NET
		columns := ColumnsTransactionsListAll()

		details := admin.BuildListDetails(columns, listData, q, total)
		c.JSON(http.StatusOK, gin.H{
			"status":  "1",
			"id":      siteUsersId,
			"errors":  []interface{}{},
			"details": details,
		})
	}

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

	listData := make([]map[string]interface{}, 0, len(rows))
	total := 0

	for _, row := range rows {
		item := admin.NormalizeRowKeys(row)
		FormatMoneyFields2dp(item)
		listData = append(listData, item)

		if v, ok := row["HowManyResults"]; ok && v != nil {
			switch t := v.(type) {
			case int:
				total = t
			case int64:
				total = int(t)
			case float64:
				total = int(t)
			default:
				i, _ := strconv.Atoi(fmt.Sprint(t))
				if i > 0 {
					total = i
				}
			}
		}
	}

	respond(listData, total)
}

func TransactionDetails(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"id":      0,
			"details": nil,
			"status":  "0",
			"errors": []gin.H{
				{"fieldName": "Authorization", "messageCode": "Unauthorized"},
			},
		})
		return
	}

	// .NET expects: ?transactionsId=123
	transactionsIdStr := c.Query("transactionsId")
	if transactionsIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"id":      0,
			"details": nil,
			"status":  "0",
			"errors": []gin.H{
				{"fieldName": "transactionsId", "messageCode": "Required"},
			},
		})
		return
	}

	transactionsId, err := strconv.Atoi(transactionsIdStr)
	if err != nil || transactionsId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"id":      0,
			"details": nil,
			"status":  "0",
			"errors": []gin.H{
				{"fieldName": "transactionsId", "messageCode": "Invalid"},
			},
		})
		return
	}

	// Call the same SP as .NET DB client:
	// v2_AdminRole_TransactionsModule_GetTransactionDetails
	res, err := auth.ExecSP(
		db.DB,
		"v2_AdminRole_TransactionsModule_GetTransactionDetails",
		map[string]interface{}{
			"SiteUsersId":                         user["id"],
			"CustomerAssetAccountsTransactionsId": transactionsId,
		},
		1,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"id":      0,
			"details": nil,
			"status":  "0",
			"errors": []gin.H{
				{"fieldName": "transactionsId", "messageCode": "SP_Error"},
			},
		})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"id":      0,
			"details": nil,
			"status":  "0",
			"errors": []gin.H{
				{"fieldName": "", "messageCode": "Invalid_SP_Response"},
			},
		})
		return
	}

	// Details is JSON string in SQL result (same pattern as your other endpoints)
	var details map[string]interface{}
	if detailsStr, ok := row["Details"].(string); ok && detailsStr != "" {
		_ = json.Unmarshal([]byte(detailsStr), &details)
	}
	if details != nil {
		if normalized, ok := normalizeJSONKeysLowerCamel(details).(map[string]interface{}); ok {
			details = normalized
		}
		FormatMoneyFields2dp(details)
	}

	// .NET returns:
	// { "Id": <int>, "Details": <object>, "Metadata": <array>, "Status": "1|0", "Errors": [...] }
	// (But your project generally uses lowercase keys. You can switch keys if your frontend expects lowercase.)
	c.JSON(http.StatusOK, gin.H{
		"id":       row["Id"],
		"details":  details,
		"metadata": TransactionDetailsMetadata,
		"status":   row["Status"],
		"errors":   []interface{}{},
	})
}

func ListPending(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	// .NET signature: ([FromQuery] QueryRecordListDto query, [FromQuery] bool bFilterToOwnAssignments)
	bFilterToOwnAssignments := strings.EqualFold(strings.TrimSpace(c.Query("bFilterToOwnAssignments")), "true")

	// Parse list query (filters/search/sort/pagination + clear flags)
	q := admin.ParseQueryRecordList(c.Request.URL.Query())

	// IMPORTANT to match .NET behavior:
	// If pageSize/pageNumber not present in query, .NET uses saved selections (often pageSize=50).
	// Your current ParseQueryRecordList sets defaults (10/1), which blocks override.
	// So: if query param missing, set to 0 to allow override with selections.
	if c.Query("pageSize") == "" {
		q.PageSize = 0
	}
	if c.Query("pageNumber") == "" {
		q.PageNumber = 0
	}

	if q.PageSize > 200 {
		q.PageSize = 200
	}

	// Load saved selections (.NET uses: Admin_GetPendingTransactions)
	loadSelections := func() *admin.ListSelections {
		sp := "v1_General_ListSelectionsModule_GetSiteUsersListSelections"
		params := map[string]interface{}{
			"SiteUsersId": siteUsersId,
			"TrackingId":  "DefaultTrackingID",
			"ListKey":     "Admin_GetPendingTransactions",
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

	ex := loadSelections()
	admin.OverrideWithSelections(&q, ex)

	// SP config (filters/sort/search whitelist)
	cfg := admin.ListSPConfig{
		ListKey:    "Admin_GetPendingTransactions",
		TrackingID: "DefaultTrackingID",
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
			// non-filterable/sortable columns intentionally omitted from whitelist
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
	}

	spParams := admin.BuildListSPParams(q, siteUsersId, cfg)

	// .NET extra param:
	spParams["bFilterToOwnAssignments"] = bFilterToOwnAssignments

	spName := "v1_AdminRole_PendingTransactionsModule_GetTransactionList"
	res, err := auth.ExecSP(db.DB, spName, spParams, 2)

	respond := func(listData []map[string]interface{}, total int) {
		columns := ColumnsPendingTransactionsList()
		details := admin.BuildListDetails(columns, listData, q, total)

		// ✅ .NET behavior: return the raw filters string (not structured object)
		// Example: "products__ProductName EQ (CrossRiverBank)"
		if raw := strings.TrimSpace(c.Query("filters")); raw != "" {
			details["filters"] = raw
		} else {
			// optional: if you want to always match .NET and return "" when no filters
			// details["filters"] = ""
		}

		c.JSON(http.StatusOK, gin.H{
			"id":      siteUsersId,
			"details": details,
			"status":  "1",
			"errors":  []interface{}{},
		})
	}

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

	listData := make([]map[string]interface{}, 0, len(rows))
	total := 0

	for _, row := range rows {
		item := admin.NormalizeRowKeys(row)
		FormatMoneyFields2dp(item)
		listData = append(listData, item)

		// results count
		if v, ok := row["HowManyResults"]; ok && v != nil {
			switch t := v.(type) {
			case int:
				total = t
			case int64:
				total = int(t)
			case float64:
				total = int(t)
			default:
				i, _ := strconv.Atoi(fmt.Sprint(t))
				if i > 0 {
					total = i
				}
			}
		}
	}

	respond(listData, total)
}

func ListProducts(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	res, err := auth.ExecSP(
		db.DB,
		"v1_AdminRole_PendingTransactionsModule_ListProducts",
		map[string]interface{}{
			"SiteUsersId": user["id"],
		},
		1,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid SP response"})
		return
	}

	// Details is JSON string in DBResultDto style
	var raw []pendingProductsRaw
	if detailsStr, ok := row["Details"].(string); ok && strings.TrimSpace(detailsStr) != "" {
		// some SPs return single-quoted JSON, same cleanup used in dashboard module
		cleanJSON := strings.ReplaceAll(detailsStr, "'", "\"")
		_ = json.Unmarshal([]byte(cleanJSON), &raw)
	}

	out := make([]pendingProductsOut, 0, len(raw))
	for _, r := range raw {
		out = append(out, pendingProductsOut{
			ProductId:               r.ProductId,
			ProductName:             r.ProductName,
			Asset:                   r.Asset,
			DisplayName:             r.DisplayName,
			Amount:                  r.Amount,
			PendingTransactionCount: r.PendingTransactionCount,
		})
	}

	c.JSON(http.StatusOK, pendingProductsResponse{
		Id:      row["Id"],
		Details: out,
		Status:  row["Status"],
		Errors:  []interface{}{},
	})
}

func ListPendingTreasury(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	// Parse list query
	q := admin.ParseQueryRecordList(c.Request.URL.Query())

	// allow selections to override if pageSize/pageNumber missing
	if c.Query("pageSize") == "" {
		q.PageSize = 0
	}
	if c.Query("pageNumber") == "" {
		q.PageNumber = 0
	}
	if q.PageSize > 200 {
		q.PageSize = 200
	}

	// ---------------------------
	// CHANGE THESE IF YOUR DB USES DIFFERENT VALUES
	// ---------------------------
	const listKey = "Admin_GetPendingTransactions_Treasury"
	const spListSelections = "v1_General_ListSelectionsModule_GetSiteUsersListSelections"
	const spName = "v1_AdminRole_PendingTransactionsTreasuryModule_GetTransactionList"
	// ---------------------------

	loadSelections := func() *admin.ListSelections {
		params := map[string]interface{}{
			"SiteUsersId": siteUsersId,
			"TrackingId":  "DefaultTrackingID",
			"ListKey":     listKey,
		}
		res, err := auth.ExecSP(db.DB, spListSelections, params, 1)
		if err != nil {
			return nil
		}
		row, ok := auth.AsSingleRow(res)
		if !ok {
			return nil
		}
		return admin.SelectionsFromRow(row)
	}

	ex := loadSelections()
	admin.OverrideWithSelections(&q, ex)

	// SP config whitelist (include Refference)
	cfg := admin.ListSPConfig{
		ListKey:    listKey,
		TrackingID: "DefaultTrackingID",
		ColumnMap: map[string]string{
			"customerAssetAccountsTransactions__Id":          "CustomerAssetAccountsTransactions.Id",
			"customerAssetAccountsTransactions__Date":        "CustomerAssetAccountsTransactions.Date",
			"customerAssetAccountsTransactions__Refference":  "CustomerAssetAccountsTransactions.Refference",
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
			"CustomerAssetAccountsTransactions.Refference",
			"Assets.Code",
			"Customers.CustomersCode",
			"CustomerUsers.FullName",
			"Customers.CompanyName",
			"LicenseesBrands.StatementDescriptor",
			"WidgetClientUsers.WidgetClientUsersCode",
			"FeeIncurringActions.DisplayName",
		},
	}

	spParams := admin.BuildListSPParams(q, siteUsersId, cfg)

	res, err := auth.ExecSP(db.DB, spName, spParams, 2)

	respond := func(listData []map[string]interface{}, total int) {
		columns := ColumnsPendingTransactionsTreasuryList()
		details := admin.BuildListDetails(columns, listData, q, total)

		// ✅ match .NET: keep filters as raw string or null
		if raw := strings.TrimSpace(c.Query("filters")); raw != "" {
			details["filters"] = raw
		} else {
			details["filters"] = nil
		}

		c.JSON(http.StatusOK, gin.H{
			"id":      siteUsersId,
			"details": details,
			"status":  "1",
			"errors":  []interface{}{},
		})
	}

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

	listData := make([]map[string]interface{}, 0, len(rows))
	total := 0

	for _, row := range rows {
		item := admin.NormalizeRowKeys(row)

		// If you already use FormatMoneyFields2dp in your list, keep it:
		FormatMoneyFields2dp(item)

		listData = append(listData, item)

		if v, ok := row["HowManyResults"]; ok && v != nil {
			switch t := v.(type) {
			case int:
				total = t
			case int64:
				total = int(t)
			case float64:
				total = int(t)
			default:
				i, _ := strconv.Atoi(fmt.Sprint(t))
				if i > 0 {
					total = i
				}
			}
		}
	}

	respond(listData, total)
}

func ListPendingTreasuryProducts(c *gin.Context) {
	// .NET response shows id: 0 always
	const responseId = 0

	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	siteUsersId := user["id"].(int)

	// ---------------------------
	// CHANGE THIS IF YOUR DB USES DIFFERENT SP NAME
	// ---------------------------
	const spName = "v1_AdminRole_PendingTransactionsTreasuryModule_ListProducts"
	// ---------------------------

	// ✅ FIX: SP expects @SiteUsersId
	params := map[string]interface{}{
		"SiteUsersId": siteUsersId,
	}

	res, err := auth.ExecSP(db.DB, spName, params, 1)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			c.JSON(http.StatusOK, pendingProductsResponse{
				Id:      responseId,
				Details: []pendingProductsOut{},
				Status:  "1",
				Errors:  []interface{}{},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "0",
			"error":  err.Error(),
		})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid SP response"})
		return
	}

	// Details is JSON string in DBResultDto style (same pattern as ListProducts)
	detailsStr := ""
	switch t := row["Details"].(type) {
	case string:
		detailsStr = t
	case []byte:
		detailsStr = string(t)
	}

	var raw []pendingProductsRaw
	if strings.TrimSpace(detailsStr) != "" {
		// some SPs return single-quoted JSON
		cleanJSON := strings.ReplaceAll(detailsStr, "'", "\"")
		_ = json.Unmarshal([]byte(cleanJSON), &raw)
	}

	out := make([]pendingProductsOut, 0, len(raw))
	for _, r := range raw {
		out = append(out, pendingProductsOut{
			ProductId:               r.ProductId,
			ProductName:             r.ProductName,
			Asset:                   r.Asset,
			DisplayName:             r.DisplayName,
			Amount:                  r.Amount,
			PendingTransactionCount: r.PendingTransactionCount,
		})
	}

	status := row["Status"]
	if status == nil {
		status = "1"
	}

	c.JSON(http.StatusOK, pendingProductsResponse{
		Id:      responseId,
		Details: out,
		Status:  status,
		Errors:  []interface{}{},
	})
}

func TransactionJSON(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	// Query params: Id + transactionsId (both appear in your URL)
	idStr := c.Query("Id")
	trxIdStr := c.Query("transactionsId")

	// Decide "requested id" for the "no result" response
	// In your sample, when no result: id = 249074 (the request value)
	requestedID := parseInt64Prefer(trxIdStr, idStr)

	if requestedID == 0 {
		// If caller didn't send a valid number, follow your existing API style
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"error":  "Missing or invalid Id/transactionsId",
		})
		return
	}

	siteUsersId := user["id"].(int)

	// ✅ SP NAME (adjust if your DB uses a slightly different name)
	// Based on your endpoint: /transactionsmodule/transaction-json

	const spName = "v1_AdminRole_TransactionsModule_GetTransactionJson"

	// ✅ Params: include both ids because your endpoint includes both.
	// Also include SiteUsersId because many AdminRole SPs require it.
	params := map[string]interface{}{
		"SiteUsersId": siteUsersId,
		// "Id":             requestedID,
		"CustomerAssetAccountsTransactionsId": requestedID,
	}

	res, err := auth.ExecSP(db.DB, spName, params, 1)
	if err != nil {
		// If SP returns no rows, your API still returns status=1 with details=null
		if err.Error() == "sql: no rows in result set" {
			c.JSON(http.StatusOK, transactionJSONResponse{
				Id:      requestedID,
				Details: nil,
				Status:  "1",
				Errors:  []interface{}{},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "0",
			"error":  err.Error(),
		})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok || row == nil {
		// Treat as "no result" to match your sample response
		c.JSON(http.StatusOK, transactionJSONResponse{
			Id:      requestedID,
			Details: nil,
			Status:  "1",
			Errors:  []interface{}{},
		})
		return
	}

	// Expected columns from SP (common patterns):
	// - "Id" or "id"
	// - "Details" or "details"
	respID := firstNonZeroInt(row, "Id", "id", "TransactionsId", "transactionsId")
	if respID == 0 {
		respID = requestedID
	}

	details := firstStringPtr(row, "Details", "details") // keep it as string JSON

	if details == nil || *details == "" {
		c.JSON(http.StatusOK, transactionJSONResponse{
			Id:      respID,
			Details: nil,
			Status:  "1",
			Errors:  []interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, transactionJSONResponse{
		Id:      respID,
		Details: *details, // IMPORTANT: details is a JSON string
		Status:  "1",
		Errors:  []interface{}{},
	})
}

func ListNotes(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	transactionsId, ok := mustInt64FromQueryOrForm(c.Query("transactionsId"))
	if !ok || transactionsId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"error":  "Missing or invalid transactionsId",
		})
		return
	}

	// ✅ Adjust SP name if your DB uses a different one
	const spName = "v1_AdminRole_TransactionsModule_GetNotesList"

	params := map[string]interface{}{
		"SiteUsersId":    siteUsersId,
		"TransactionsId": transactionsId,
	}

	res, err := auth.ExecSP(db.DB, spName, params, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok || row == nil {
		// no result => match your standard response
		c.JSON(http.StatusOK, gin.H{
			"id":      0,
			"details": gin.H{"notes": []interface{}{}},
			"status":  "1",
			"errors":  []interface{}{},
		})
		return
	}

	// ✅ The SP returns JSON in a column; your sample shows it's the 3rd column value.
	// We'll robustly find the FIRST string column that looks like JSON.
	jsonStr := ""
	for _, v := range row {
		s, ok := v.(string)
		if !ok {
			continue
		}
		ss := strings.TrimSpace(s)
		if strings.HasPrefix(ss, "{") || strings.HasPrefix(ss, "[") {
			jsonStr = ss
			break
		}
	}

	if jsonStr == "" {
		// If SP returned no JSON, respond empty
		c.JSON(http.StatusOK, gin.H{
			"id":      0,
			"details": gin.H{"notes": []interface{}{}},
			"status":  "1",
			"errors":  []interface{}{},
		})
		return
	}

	var env spNotesEnvelope
	if err := json.Unmarshal([]byte(jsonStr), &env); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "0",
			"error":  "Failed to parse notes JSON: " + err.Error(),
		})
		return
	}

	// Convert SP notes -> API notes (camelCase keys + id as string)
	out := make([]apiNoteItem, 0, len(env.Notes))
	for _, n := range env.Notes {
		out = append(out, apiNoteItem{
			Id:        strconv.Itoa(n.Id),
			Text:      n.Text,
			AddDate:   n.AddDate,
			AddedBy:   n.AddedBy,
			BEditable: n.BEditable,
			BPinned:   n.BPinned,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"id": 0,
		"details": gin.H{
			"notes": out,
		},
		"status": "1",
		"errors": []interface{}{},
	})

}

// helper for picking first existing key from a row
func firstNonNil(row map[string]interface{}, keys ...string) interface{} {
	for _, k := range keys {
		if v, ok := row[k]; ok && v != nil {
			return v
		}
	}
	return nil
}

func AddNote(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	// form-data fields
	transactionsIdStr := c.PostForm("transactionsId")
	text := c.PostForm("text")

	transactionsId, ok := mustInt64FromQueryOrForm(transactionsIdStr)
	if !ok || transactionsId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"error":  "Missing or invalid transactionsId",
		})
		return
	}
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"error":  "Missing text",
		})
		return
	}

	fullName := ""
	userCode := ""

	if v, ok := user["FullName"].(string); ok {
		fullName = v
	}
	if v, ok := user["UserCode"].(string); ok {
		userCode = v
	}

	addedBy := fullName
	if userCode != "" {
		addedBy = fullName + " (Admin - " + userCode + ")"
	}

	// ✅ Adjust SP name if your DB uses a different one
	const spName = "v1_AdminRole_TransactionsModule_AddNote"

	params := map[string]interface{}{
		"SiteUsersId":    siteUsersId,
		"TransactionsId": transactionsId,
		"Text":           text,
		"AddedBy":        addedBy,
	}

	res, err := auth.ExecSP(db.DB, spName, params, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "0",
			"error":  err.Error(),
		})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok || row == nil {
		// Treat as failure (SP should return id)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "0",
			"error":  "Invalid SP response",
		})
		return
	}

	// SP should return the new note id; key names may differ
	newId := asString(firstNonNil(row, "Id", "id", "NoteId", "noteId"))
	if newId == "" {
		newId = "0"
	}

	c.JSON(http.StatusOK, addNoteResponse{
		Id: newId,
		Details: addNoteDetails{
			TransactionsId: transactionsId,
			Text:           text,
		},
		Status: "1",
		Errors: []interface{}{},
	})
}

func EditNote(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	siteUsersId := user["id"].(int)
	// Form-data
	noteId := strings.TrimSpace(c.PostForm("id"))
	text := strings.TrimSpace(c.PostForm("text"))

	if noteId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"error":  "id is required",
		})
		return
	}
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"error":  "text is required",
		})
		return
	}

	// Optional: build EditedBy if your SP requires it
	editedBy := ""
	if v, ok := user["firstName"]; ok {
		editedBy = strings.TrimSpace(fmt.Sprint(v))
	}
	if v, ok := user["lastName"]; ok {
		ln := strings.TrimSpace(fmt.Sprint(v))
		if ln != "" {
			if editedBy != "" {
				editedBy += " "
			}
			editedBy += ln
		}
	}
	userCode := ""
	if v, ok := user["userCode"]; ok {
		userCode = strings.TrimSpace(fmt.Sprint(v))
	}
	if userCode != "" {
		if editedBy == "" {
			editedBy = fmt.Sprintf("(Admin - %s)", userCode)
		} else {
			editedBy = fmt.Sprintf("%s (Admin - %s)", editedBy, userCode)
		}
	}

	// ✅ Use the correct SP name from your DB (replace this with the exact one from .NET)
	// Example:
	spName := "v1_AdminRole_TransactionsModule_EditNote"

	params := map[string]interface{}{
		"SiteUsersId": siteUsersId,
		"NotesID":     noteId,
		"Text":        text,
		// If SP expects AddedBy, keep this:
		"EditedBy": editedBy,
	}

	res, err := auth.ExecSP(db.DB, spName, params, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "0",
			"error":  err.Error(),
		})
		return
	}

	// Many of your SPs return either:
	// 1) one row with edited note fields
	// 2) or nothing but success
	row, ok := auth.AsSingleRow(res)
	if !ok || row == nil {
		// still return expected response
		c.JSON(http.StatusOK, gin.H{
			"id": 0,
			"details": editNoteDetails{
				Id:        noteId,
				Text:      text,
				BEditable: false,
			},
			"status": "1",
			"errors": []interface{}{},
		})
		return
	}

	// bEditable might come from SP, otherwise default false
	bEditable := false
	if v, ok := row["bEditable"]; ok && v != nil {
		switch t := v.(type) {
		case bool:
			bEditable = t
		case int:
			bEditable = t != 0
		case int64:
			bEditable = t != 0
		case float64:
			bEditable = t != 0
		default:
			s := strings.TrimSpace(fmt.Sprint(t))
			bEditable = strings.EqualFold(s, "true") || s == "1"
		}
	}

	// id/text might also come from SP
	outId := noteId
	if v, ok := row["Id"]; ok && v != nil {
		outId = strings.TrimSpace(fmt.Sprint(v))
	}
	outText := text
	if v, ok := row["Text"]; ok && v != nil {
		outText = fmt.Sprint(v)
	}

	c.JSON(http.StatusOK, gin.H{
		"id": 0,
		"details": editNoteDetails{
			Id:        outId,
			Text:      outText,
			BEditable: bEditable,
		},
		"status": "1",
		"errors": []interface{}{},
	})
}

func PinNote(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	siteUsersId := user["id"].(int)

	var req pinNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"error":  "Invalid JSON body",
		})
		return
	}

	req.Id = strings.TrimSpace(req.Id)
	if req.Id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"error":  "id is required",
		})
		return
	}

	// Optional: build EditedBy if your SP requires it
	editedBy := ""
	if v, ok := user["firstName"]; ok {
		editedBy = strings.TrimSpace(fmt.Sprint(v))
	}
	if v, ok := user["lastName"]; ok {
		ln := strings.TrimSpace(fmt.Sprint(v))
		if ln != "" {
			if editedBy != "" {
				editedBy += " "
			}
			editedBy += ln
		}
	}
	userCode := ""
	if v, ok := user["userCode"]; ok {
		userCode = strings.TrimSpace(fmt.Sprint(v))
	}
	if userCode != "" {
		if editedBy == "" {
			editedBy = fmt.Sprintf("(Admin - %s)", userCode)
		} else {
			editedBy = fmt.Sprintf("%s (Admin - %s)", editedBy, userCode)
		}
	}

	// ✅ Replace with exact SP name from your .NET mapping if different
	spName := "v1_AdminRole_TransactionsModule_PinNote"

	params := map[string]interface{}{
		"SiteUsersId": siteUsersId,
		"NotesID":     req.Id,
		"bPinned":     req.BPinned,
		"EditedBy":    editedBy,
	}

	_, err := auth.ExecSP(db.DB, spName, params, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "0",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": 0,
		"details": pinNoteDetails{
			Id:      req.Id,
			BPinned: req.BPinned,
		},
		"status": "1",
		"errors": []interface{}{},
	})
}

func ListFrozenTransactions(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	siteUsersId := user["id"].(int)

	// ✅ required query param in .NET: customerAssetAccountsId
	customerAssetAccountsIdStr := c.Query("customerAssetAccountsId")
	if customerAssetAccountsIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"errors": []gin.H{
				{"fieldName": "customerAssetAccountsId", "messageCode": "Required"},
			},
		})
		return
	}

	customerAssetAccountsId, err := strconv.Atoi(customerAssetAccountsIdStr)
	if err != nil || customerAssetAccountsId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "0",
			"errors": []gin.H{
				{"fieldName": "customerAssetAccountsId", "messageCode": "Invalid"},
			},
		})
		return
	}

	// .NET: QueryRecordListDto bind (filters/sort/search/pageNumber/pageSize)
	q := admin.ParseQueryRecordList(c.Request.URL.Query())

	// .NET list endpoints usually cap pagesize (same as other Go endpoints)
	if q.PageSize > 200 {
		q.PageSize = 200
	}

	// .NET: load existing list selections:
	// _listSelectionsDbClient.GetSiteUsersListSelectionsAsync(siteUsersId, "DefaultTrackingID", "FrozenTransactions")
	loadSelections := func() *admin.ListSelections {
		sp := "v1_General_ListSelectionsModule_GetSiteUsersListSelections"
		params := map[string]interface{}{
			"SiteUsersId": siteUsersId,
			"TrackingId":  "DefaultTrackingID",
			"ListKey":     "FrozenTransactions",
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

	ex := loadSelections()
	admin.OverrideWithSelections(&q, ex)

	// Column map + search fields based on FrozenTransactionListDataRow attributes in .NET
	cfg := admin.ListSPConfig{
		ListKey:    "FrozenTransactions",
		TrackingID: "DefaultTrackingID",
		ColumnMap: map[string]string{
			"assets__Name": "Assets.Name",
			"assets__Code": "Assets.Code",
			"customerAssetAccountsTransactions__CustomerAssetAccountsTransactionsCode": "CustomerAssetAccountsTransactions.CustomerAssetAccountsTransactionsCode",
			"customerAssetAccountsTransactions__Date":                                  "CustomerAssetAccountsTransactions.Date",
			"transactionTypes__Type":                                                   "TransactionTypes.Type",
			"customerAssetAccountsTransactions__Amount":                                "CustomerAssetAccountsTransactions.Amount",
			"feeTransactions__Amount":                                                  "FeeTransactions.Amount",
			"trmLabsHelper__bSupportedCurrency":                                        "TRMLabsHelper.bSupportedCurrency",
			"customerAssetAccountsTransactions__bAwaitingUnfreeze":                     "CustomerAssetAccountsTransactions.bAwaitingUnfreeze",
		},
		SearchFields: []string{
			"Assets.Name",
			"Assets.Code",
			"CustomerAssetAccountsTransactions.CustomerAssetAccountsTransactionsCode",
			"TRMLabsHelper.bSupportedCurrency",
			"CustomerAssetAccountsTransactions.bAwaitingUnfreeze",
		},
	}

	// Build SP params (pagination + filters + sort + search + tracking/listkey + user)
	spParams := admin.BuildListSPParams(q, siteUsersId, cfg)

	// add required param: @CustomerAssetAccountsId
	spParams["CustomerAssetAccountsId"] = customerAssetAccountsId

	// .NET DB SP:
	// const string storedProcName = "v1_AdminRole_FrozenTransactionsModule_GetList";
	spName := "v1_AdminRole_FrozenTransactionsModule_GetList"
	res, err := auth.ExecSP(db.DB, spName, spParams, 2)

	respond := func(listData []map[string]interface{}, total int) {
		columns := ColumnsFrozenTransactionsList()
		details := admin.BuildListDetails(columns, listData, q, total)

		c.JSON(http.StatusOK, gin.H{
			"status":  "1",
			"id":      siteUsersId,
			"errors":  []interface{}{},
			"details": details,
		})
	}

	if err != nil {
		// same behavior as other list endpoints in this repo
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

	listData := make([]map[string]interface{}, 0, len(rows))
	total := 0

	for _, row := range rows {
		// convert "Assets__Name" -> "assets__Name" (lowercase first char)
		item := admin.NormalizeRowKeys(row)
		listData = append(listData, item)

		// total count (same as other modules)
		if v, ok := row["HowManyResults"]; ok && v != nil {
			switch t := v.(type) {
			case int:
				total = t
			case int64:
				total = int(t)
			case float64:
				total = int(t)
			default:
				i, _ := strconv.Atoi(fmt.Sprint(t))
				if i > 0 {
					total = i
				}
			}
		}
	}

	respond(listData, total)
}

func ViewAlerts(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	// .NET uses query param "Id"
	idStr := c.Query("Id")
	if idStr == "" {
		idStr = c.Query("id")
	}

	transactionsId := 0
	if idStr != "" {
		n, err := strconv.Atoi(idStr)
		if err != nil || n < 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"id":      0,
				"details": nil,
				"status":  "0",
				"errors": []gin.H{
					{"fieldName": "Id", "messageCode": "Invalid"},
				},
			})
			return
		}
		transactionsId = n
	}

	res, err := auth.ExecSP(
		db.DB,
		"v1_AdminRole_TransactionsModule_ViewAlerts",
		map[string]interface{}{
			"TransactionsId": transactionsId,
			"SiteUsersID":    user["id"],
		},
		1,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"id":      0,
			"details": nil,
			"status":  "0",
			"errors": []gin.H{
				{"fieldName": "Id", "messageCode": "SP_Error"},
			},
		})
		return
	}

	row, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"id":      0,
			"details": nil,
			"status":  "0",
			"errors": []gin.H{
				{"fieldName": "", "messageCode": "Invalid_SP_Response"},
			},
		})
		return
	}

	// Details from SP is usually JSON string
	var details map[string]interface{}
	if s, ok := row["Details"].(string); ok && s != "" {
		_ = json.Unmarshal([]byte(s), &details)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       row["Id"],
		"details":  details,
		"metadata": ViewAlertsMetadata,
		"status":   row["Status"],
		"errors":   []interface{}{},
	})
}
