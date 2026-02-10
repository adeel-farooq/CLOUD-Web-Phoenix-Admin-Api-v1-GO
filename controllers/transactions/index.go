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
