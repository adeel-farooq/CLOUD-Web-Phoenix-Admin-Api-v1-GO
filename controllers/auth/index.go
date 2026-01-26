package auth

import (
	"cloud-web-phoenix-customer-v1-go/db"
	"time"

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

// func TfaLoginOld(c *gin.Context) {

// 	// ---------- Bind Request ----------
// 	var requestBody SignInRequest
// 	if err := c.ShouldBindJSON(&requestBody); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error":   "Invalid request body",
// 			"details": err.Error(),
// 		})
// 		return
// 	}

// 	if db.DB == nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
// 		return
// 	}

// 	// ---------- Call Stored Procedure (Single Row) ----------
// 	res, err := ExecSP(
// 		db.DB,
// 		"v1_PublicRole_AuthModule_GetSiteUsersAuthData",
// 		map[string]interface{}{
// 			"Username":    requestBody.Username,
// 			"AccountType": "Admin",
// 		},
// 		1, // 1 = single row
// 	)

// 	if err != nil {
// 		SendAuthError(c, 0)
// 		return
// 	}

// 	result, ok := AsSingleRow(res)
// 	if !ok {
// 		SendAuthError(c, 0)
// 		return
// 	}

// 	// ---------- Password Validation ----------
// 	passHash, ok := result["PasswordHash"].(string)
// 	if !ok {
// 		SendAuthError(c, 0)
// 		return
// 	}

// 	okPass, err := VerifyPassword(requestBody.Password, passHash)
// 	if err != nil || !okPass {
// 		SendAuthError(c, 0)
// 		return
// 	}

// 	// ---------- Extract User ID ----------
// 	id := 0
// 	if v, ok := result["SiteUsersId"].(int64); ok {
// 		id = int(v)
// 	} else {
// 		SendAuthError(c, 0)
// 		return
// 	}

// 	// ---------- TFA FLOW ----------
// 	if requestBody.TFAType != "" && requestBody.TFACode != "" {

// 		var secret string
// 		if v, ok := result["TotpSharedSecret"]; ok && v != nil {
// 			switch s := v.(type) {
// 			case string:
// 				secret = s
// 			case []byte:
// 				secret = string(s)
// 			default:
// 				SendAuthError(c, 2)
// 				return
// 			}
// 		} else {
// 			SendAuthError(c, 2)
// 			return
// 		}

// 		status := ValidateTfaCode(
// 			requestBody.TFAType,
// 			requestBody.TFACode,
// 			map[string]interface{}{
// 				"bTwoFactorAppAuthEnabled": result["bTwoFactorAppAuthEnabled"],
// 				"bTwoFactorSMSAuthEnabled": result["bTwoFactorSMSAuthEnabled"],
// 			},
// 			true,
// 			secret,
// 		)

// 		if status == TfaTypeInvalid {
// 			SendAuthError(c, 2)
// 			return
// 		}
// 		if status != Success {
// 			SendAuthError(c, 1)
// 			return
// 		}

// 		// ---------- Generate Tokens ----------
// 		firstName, _ := result["FirstName"].(string)
// 		lastName, _ := result["LastName"].(string)
// 		userCode, _ := result["UserCode"].(string)
// 		accountType := "Admin"

// 		accessToken, refreshToken, expiresIn, refreshTokenExpiresIn, err :=
// 			GenerateTokens(id, requestBody.RememberMe, firstName, lastName, accountType, userCode)

// 		if err != nil {
// 			SendAuthError(c, 0)
// 			return
// 		}

// 		details := PrepareUserDetails(
// 			result,
// 			accessToken,
// 			refreshToken,
// 			expiresIn,
// 			refreshTokenExpiresIn,
// 		)

// 		c.JSON(http.StatusOK, gin.H{
// 			"id":      id,
// 			"details": details,
// 			"status":  "1",
// 			"errors":  []string{},
// 		})
// 		return
// 	}

// }

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

func TfaLogin(c *gin.Context) {

	// ---------- Bind Request (IMPORTANT: correct struct tags) ----------
	var requestBody TfaLoginRequest
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		sendError(c, http.StatusBadRequest, "", "Invalid_JSON")
		return
	}

	// ---------- Required validations (.NET DefaultRequired) ----------
	if requestBody.Username == "" {
		sendError(c, http.StatusBadRequest, "Username", "Required")
		return
	}
	if requestBody.Password == "" {
		sendError(c, http.StatusBadRequest, "Password", "Required")
		return
	}
	if requestBody.TfaType == "" {
		sendError(c, http.StatusBadRequest, "TfaType", "Required")
		return
	}
	if requestBody.TfaCode == "" {
		sendError(c, http.StatusBadRequest, "TfaCode", "Required")
		return
	}
	if requestBody.TfaType != "SMS" && requestBody.TfaType != "AuthenticatorApp" {
		sendError(c, http.StatusBadRequest, "TfaType", "TfaType_Invalid")
		return
	}

	// ---------- Call SP ----------
	res, err := ExecSP(
		db.DB,
		"v1_PublicRole_AuthModule_GetSiteUsersAuthData",
		map[string]interface{}{
			"Username":    requestBody.Username,
			"AccountType": "Admin", // .NET fixed
		},
		1,
	)
	if err != nil {
		// .NET UserNotFound is 400 (NOT 401)
		sendUserPassError(c, http.StatusBadRequest, "Username_Or_Password_Incorrect")
		return
	}

	result, ok := AsSingleRow(res)
	if !ok || result == nil {
		sendUserPassError(c, http.StatusBadRequest, "Username_Or_Password_Incorrect")
		return
	}

	// ---------- Password Validation ----------
	passHash, _ := result["PasswordHash"].(string)
	if passHash == "" {
		sendUserPassError(c, http.StatusBadRequest, "Username_Or_Password_Incorrect")
		return
	}

	okPass, err := VerifyPassword(requestBody.Password, passHash)
	if err != nil || !okPass {
		// .NET invalid password => 401
		sendUserPassError(c, http.StatusUnauthorized, "Username_Or_Password_Incorrect")
		return
	}

	// ---------- Extract User ID ----------
	var id int
	if v, ok := result["SiteUsersId"].(int64); ok {
		id = int(v)
	} else {
		sendUserPassError(c, http.StatusBadRequest, "Username_Or_Password_Incorrect")
		return
	}

	// ---------- TFA VALIDATION (exact .NET behavior) ----------
	// AuthenticatorApp -> VerifyTotp(tfaCode, authData.TotpSharedSecret)
	// SMS -> expiry check + code compare

	status := Success

	if requestBody.TfaType == "AuthenticatorApp" {
		// .NET: if login and app not enabled => NoTfaEnabled -> treated as failure
		if b, ok := result["bTwoFactorAppAuthEnabled"].(bool); ok && !b {
			sendError(c, http.StatusBadRequest, "TfaCode", "Invalid")
			return
		}

		secret := ""
		if v, ok := result["TotpSharedSecret"]; ok && v != nil {
			switch s := v.(type) {
			case string:
				secret = s
			case []byte:
				secret = string(s)
			}
		}
		// IMPORTANT: secret empty => must fail
		if secret == "" {
			sendError(c, http.StatusBadRequest, "TfaCode", "Invalid")
			return
		}

		if !VerifyTotp(requestBody.TfaCode, secret) {
			sendError(c, http.StatusBadRequest, "TfaCode", "Invalid")
			return
		}
	}

	if requestBody.TfaType == "SMS" {
		if b, ok := result["bTwoFactorSMSAuthEnabled"].(bool); ok && !b {
			sendError(c, http.StatusBadRequest, "TfaCode", "Invalid")
			return
		}

		// NOTE: aapke current ValidateTfaCode me authData["TfaCodeExpiry"] pass hi nahi ho raha tha.
		// Yahan direct result se read kar rahe hain.
		exp, ok := result["TfaCodeExpiry"].(time.Time)
		if ok {
			if time.Now().UTC().After(exp.UTC()) {
				sendError(c, http.StatusBadRequest, "TfaCode", "Expired")
				return
			}
		}

		code, _ := result["TfaCode"].(string)
		if code == "" || code != requestBody.TfaCode {
			sendError(c, http.StatusBadRequest, "TfaCode", "Invalid")
			return
		}
	}

	_ = status

	// ---------- Generate tokens (TEMP: your existing GenerateTokens, next step we’ll make it .NET exact) ----------
	firstName, _ := result["FirstName"].(string)
	lastName, _ := result["LastName"].(string)
	userCode, _ := result["UserCode"].(string)
	// pkg.Log(result, "AllowedApiEndpointsCdl:", getString(result, "AllowedApiEndpointsCdl"))

	// accessToken, refreshToken, expiresIn, refreshTokenExpiresIn, err :=
	// 	GenerateTokens(id, requestBody.RememberMe, firstName, lastName, "Admin", userCode)
	allowedCdl := getString(result, "AllowedAPIEndpointsCDL")

	accessToken, refreshToken, expiresIn, refreshTokenExpiresIn, err :=
		GenerateTokens(id, requestBody.RememberMe, firstName, lastName, "Admin", userCode, allowedCdl)

	if err != nil || accessToken == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	details := PrepareUserDetails(result, accessToken, refreshToken, expiresIn, refreshTokenExpiresIn)

	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"details": details,
		"status":  "1",
		"errors":  []string{},
	})
}
