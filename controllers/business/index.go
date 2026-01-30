// index.go
package business

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"cloud-web-phoenix-customer-v1-go/controllers/admin"
	"cloud-web-phoenix-customer-v1-go/controllers/auth"
	"cloud-web-phoenix-customer-v1-go/db"

	"github.com/gin-gonic/gin"
)

func GetOperationalAssetAccountsList(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	siteUsersId := user["id"].(int)

	// 1) Parse query (page/size/sort/filter/search + clear flags)
	q := admin.ParseQueryRecordList(c.Request.URL.Query())

	// 2) .NET flow: Load selections & override
	// ListKey .NET style: Admin_OperationalAssetAccounts
	ex := admin.LoadListSelections(siteUsersId, "Admin_OperationalAssetAccounts")
	admin.OverrideWithSelections(&q, ex)

	// 3) Config (ColumnMap/SearchFields) for Filters/Sort/Search SQL generation
	cfg := admin.ListSPConfig{
		ListKey:      "Admin_OperationalAssetAccounts",
		TrackingID:   "DefaultTrackingID",
		ColumnMap:    OperationalAssetAccountsColumnMap(),
		SearchFields: OperationalAssetAccountsSearchFields(),
	}

	// 4) Build SP params (NET-like: NULLs where needed)
	spName := "v1_AdminRole_OperationalAssetAccountsModule_GetAccountList"
	spParams := BuildOperationalAssetAccountsSPParamsNetLike(q, siteUsersId, cfg)

	// 5) Execute
	res, err := auth.ExecSP(db.DB, spName, spParams, 2)

	// ✅ .NET list style: SP error -> still 200 with details.errors
	if err != nil {
		cols := OperationalAssetAccountsColumns() // fixed columns (recommended)
		cols = NormalizeColumnsNetLike(cols)

		c.JSON(200, gin.H{
			"id":     siteUsersId,
			"status": "1",
			"errors": []string{},
			"details": BuildListDetailsNetLike(
				cols,
				[]map[string]interface{}{},
				q,
				0,
				[]map[string]string{{"message": err.Error()}},
			),
		})
		return
	}

	rows := admin.AsRows(res)

	total := 0
	if len(rows) > 0 {
		if v, ok := rows[0]["HowManyResults"]; ok {
			total = toInt(v)
		}
	}

	// 6) listData: .NET jaisa keys normalize + values normalize (decimals/datetime)
	listData := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		item := NormalizeRowNetLike(row) // removes RowNum/HowManyResults, lowercases first char, fixes decimals
		listData = append(listData, item)
	}

	cols := OperationalAssetAccountsColumns()
	cols = NormalizeColumnsNetLike(cols)

	// 7) Response
	c.JSON(200, gin.H{
		"id":     siteUsersId,
		"status": "1",
		"errors": []string{},
		"details": BuildListDetailsNetLike(
			cols,
			listData,
			q,
			total,
			[]map[string]string{},
		),
	})

}

func GetCustomerAssetAccountsList(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	siteUsersId := user["id"].(int)

	// 1) Query
	q := admin.ParseQueryRecordList(c.Request.URL.Query())

	// 2) .NET flow: saved selections -> override
	ex := admin.LoadListSelections(siteUsersId, "Admin_CustomerAssetAccounts")
	admin.OverrideWithSelections(&q, ex)

	// 3) SP + config
	spName := "v1_AdminRole_CustomerAssetAccountsModule_GetAccountList" // <-- verify in DB if different

	cfg := admin.ListSPConfig{
		ListKey:      "Admin_CustomerAssetAccounts",
		TrackingID:   "DefaultTrackingID",
		ColumnMap:    CustomerAssetAccountsColumnMap(),
		SearchFields: CustomerAssetAccountsSearchFields(),
	}

	// 4) Params (NET-like NULL behavior)
	spParams := BuildCustomerAssetAccountsSPParamsNetLike(q, siteUsersId, cfg)

	// 5) Call
	res, err := auth.ExecSP(db.DB, spName, spParams, 2)

	// ✅ .NET style: list endpoint error -> still 200 with details.errors
	if err != nil {
		cols := NormalizeColumnsNetLike(CustomerAssetAccountsColumns())
		c.JSON(200, gin.H{
			"id":     siteUsersId,
			"status": "1",
			"errors": []string{},
			"details": BuildListDetailsNetLike(
				cols,
				[]map[string]interface{}{},
				q,
				0,
				[]map[string]string{{"message": err.Error()}},
			),
		})
		return
	}

	rows := admin.AsRows(res)

	total := 0
	if len(rows) > 0 {
		if v, ok := rows[0]["HowManyResults"]; ok {
			total = toInt(v)
		}
	}

	// 6) listData net-like
	listData := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		listData = append(listData, NormalizeRowNetLike(row))
	}

	cols := NormalizeColumnsNetLike(CustomerAssetAccountsColumns())

	c.JSON(200, gin.H{
		"id":     siteUsersId,
		"status": "1",
		"errors": []string{},
		"details": BuildListDetailsNetLike(
			cols,
			listData,
			q,
			total,
			[]map[string]string{},
		),
	})

}

func GetBusinessList(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	siteUsersId := user["id"].(int)

	// query params
	q := admin.ParseQueryRecordList(c.Request.URL.Query())

	// ✅ .NET flow: list selections ALWAYS load, then override
	// .NET: GetSiteUsersListSelectionsAsync(..., "Business") then query.OverrideWithExistingListSelections(...)
	ex := admin.LoadListSelections(siteUsersId, "Business")
	admin.OverrideWithSelections(&q, ex)

	// ✅ SP
	spName := "v1_AdminRole_BusinessModule_List"

	cfg := admin.ListSPConfig{
		ListKey:      "Business",
		TrackingID:   "DefaultTrackingID",
		ColumnMap:    BusinessColumnMap(),
		SearchFields: BusinessSearchFields(),
	}

	// ✅ VERY IMPORTANT: empty -> NULL (NOT empty string)
	spParams := BuildBusinessListSPParamsNetLike(q, siteUsersId, cfg)

	res, err := auth.ExecSP(db.DB, spName, spParams, 2)

	// ✅ .NET style: list endpoint pe SP error => 200 with details.errors (status "1")
	if err != nil {
		cols := NormalizeColumnsDetailsNull(BusinessColumns())

		c.JSON(200, gin.H{
			"id":     siteUsersId,
			"status": "1",
			"errors": []string{},
			"details": BuildListDetailsNetLike(
				cols,
				[]map[string]interface{}{},
				q,
				0,
				[]map[string]string{{"message": err.Error()}},
			),
		})
		return
	}

	rows := admin.AsRows(res)

	total := 0
	if len(rows) > 0 {
		if v, ok := rows[0]["HowManyResults"]; ok {
			total = toInt(v)
		}
	}

	// ✅ listData: .NET response me keys lowercase-first hoti hain
	listData := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		item := admin.NormalizeRowKeys(row) // removes RowNum/HowManyResults + lowercases first char
		listData = append(listData, item)
	}

	cols := NormalizeColumnsDetailsNull(BusinessColumns())

	c.JSON(200, gin.H{
		"id":     siteUsersId,
		"status": "1",
		"errors": []string{},
		"details": BuildListDetailsNetLike(
			cols,
			listData,
			q,
			total,
			[]map[string]string{},
		),
	})
}
func GetCustomerList(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	siteUsersId := user["id"].(int)

	// query params
	q := admin.ParseQueryRecordList(c.Request.URL.Query())

	ex := admin.LoadListSelections(siteUsersId, "CustomerUsers")
	admin.OverrideWithSelections(&q, ex)

	// ✅ SP
	spName := "v1_AdminRole_CustomersModule_List"

	cfg := admin.ListSPConfig{
		ListKey:      "CustomerUsers",
		TrackingID:   "DefaultTrackingID",
		ColumnMap:    CustomerUsersColumnMap(),
		SearchFields: CustomerUsersSearchFields(),
	}

	// ✅ VERY IMPORTANT: empty -> NULL (NOT empty string)
	spParams := BuildCustomerListSPParamsNetLike(q, siteUsersId, cfg)

	res, err := auth.ExecSP(db.DB, spName, spParams, 2)

	// ✅ .NET style: list endpoint pe SP error => 200 with details.errors (status "1")
	if err != nil {
		cols := NormalizeColumnsDetailsNull(CustomerColumns())

		c.JSON(200, gin.H{
			"id":     siteUsersId,
			"status": "1",
			"errors": []string{},
			"details": BuildListDetailsNetLike(
				cols,
				[]map[string]interface{}{},
				q,
				0,
				[]map[string]string{{"message": err.Error()}},
			),
		})
		return
	}

	rows := admin.AsRows(res)

	total := 0
	if len(rows) > 0 {
		if v, ok := rows[0]["HowManyResults"]; ok {
			total = toInt(v)
		}
	}

	// ✅ listData: .NET response me keys lowercase-first hoti hain
	listData := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		item := admin.NormalizeRowKeys(row) // removes RowNum/HowManyResults + lowercases first char
		listData = append(listData, item)
	}

	cols := NormalizeColumnsDetailsNull(CustomerColumns())

	c.JSON(200, gin.H{
		"id":     siteUsersId,
		"status": "1",
		"errors": []string{},
		"details": BuildListDetailsNetLike(
			cols,
			listData,
			q,
			total,
			[]map[string]string{},
		),
	})
}
func GetProductList(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	// This endpoint returns all products; SP call should not receive list/query params.
	siteUsersId := 0

	// ✅ SP (no params)
	spName := "v1_AdminRole_CustomersModule_GetAllProducts"
	res, err := auth.ExecSP(db.DB, spName, nil, 2)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(200, gin.H{
				"id":      siteUsersId,
				"details": []map[string]interface{}{},
				"status":  "1",
				"errors":  []string{},
			})
			return
		}
		c.JSON(200, gin.H{
			"id":      siteUsersId,
			"details": []map[string]interface{}{},
			"status":  "0",
			"errors":  []string{err.Error()},
		})
		return
	}

	rows := admin.AsRows(res)

	// Some SPs return a single row where Details is a JSON string (array).
	// In that case, parse it and return the parsed list directly.
	if len(rows) == 1 {
		var detailsRaw string
		if v, ok := rows[0]["Details"]; ok && v != nil {
			detailsRaw = strings.TrimSpace(toString(v))
		} else if v, ok := rows[0]["details"]; ok && v != nil {
			detailsRaw = strings.TrimSpace(toString(v))
		}
		if detailsRaw != "" {
			var parsed interface{}
			b := []byte(detailsRaw)
			if err := json.Unmarshal(b, &parsed); err != nil {
				// Some SPs return JSON using single quotes
				clean := strings.ReplaceAll(detailsRaw, "'", "\"")
				_ = json.Unmarshal([]byte(clean), &parsed)
			}

			if arr, ok := parsed.([]interface{}); ok {
				out := make([]map[string]interface{}, 0, len(arr))
				for _, it := range arr {
					if m, ok := it.(map[string]interface{}); ok {
						out = append(out, NormalizeRowNetLike(m))
					}
				}
				c.JSON(200, gin.H{
					"id":      siteUsersId,
					"details": out,
					"status":  "1",
					"errors":  []string{},
				})
				return
			}
		}
	}

	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, NormalizeRowNetLike(row))
	}

	c.JSON(200, gin.H{
		"id":      siteUsersId,
		"details": out,
		"status":  "1",
		"errors":  []string{},
	})
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}
