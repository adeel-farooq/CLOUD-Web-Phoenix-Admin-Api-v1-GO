package admin

import (
	"cloud-web-phoenix-customer-v1-go/controllers/auth"
	"cloud-web-phoenix-customer-v1-go/db"
	"encoding/json"
	"net/http"
	"strings"
	"time"

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

	// Call SP to get list data
	res, err := auth.ExecSP(
		db.DB,
		"v1_AdminRole_AdminUsersModule_List",
		map[string]interface{}{"User_SiteUsersID": siteUsersId},
		2,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Map rows into AdminUser slice
	users := []AdminUser{}
	if res != nil {
		for _, row := range res.([]map[string]interface{}) {
			users = append(users, AdminUser{
				ID:         int(row["AdminUsers__Id"].(int64)),
				Code:       row["AdminUsers__AdminUsersCode"].(string),
				Email:      row["SiteUsers__EmailAddress"].(string),
				FirstName:  row["AdminUsers__FirstName"].(string),
				LastName:   row["AdminUsers__LastName"].(string),
				Suppressed: row["SiteUsers__bSuppressed"].(bool),
				AddDate:    row["AdminUsers__AddDate"].(time.Time),
			})
		}
	}

	response := AdminUsersListResponse{
		ID:     siteUsersId,
		Status: "1",
		Errors: []string{},
		Details: AdminUsersListDetails{
			ListData:    users,
			SummaryRows: []interface{}{},
			Columns: []ColumnMetadata{
				{ColumnKey: "AdminUsers__Id", LabelKey: "Id", LabelValue: "Id", OrderNumber: 1, BSortable: true, BFilterable: true, BVisible: true, BLocked: false, Type: "Integer"},
				// add other column definitions...
			},
			PageNumber:   1,
			PageSize:     10,
			ResultsCount: len(users),
			Errors:       []string{},
			Metadata:     map[string]interface{}{},
		},
	}

	c.JSON(http.StatusOK, response)
}
