package dashboard

import (
	"cloud-web-phoenix-customer-v1-go/controllers/auth"
	"cloud-web-phoenix-customer-v1-go/db"
	"encoding/json"

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

	result, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(500, gin.H{"error": "Invalid SP response"})
		return
	}

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
	spName := getSPName(c.Request.URL.Path)
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

	result, ok := auth.AsSingleRow(res)
	if !ok {
		c.JSON(500, gin.H{"error": "Invalid SP response"})
		return
	}

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

	c.JSON(200, gin.H{
		"id":      result["Id"],
		"details": details,
		"status":  result["Status"],
		"errors":  []interface{}{},
		"message": message,
	})
}
