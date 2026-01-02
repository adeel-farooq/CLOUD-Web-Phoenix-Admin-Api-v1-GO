package auth

import (
	"cloud-web-phoenix-customer-v1-go/db"
	"cloud-web-phoenix-customer-v1-go/pkg"
	"strings"

	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
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

	query := "exec v1_PublicRole_AuthModule_GetSiteUsersAuthData @Username=@Username, @AccountType='Admin'"
	rows, err := db.DB.Query(query,
		sql.Named("Username", requestBody.Username),
	)
	if err != nil {
		pkg.Log("[DB QUERY ERROR]", err)
		SendAuthError(c, 0)
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
			SendAuthError(c, 0)
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
		SendAuthError(c, 0)
		return
	}

	// ✅ Password check
	if passHash, ok := result["PasswordHash"].(string); ok {
		ok, err := VerifyPassword(requestBody.Password, passHash)
		if err != nil || !ok {
			SendAuthError(c, 0)
			return
		}
	} else {
		SendAuthError(c, 0)
		return
	}
	id := 0
	details := gin.H{}
	if requestBody.TFAType != "" && requestBody.TFACode != "" {
		// write login logic with TFA
		var secret string
		if v, ok := result["TotpSharedSecret"]; ok && v != nil {
			switch val := v.(type) {
			case string:
				secret = val
			case []byte:
				secret = string(val)
			default:
				// Unexpected type → reject gracefully
				SendAuthError(c, 2)
				return
			}
		} else {
			// Secret not configured → reject TFA login
			SendAuthError(c, 2)
			return
		}

		status := ValidateTfaCode(
			requestBody.TFAType,
			requestBody.TFACode,
			map[string]interface{}{
				"bTwoFactorAppAuthEnabled": result["bTwoFactorAppAuthEnabled"],
				"bTwoFactorSMSAuthEnabled": result["bTwoFactorSMSAuthEnabled"],
			},
			true,
			secret,
		)
		if status == TfaTypeInvalid {
			SendAuthError(c, 2)
			return
		}
		if status != Success {
			SendAuthError(c, 1)
			return
		} else {
			if v, ok := result["SiteUsersId"].(int64); ok {
				id = int(v)
			}
			// ✅ Safely extract user info for token
			firstName, _ := result["FirstName"].(string)
			lastName, _ := result["LastName"].(string)
			accountType := "Admin"
			accessToken, refreshToken, expiresIn, refreshTokenExpiresIn, err := GenerateTokens(id, requestBody.RememberMe, firstName, lastName, accountType)
			if err != nil {
				SendAuthError(c, 0)
				return
			}

			// ✅ Prepare user details with tokens

			details = PrepareUserDetails(result, accessToken, refreshToken, expiresIn, refreshTokenExpiresIn)
		}

	} else {

		if v, ok := result["SiteUsersId"].(int64); ok {
			id = int(v)
		}
		delete(result, "NavigationScopesJson") // Remove sensitive info
		details = PrepareUserDetails(result, nil, nil, 0, 0)
	}

	// ✅ Success response
	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"details": details,
		"status":  "1",
		"errors":  []string{},
	})
}

func UserFromToken(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")

	var jwtToken *jwt.Token
	var err error
	if tokenString != "" {
		// Handle both 'Bearer <token>' and raw token
		token := tokenString
		if len(tokenString) > 7 && strings.ToLower(tokenString[0:7]) == "bearer " {
			token = strings.TrimSpace(tokenString[7:])
		}
		jwtToken, _, err = new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		user := ExtractUser(jwtToken)
		if user == nil {
			c.JSON(http.StatusOK, gin.H{"details": nil, "id": 0, "status": "1", "errors": []string{}})
		} else {
			c.JSON(http.StatusOK, gin.H{"details": user, "id": 0, "status": "1", "errors": []string{}})
		}
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
	}

}
