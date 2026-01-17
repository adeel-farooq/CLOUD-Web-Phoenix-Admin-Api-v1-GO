package admin

import (
	"cloud-web-phoenix-customer-v1-go/controllers/auth"
	"cloud-web-phoenix-customer-v1-go/controllers/dashboard"

	"cloud-web-phoenix-customer-v1-go/db"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetUserCounts(c *gin.Context) {

	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	timeOption := c.Query("timeOption")
	spName := "v1_AdminRole_DashboardModule_GetUserCounts"
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

	c.JSON(200, gin.H{
		"id":      result["Id"],
		"details": details,
		"status":  result["Status"],
		"errors":  []interface{}{},
		"message": "User counts fetched successfully",
	})
}
func GetAdminUsersData(c *gin.Context) {
	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}
	siteUsersId := user["id"].(int)

	// Use helper to build SQL-style SP params
	spParams := dashboard.BuildSPParams(c.Request.URL.Query(), siteUsersId, c.Request.URL.Path)

	columns := dashboard.GetAdminUsersColumns(c.Request.URL.Path)
	spName := "v1_AdminRole_AdminUsersModule_List"
	res, err := auth.ExecSP(
		db.DB,
		spName,
		spParams,
		2, // multi row
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			details := map[string]interface{}{
				"listData":        []interface{}{},
				"summaryRows":     []interface{}{},
				"columns":         columns,
				"pageNumber":      spParams["PageNumber"],
				"pageSize":        spParams["PageSize"],
				"filters":         c.Query("filters"),
				"sortBy":          nil,
				"searchString":    nil,
				"bHasSearchField": true,
				"customColumns":   nil,
				"resultsCount":    0,
				"errors":          []string{"No records found."},
				"metadata":        map[string]interface{}{},
			}
			c.JSON(http.StatusOK, gin.H{
				"details": details,
				"id":      siteUsersId,
				"status":  "1",
				"errors":  []string{},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	totalCount := 0
	listData := []map[string]interface{}{}
	if res != nil {
		rawList := res.([]map[string]interface{})
		for _, row := range rawList {
			item := dashboard.GetAdminUsersListData(c.Request.URL.Path, row)
			if v, ok := row["HowManyResults"].(int64); ok {
				totalCount = int(v)
			}
			listData = append(listData, item)
		}
	}

	// Extract values from spParams for response
	searchString := c.Query("search")
	details := map[string]interface{}{
		"bHasSearchField": true,
		"columns":         columns,
		"customColumns":   nil,
		"filters":         c.Query("filters"), // original filter string
		"listData":        listData,
		"metadata":        map[string]interface{}{},
		"pageNumber":      spParams["PageNumber"],
		"pageSize":        spParams["PageSize"],
		"resultsCount":    totalCount,
		"searchString":    searchString,
		"sortBy":          nil,
		"summaryRows":     []interface{}{},
	}

	c.JSON(http.StatusOK, gin.H{
		"details": details,
		"errors":  []string{},
		"id":      siteUsersId,
		"status":  "1",
	})

}
