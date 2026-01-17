package auth

import (
	"cloud-web-phoenix-customer-v1-go/db"

	"net/http"

	"github.com/gin-gonic/gin"
)

func SignIn(c *gin.Context) {

	// ---------- Bind Request ----------
	var requestBody SignInRequest
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
		return
	}

	// ---------- Call Stored Procedure (Single Row) ----------
	res, err := ExecSP(
		db.DB,
		"v1_PublicRole_AuthModule_GetSiteUsersAuthData",
		map[string]interface{}{
			"Username":    requestBody.Username,
			"AccountType": "Admin",
		},
		1, // 1 = single row
	)

	if err != nil {
		SendAuthError(c, 0)
		return
	}

	result, ok := AsSingleRow(res)
	if !ok {
		SendAuthError(c, 0)
		return
	}

	// ---------- Password Validation ----------
	passHash, ok := result["PasswordHash"].(string)
	if !ok {
		SendAuthError(c, 0)
		return
	}

	okPass, err := VerifyPassword(requestBody.Password, passHash)
	if err != nil || !okPass {
		SendAuthError(c, 0)
		return
	}

	// ---------- Extract User ID ----------
	id := 0
	if v, ok := result["SiteUsersId"].(int64); ok {
		id = int(v)
	} else {
		SendAuthError(c, 0)
		return
	}

	// ---------- TFA FLOW ----------
	if requestBody.TFAType != "" && requestBody.TFACode != "" {

		var secret string
		if v, ok := result["TotpSharedSecret"]; ok && v != nil {
			switch s := v.(type) {
			case string:
				secret = s
			case []byte:
				secret = string(s)
			default:
				SendAuthError(c, 2)
				return
			}
		} else {
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
		}

		// ---------- Generate Tokens ----------
		firstName, _ := result["FirstName"].(string)
		lastName, _ := result["LastName"].(string)
		accountType := "Admin"

		accessToken, refreshToken, expiresIn, refreshTokenExpiresIn, err :=
			GenerateTokens(id, requestBody.RememberMe, firstName, lastName, accountType)

		if err != nil {
			SendAuthError(c, 0)
			return
		}

		details := PrepareUserDetails(
			result,
			accessToken,
			refreshToken,
			expiresIn,
			refreshTokenExpiresIn,
		)

		c.JSON(http.StatusOK, gin.H{
			"id":      id,
			"details": details,
			"status":  "1",
			"errors":  []string{},
		})
		return
	}

	// ---------- NO TFA (First Step Login) ----------
	delete(result, "NavigationScopesJson") // sensitive

	details := PrepareUserDetails(result, nil, nil, 0, 0)

	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"details": details,
		"status":  "1",
		"errors":  []string{},
	})
}

func UserFromToken(c *gin.Context) {
	user := ExtractUser(c)

	if user == nil {
		c.JSON(http.StatusOK, gin.H{
			"details": nil,
			"id":      0,
			"status":  "0",
			"errors":  []string{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"details": user,
		"id":      user["id"],
		"status":  "1",
		"errors":  []string{},
	})
}
