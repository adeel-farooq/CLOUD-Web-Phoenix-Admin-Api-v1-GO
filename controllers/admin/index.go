package admin

import (
	"cloud-web-phoenix-customer-v1-go/controllers/auth"
	"cloud-web-phoenix-customer-v1-go/db"
	"encoding/json"

	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func ProductsList(c *gin.Context) {

	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	res, err := auth.ExecSP(
		db.DB,
		"v1_AdminRole_DashboardModule_ListProducts",
		map[string]interface{}{
			"SiteUsersId": user["id"],
		},
		1,
	)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	result := res.(map[string]interface{})

	// ---- Parse Details JSON ----
	var products []map[string]interface{}
	if detailsStr, ok := result["Details"].(string); ok {
		_ = json.Unmarshal([]byte(detailsStr), &products)
	}

	c.JSON(200, gin.H{
		"id":      result["Id"],
		"details": products,
		"status":  result["Status"],
		"errors":  []interface{}{},
		"message": "Products fetched successfully",
	})
}

func GetChartsData(c *gin.Context) {

	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	timeOption := c.Query("timeOption")
	spName := getChartSPName(c.Request.URL.Path)
	if spName == "" {
		c.JSON(400, gin.H{"error": "Invalid chart endpoint"})
		return
	}

	params := map[string]interface{}{
		"SiteUsersId": user["id"],
	}
	if timeOption != "" {
		params["TimeOption"] = timeOption
	}
	res, err := auth.ExecSP(
		db.DB,
		spName,
		params,
		1, // single row
	)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	result := res.(map[string]interface{})

	// ---------- Parse Details JSON safely ----------
	var details interface{} = []interface{}{}

	if raw, ok := result["Details"].(string); ok && raw != "" {
		cleanJSON := strings.ReplaceAll(raw, "'", "\"")
		if err := json.Unmarshal([]byte(cleanJSON), &details); err != nil {
			c.JSON(500, gin.H{
				"error":   "Invalid details data",
				"details": err.Error(),
			})
			return
		}
	}
	message := "Chart data fetched successfully"
	if strings.Contains(c.Request.URL.Path, "usercounts") {
		message = "User counts fetched successfully"
	}

	c.JSON(200, gin.H{
		"id":      result["Id"],
		"details": details,
		"status":  result["Status"],
		"errors":  []interface{}{},
		"message": message,
	})
}
func GetAdminUsersList(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	siteUsersId := user["id"].(int)

	// Fetch admin users list (SP call via ExecSP)
	res, err := auth.ExecSP(
		db.DB,
		"v1_AdminRole_AdminUsersModule_List",
		map[string]interface{}{
			"User_SiteUsersID": siteUsersId,
			// Add more params if needed from query struct
		},
		2, // multi row
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	totalCount := 0
	listData := []map[string]interface{}{}
	if res != nil {
		rawList := res.([]map[string]interface{})
		for _, row := range rawList {
			item := map[string]interface{}{
				"adminUsers__AddDate":        row["AdminUsers__AddDate"],
				"adminUsers__AdminUsersCode": row["AdminUsers__AdminUsersCode"],
				"adminUsers__FirstName":      row["AdminUsers__FirstName"],
				"adminUsers__Id":             row["AdminUsers__Id"],
				"adminUsers__LastName":       row["AdminUsers__LastName"],
				"siteUsers__bSuppressed":     row["SiteUsers__bSuppressed"],
				"siteUsers__EmailAddress":    row["SiteUsers__EmailAddress"],
			}
			if v, ok := row["HowManyResults"].(int64); ok {
				totalCount = int(v)
			}

			listData = append(listData, item)
		}
	}

	// Prepare columns metadata (static, can be moved to helper)
	columns := []map[string]interface{}{
		{"columnKey": "AdminUsers__Id", "labelKey": "Id", "labelValue": "Id", "orderNumber": 1, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "Integer", "filterMetadata": map[string]interface{}{"details": nil, "filterType": "Amount"}},
		{"columnKey": "AdminUsers__AdminUsersCode", "labelKey": "AdminUsersCode", "labelValue": "Code", "orderNumber": 2, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "String", "filterMetadata": map[string]interface{}{"details": nil, "filterType": "TextContains"}},
		{"columnKey": "SiteUsers__EmailAddress", "labelKey": "EmailAddress", "labelValue": "Email Address", "orderNumber": 3, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "String", "filterMetadata": map[string]interface{}{"details": nil, "filterType": "TextContains"}},
		{"columnKey": "AdminUsers__FirstName", "labelKey": "FirstName", "labelValue": "First Name", "orderNumber": 4, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "String", "filterMetadata": map[string]interface{}{"details": nil, "filterType": "TextContains"}},
		{"columnKey": "AdminUsers__LastName", "labelKey": "LastName", "labelValue": "Last Name", "orderNumber": 5, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "String", "filterMetadata": map[string]interface{}{"details": nil, "filterType": "TextContains"}},
		{"columnKey": "SiteUsers__bSuppressed", "labelKey": "bSuppressed", "labelValue": "Suppressed", "orderNumber": 6, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "Boolean", "filterMetadata": map[string]interface{}{"filterType": "SingleChoice", "details": map[string]interface{}{"PossibleValues": []map[string]string{{"value": "0", "label": "Active"}, {"value": "1", "label": "Inactive"}}}}},
		{"columnKey": "AdminUsers__AddDate", "labelKey": "AddDate", "labelValue": "Add Date", "orderNumber": 7, "tooltip": nil, "bSortable": true, "bFilterable": true, "bVisible": true, "bLocked": false, "type": "DateTime", "filterMetadata": map[string]interface{}{"details": map[string]interface{}{"start": nil, "end": nil}, "filterType": "DateTime:Range"}},
	}

	// Use totalCount variable for resultsCount, since HowManyRows is not in listData anymore
	resultsCount := totalCount
	details := map[string]interface{}{
		"listData":        listData,
		"summaryRows":     []interface{}{},
		"columns":         columns,
		"pageNumber":      1,
		"pageSize":        10,
		"filters":         nil,
		"sortBy":          nil,
		"searchString":    nil,
		"bHasSearchField": true,
		"customColumns":   nil,
		"resultsCount":    resultsCount,
		"metadata":        map[string]interface{}{},
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      siteUsersId,
		"details": details,
		"status":  "1",
		"errors":  []string{},
	})

	// removed obsolete struct-based response, now using gin.H response above
}
