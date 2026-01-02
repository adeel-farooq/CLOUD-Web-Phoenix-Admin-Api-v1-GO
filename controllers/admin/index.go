package admin

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
func GetDeposits(c *gin.Context) {

	user := auth.ExtractUser(c)
	if user == nil {
		c.JSON(401, gin.H{"message": "Unauthorized"})
		return
	}

	timeOption := c.Query("timeOption")

	res, err := auth.ExecSP(
		db.DB,
		"v1_AdminRole_DashboardModule_GetDeposits",
		map[string]interface{}{
			"SiteUsersId": user["id"],
			"TimeOption":  timeOption,
		},
		1, // single row
	)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	result := res.(map[string]interface{})

	// ---------- Parse Details JSON safely ----------
	var deposits interface{} = []interface{}{}

	if raw, ok := result["Details"].(string); ok && raw != "" {

		// 🔧 SQL JSON fix: single quotes → double quotes
		cleanJSON := strings.ReplaceAll(raw, "'", "\"")

		if err := json.Unmarshal([]byte(cleanJSON), &deposits); err != nil {
			c.JSON(500, gin.H{
				"error":   "Invalid deposits data",
				"details": err.Error(),
			})
			return
		}
	}

	// ---------- Final response ----------
	c.JSON(200, gin.H{
		"id":      result["Id"],
		"details": deposits,
		"status":  result["Status"],
		"errors":  []interface{}{},
		"message": "Deposits fetched successfully",
	})
}
