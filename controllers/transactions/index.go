package transactions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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

	// .NET: load list selections for this endpoint
	// .NET uses ListKey: "Admin_CustomerAssetAccounts"
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

	// Endpoint-specific whitelist for filtering/sorting/search
	// (These table/field names are best-guess based on DTO; if your SP expects different aliases,
	// we’ll adjust after first run using SP logs)
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

	// Call same SP as .NET
	spName := "v1_AdminRole_TransactionsModule_GetTransactionList"
	res, err := auth.ExecSP(db.DB, spName, spParams, 2)

	respond := func(columns []map[string]interface{}, listData []map[string]interface{}, total int) {
		details := admin.BuildListDetails(columns, listData, q, total)
		c.JSON(http.StatusOK, gin.H{
			"status":  "1",
			"id":      siteUsersId,
			"errors":  []string{},
			"details": details,
		})
	}

	if err != nil {
		// return empty list if no rows (same pattern used in admin module)
		if err.Error() == "sql: no rows in result set" {
			respond([]map[string]interface{}{}, []map[string]interface{}{}, 0)
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

	// If SP returns no rows
	if len(rows) == 0 {
		respond([]map[string]interface{}{}, []map[string]interface{}{}, 0)
		return
	}

	// Build columns automatically from first row (so frontend gets columns)
	columns := admin.BuildColumnsFromSPRow(rows[0], 1)

	// Normalize row keys to match the repo style (.NET camelCase output)
	listData := make([]map[string]interface{}, 0, len(rows))
	total := 0

	for _, row := range rows {
		item := admin.NormalizeRowKeys(row)
		listData = append(listData, item)

		// total count (HowManyResults)
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

	respond(columns, listData, total)
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
		ColumnMap:    map[string]string{}, // we’ll tighten this once we see actual row keys
		SearchFields: []string{
			// add later once confirmed
		},
	}

	spParams := admin.BuildListSPParams(q, siteUsersId, cfg)

	spName := "v1_AdminRole_TransactionsModule_GetAllTransactionsList"
	res, err := auth.ExecSP(db.DB, spName, spParams, 2)

	respond := func(columns []map[string]interface{}, listData []map[string]interface{}, total int) {
		details := admin.BuildListDetails(columns, listData, q, total)
		c.JSON(http.StatusOK, gin.H{
			"status":  "1",
			"id":      siteUsersId,
			"errors":  []string{},
			"details": details,
		})
	}

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			respond([]map[string]interface{}{}, []map[string]interface{}{}, 0)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": err.Error()})
		return
	}

	rows, ok := res.([]map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "0", "error": "Invalid SP response"})
		return
	}

	if len(rows) == 0 {
		respond([]map[string]interface{}{}, []map[string]interface{}{}, 0)
		return
	}

	columns := admin.BuildColumnsFromSPRow(rows[0], 1)

	listData := make([]map[string]interface{}, 0, len(rows))
	total := 0

	for _, row := range rows {
		item := admin.NormalizeRowKeys(row)
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

	respond(columns, listData, total)
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
