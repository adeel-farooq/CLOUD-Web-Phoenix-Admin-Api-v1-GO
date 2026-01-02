package auth

import (
	"cloud-web-phoenix-customer-v1-go/db"
	"cloud-web-phoenix-customer-v1-go/pkg"

	"net/http"

	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func SignIn(c *gin.Context) {
	var requestBody SignInRequest
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		var details interface{}
		if errs, ok := err.(validator.ValidationErrors); ok {
			details = errs.Error()
		} else {
			details = err.Error()
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": details,
		})
		return
	}

	// Use global DB connection
	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
		return
	}

	// Map AccountType int to string for proc
	accountTypeStr := "Customer"
	if requestBody.AccountType == 0 {
		accountTypeStr = "Admin"
	}

	query := "exec v1_PublicRole_AuthModule_GetSiteUsersAuthData @Username=@Username, @AccountType=@AccountType"
	rows, err := db.DB.Query(query,
		sql.Named("Username", requestBody.Username),
		sql.Named("AccountType", accountTypeStr),
	)
	if err != nil {
		pkg.Log("[DB QUERY ERROR]", err)
		sendAuthError(c)
		return
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	var result map[string]interface{}

	if rows.Next() {
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			pkg.Log("[DB SCAN ERROR]", err)
			sendAuthError(c)
			return
		}

		result = make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				result[col] = string(b)
			} else {
				result[col] = val
			}
		}
	} else {
		// ✅ No user found
		sendAuthError(c)
		return
	}

	// ✅ Password check
	if passHash, ok := result["PasswordHash"].(string); ok {
		ok, err := VerifyPassword(requestBody.Password, passHash)
		if err != nil || !ok {
			sendAuthError(c)
			return
		}
	} else {
		sendAuthError(c)
		return
	}

	id := result["SiteUsersId"]

	details := gin.H{
		"accessToken":              nil,
		"expiresIn":                0,
		"refreshToken":             nil,
		"refreshTokenExpiresIn":    0,
		"bTwoFactorAppAuthEnabled": result["bTwoFactorAppAuthEnabled"],
		"bTwoFactorSMSAuthEnabled": result["bTwoFactorSMSAuthEnabled"],
		"bEmailVerified":           result["bEmailVerified"],
		"bSuppressed":              result["bSuppressed"],
		"accountType":              result["AccountType"],
	}

	// ✅ Success response
	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"details": details,
		"status":  "1",
		"errors":  []string{},
	})
}

// Helper for error response
func sendAuthError(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"id":      0,
		"details": nil,
		"status":  "0",
		"errors": []gin.H{
			{"fieldName": "Username", "messageCode": "Username_Or_Password_Incorrect"},
			{"fieldName": "Password", "messageCode": "Username_Or_Password_Incorrect"},
		},
	})
}
