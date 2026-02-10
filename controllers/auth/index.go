package auth

import (
	"cloud-web-phoenix-customer-v1-go/db"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
)

func SignIn(c *gin.Context) {

	// ---------- Bind Request ----------
	var requestBody SignInRequest
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		sendError(c, http.StatusBadRequest, "Body", "Invalid")
		return
	}

	// ---------- Required validation (.NET style) ----------
	if strings.TrimSpace(requestBody.Username) == "" {
		sendError(c, http.StatusBadRequest, "Username", "Required")
		return
	}
	if strings.TrimSpace(requestBody.Password) == "" {
		sendError(c, http.StatusBadRequest, "Password", "Required")
		return
	}

	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
		return
	}

	// ---------- Step-1: GetSiteUsersAuthData (Single Row) ----------
	res, err := ExecSP(
		db.DB,
		"v1_PublicRole_AuthModule_GetSiteUsersAuthData",
		map[string]interface{}{
			"Username":    requestBody.Username,
			"AccountType": "Admin",
			"Domain":      nil,
		},
		1, // 1 = single row
	)

	if err != nil {
		sendUserPassError(c, http.StatusUnauthorized, "Username_Or_Password_Incorrect")
		return
	}

	result, ok := AsSingleRow(res)
	if !ok {
		sendUserPassError(c, http.StatusUnauthorized, "Username_Or_Password_Incorrect")
		return
	}

	// ---------- Extract User ID ----------
	siteUsersId := getInt64(result, "SiteUsersId")
	if siteUsersId <= 0 {
		sendUserPassError(c, http.StatusUnauthorized, "Username_Or_Password_Incorrect")
		return
	}

	// ---------- Step-2: ValidateLoginAttempt rules (.NET order) ----------
	// 2A) AccountDisabled (bSuppressed)
	if getBoolAny(result, []string{"bSuppressed", "BSuppressed"}) {
		sendUserPassError(c, http.StatusUnauthorized, "Account_Disabled")
		return
	}

	// 2B) AccountLocked (TryLoginCount >= MaxTryLoginCount)
	tryCount := getInt64(result, "TryLoginCount")
	maxTry := getIntEnvAny([]string{"TRY_LOGIN_COUNTER_MAX"}, 5)
	locked := maxTry > 0 && int(tryCount) >= maxTry

	// 2C) Password verify
	passHash, _ := result["PasswordHash"].(string)
	if strings.TrimSpace(passHash) == "" {
		sendUserPassError(c, http.StatusUnauthorized, "Username_Or_Password_Incorrect")
		return
	}

	okPass, err := VerifyPassword(requestBody.Password, passHash)

	// 2D) Locked handling: verify password; if wrong => RecordFailedLogin; always return Account_Locked
	if locked {
		if err != nil || !okPass {
			now := time.Now().UTC()
			var lastSuccess interface{} = nil
			if t, ok := getTimeAny(result, []string{"DateLastSuccessfulLogin"}); ok {
				lastSuccess = t.UTC().Truncate(time.Millisecond)
			}
			V1FailLogin(db.DB, siteUsersId, tryCount+1, lastSuccess, now)
		}
		sendUserPassError(c, http.StatusUnauthorized, "Account_Locked")
		return
	}

	// 2E) Password invalid => RecordFailedLogin + invalid creds
	if err != nil || !okPass {
		now := time.Now().UTC()
		var lastSuccess interface{} = nil
		if t, ok := getTimeAny(result, []string{"DateLastSuccessfulLogin"}); ok {
			lastSuccess = t.UTC().Truncate(time.Millisecond)
		}
		V1FailLogin(db.DB, siteUsersId, tryCount+1, lastSuccess, now)
		sendUserPassError(c, http.StatusUnauthorized, "Username_Or_Password_Incorrect")
		return
	}

	// ---------- Step-3: Status handling (TFA vs Success) ----------
	bSms := getBoolAny(result, []string{"bTwoFactorSMSAuthEnabled"})
	bApp := getBoolAny(result, []string{"bTwoFactorAppAuthEnabled"})

	// SMSTfaEnabled => generate code + expiry + SucceedTfaPhaseOne, return success without token
	if bSms {
		code, genErr := generate6DigitCode()
		if genErr != nil {
			sendUserPassError(c, http.StatusUnauthorized, "Username_Or_Password_Incorrect")
			return
		}
		expMin := getIntEnvAny([]string{"TFA_CODE_EXPIRY_MINUTES", "TFA_CODE_EXPIRE_MINUTES"}, 10)
		expiry := time.Now().UTC().Add(time.Minute * time.Duration(expMin))

		var lastSuccess interface{} = nil
		if t, ok := getTimeAny(result, []string{"DateLastSuccessfulLogin"}); ok {
			lastSuccess = t.UTC().Truncate(time.Millisecond)
		}
		V1SucceedTfaPhaseOne(db.DB, siteUsersId, lastSuccess, code, expiry)

		details := PrepareUserDetails(result, nil, nil, 0, 0)
		c.JSON(http.StatusOK, gin.H{
			"id":      int(siteUsersId),
			"details": details,
			"status":  "1",
			"errors":  []gin.H{},
		})
		return
	}

	// AppTfaEnabled => return success without token
	if bApp {
		details := PrepareUserDetails(result, nil, nil, 0, 0)
		c.JSON(http.StatusOK, gin.H{
			"id":      int(siteUsersId),
			"details": details,
			"status":  "1",
			"errors":  []gin.H{},
		})
		return
	}

	// Success (no TFA) => issue JWT + refresh, SucceedLogin
	firstName := getStringAny(result, []string{"FirstName"})
	lastName := getStringAny(result, []string{"LastName"})
	userCode := getStringAny(result, []string{"UserCode"})
	allowedCdl := getStringAny(result, []string{"AllowedAPIEndpointsCDL", "AllowedAPIEndpointsCDL", "AllowedAPIEndpointsCDL"})

	accessToken, refreshToken, expiresIn, refreshTokenExpiresIn, err :=
		GenerateTokens(int(siteUsersId), requestBody.RememberMe, firstName, lastName, "Admin", userCode, allowedCdl)
	if err != nil || accessToken == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	now := time.Now().UTC()
	ip := clientIPSafe(c)
	ua := userAgentSafe(c)
	browserSummary := getBrowserSummaryDotNetStyle(ua)
	location := getLocationSafe(ip)
	refreshExpiry := now.Add(time.Second * time.Duration(refreshTokenExpiresIn))
	V1SucceedLogin(db.DB, siteUsersId, refreshToken, refreshExpiry, browserSummary, ua, ip, location, now)

	details := PrepareUserDetails(result, accessToken, refreshToken, expiresIn, refreshTokenExpiresIn)
	// Ensure flags match the decision we just made
	details["bTwoFactorSMSAuthEnabled"] = bSms
	details["bTwoFactorAppAuthEnabled"] = bApp

	c.JSON(http.StatusOK, gin.H{
		"id":      int(siteUsersId),
		"details": details,
		"status":  "1",
		"errors":  []gin.H{},
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
		sendUserPassError(c, http.StatusUnauthorized, "Username_Or_Password_Incorrect")
		return
	}

	result, ok := AsSingleRow(res)
	if !ok || result == nil {
		sendUserPassError(c, http.StatusUnauthorized, "Username_Or_Password_Incorrect")
		return
	}

	// ---------- Extract User ID ----------
	siteUsersId := getInt64(result, "SiteUsersId")
	if siteUsersId <= 0 {
		sendUserPassError(c, http.StatusUnauthorized, "Username_Or_Password_Incorrect")
		return
	}

	// ---------- ValidateLoginAttempt rules (.NET order) ----------
	if getBoolAny(result, []string{"bSuppressed", "BSuppressed"}) {
		sendUserPassError(c, http.StatusUnauthorized, "Account_Disabled")
		return
	}
	tryCount := getInt64(result, "TryLoginCount")
	maxTry := getIntEnvAny([]string{"TRY_LOGIN_COUNTER_MAX"}, 5)
	locked := maxTry > 0 && int(tryCount) >= maxTry

	// ---------- Password Validation ----------
	passHash, _ := result["PasswordHash"].(string)
	if passHash == "" {
		sendUserPassError(c, http.StatusUnauthorized, "Username_Or_Password_Incorrect")
		return
	}

	okPass, err := VerifyPassword(requestBody.Password, passHash)
	if locked {
		if err != nil || !okPass {
			now := time.Now().UTC()
			var lastSuccess interface{} = nil
			if t, ok := getTimeAny(result, []string{"DateLastSuccessfulLogin"}); ok {
				lastSuccess = t.UTC().Truncate(time.Millisecond)
			}
			V1FailLogin(db.DB, siteUsersId, tryCount+1, lastSuccess, now)
		}
		sendUserPassError(c, http.StatusUnauthorized, "Account_Locked")
		return
	}
	if err != nil || !okPass {
		now := time.Now().UTC()
		var lastSuccess interface{} = nil
		if t, ok := getTimeAny(result, []string{"DateLastSuccessfulLogin"}); ok {
			lastSuccess = t.UTC().Truncate(time.Millisecond)
		}
		V1FailLogin(db.DB, siteUsersId, tryCount+1, lastSuccess, now)
		sendUserPassError(c, http.StatusUnauthorized, "Username_Or_Password_Incorrect")
		return
	}

	id := int(siteUsersId)

	// ---------- TFA VALIDATION (exact .NET behavior) ----------
	// AuthenticatorApp -> VerifyTotp(tfaCode, authData.TotpSharedSecret)
	// SMS -> expiry check + code compare

	status := Success

	if requestBody.TfaType == "AuthenticatorApp" {
		// .NET: if login and app not enabled => NoTfaEnabled -> treated as failure
		if b, ok := result["bTwoFactorAppAuthEnabled"].(bool); ok && !b {
			now := time.Now().UTC()
			tryCount := getInt64(result, "TryLoginCount")
			var lastSuccess interface{} = nil
			if t, ok := getTimeAny(result, []string{"DateLastSuccessfulLogin"}); ok {
				lastSuccess = t.UTC().Truncate(time.Millisecond)
			}
			V1FailLogin(db.DB, siteUsersId, tryCount+1, lastSuccess, now)
			sendError(c, http.StatusBadRequest, "TfaCode", "No_Tfa_Enabled")
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
			now := time.Now().UTC()
			tryCount := getInt64(result, "TryLoginCount")
			var lastSuccess interface{} = nil
			if t, ok := getTimeAny(result, []string{"DateLastSuccessfulLogin"}); ok {
				lastSuccess = t.UTC().Truncate(time.Millisecond)
			}
			V1FailLogin(db.DB, siteUsersId, tryCount+1, lastSuccess, now)
			sendError(c, http.StatusBadRequest, "TfaCode", "Invalid")
			return
		}

		if !VerifyTotpWithSkew(requestBody.TfaCode, secret,
			getIntEnvAny([]string{"TOTP_ALLOWED_PREVIOUS_EPOCHS"}, 1),
			getIntEnvAny([]string{"TOTP_ALLOWED_FUTURE_EPOCHS"}, 1),
		) {
			now := time.Now().UTC()
			tryCount := getInt64(result, "TryLoginCount")
			var lastSuccess interface{} = nil
			if t, ok := getTimeAny(result, []string{"DateLastSuccessfulLogin"}); ok {
				lastSuccess = t.UTC().Truncate(time.Millisecond)
			}
			V1FailLogin(db.DB, siteUsersId, tryCount+1, lastSuccess, now)
			sendError(c, http.StatusBadRequest, "TfaCode", "Invalid")
			return
		}
	}

	if requestBody.TfaType == "SMS" {
		if b, ok := result["bTwoFactorSMSAuthEnabled"].(bool); ok && !b {
			now := time.Now().UTC()
			tryCount := getInt64(result, "TryLoginCount")
			var lastSuccess interface{} = nil
			if t, ok := getTimeAny(result, []string{"DateLastSuccessfulLogin"}); ok {
				lastSuccess = t.UTC().Truncate(time.Millisecond)
			}
			V1FailLogin(db.DB, siteUsersId, tryCount+1, lastSuccess, now)
			sendError(c, http.StatusBadRequest, "TfaCode", "No_Tfa_Enabled")
			return
		}

		// NOTE: aapke current ValidateTfaCode me authData["TfaCodeExpiry"] pass hi nahi ho raha tha.
		// Yahan direct result se read kar rahe hain.
		if exp, ok := getTimeAny(result, []string{"TfaCodeExpiry"}); !ok || time.Now().UTC().After(exp.UTC()) {
			now := time.Now().UTC()
			tryCount := getInt64(result, "TryLoginCount")
			var lastSuccess interface{} = nil
			if t, ok := getTimeAny(result, []string{"DateLastSuccessfulLogin"}); ok {
				lastSuccess = t.UTC().Truncate(time.Millisecond)
			}
			V1FailLogin(db.DB, siteUsersId, tryCount+1, lastSuccess, now)
			sendError(c, http.StatusBadRequest, "TfaCode", "Expired")
			return
		}

		code, _ := result["TfaCode"].(string)
		if code == "" || code != requestBody.TfaCode {
			now := time.Now().UTC()
			tryCount := getInt64(result, "TryLoginCount")
			var lastSuccess interface{} = nil
			if t, ok := getTimeAny(result, []string{"DateLastSuccessfulLogin"}); ok {
				lastSuccess = t.UTC().Truncate(time.Millisecond)
			}
			V1FailLogin(db.DB, siteUsersId, tryCount+1, lastSuccess, now)
			sendError(c, http.StatusBadRequest, "TfaCode", "Invalid")
			return
		}
	}
	_ = status

	// ---------- Generate tokens (TEMP: your existing GenerateTokens, next step we’ll make it .NET exact) ----------
	firstName, _ := result["FirstName"].(string)
	lastName, _ := result["LastName"].(string)
	userCode, _ := result["UserCode"].(string)
	// v1 SucceedLogin sets DateLastSuccessfulLogin=now and TryLoginCount=0
	// pkg.Log(result, "AllowedAPIEndpointsCDL:", getString(result, "AllowedAPIEndpointsCDL"))

	// accessToken, refreshToken, expiresIn, refreshTokenExpiresIn, err :=
	// 	GenerateTokens(id, requestBody.RememberMe, firstName, lastName, "Admin", userCode)
	allowedCdl := getStringAny(result, []string{"AllowedAPIEndpointsCDL", "AllowedAPIEndpointsCDL", "AllowedAPIEndpointsCDL"})

	accessToken, refreshToken, expiresIn, refreshTokenExpiresIn, err :=
		GenerateTokens(id, requestBody.RememberMe, firstName, lastName, "Admin", userCode, allowedCdl)

	if err != nil || accessToken == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	now := time.Now().UTC()
	ip := clientIPSafe(c)
	ua := userAgentSafe(c)
	browserSummary := getBrowserSummaryDotNetStyle(ua)
	location := getLocationSafe(ip)
	refreshExpiry := now.Add(time.Second * time.Duration(refreshTokenExpiresIn))
	V1SucceedLogin(db.DB, siteUsersId, refreshToken, refreshExpiry, browserSummary, ua, ip, location, now)

	details := PrepareUserDetails(result, accessToken, refreshToken, expiresIn, refreshTokenExpiresIn)

	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"details": details,
		"status":  "1",
		"errors":  []gin.H{},
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
	if db.DB == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
		return
	}

	// ---------- Call Stored Procedure (Single Row) ----------
	res, err := ExecSP(
		db.DB,
		"v1_AdminRole_ProfileModule_GetUserInfo",
		map[string]interface{}{
			"SiteUsersId": user["id"],
		},
		1, // 1 = single row
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user info"})
		return
	}

	result, ok := AsSingleRow(res)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user info"})
		return
	}

	// The SP returns a wrapper like: { Details: "{...json...}", Status: "1", Errors: "", Id: 0 }
	// Frontend expects: { id:0, details:{id, firstName, lastName, accountType}, status:"1", errors:[] }
	var rawDetails interface{}
	if v, exists := result["Details"]; exists {
		rawDetails = v
	}

	var detailsObj map[string]interface{}
	switch v := rawDetails.(type) {
	case map[string]interface{}:
		detailsObj = v
	case string:
		if strings.TrimSpace(v) != "" {
			_ = json.Unmarshal([]byte(v), &detailsObj)
		}
	case []byte:
		if len(v) > 0 {
			_ = json.Unmarshal(v, &detailsObj)
		}
	}
	if detailsObj == nil {
		detailsObj = map[string]interface{}{}
	}

	getDetail := func(keys ...string) interface{} {
		for _, k := range keys {
			if val, ok := detailsObj[k]; ok && val != nil {
				return val
			}
		}
		return nil
	}
	toInt := func(v interface{}) int {
		switch t := v.(type) {
		case int:
			return t
		case int64:
			return int(t)
		case float64:
			return int(t)
		case string:
			n, _ := strconv.Atoi(strings.TrimSpace(t))
			return n
		default:
			return 0
		}
	}
	toString := func(v interface{}) string {
		switch t := v.(type) {
		case string:
			return t
		case []byte:
			return string(t)
		default:
			return ""
		}
	}

	clean := gin.H{
		"id":          toInt(getDetail("Id", "id")),
		"firstName":   toString(getDetail("FirstName", "firstName")),
		"lastName":    toString(getDetail("LastName", "lastName")),
		"accountType": toString(getDetail("AccountType", "accountType")),
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      0,
		"details": clean,
		"status":  "1",
		"errors":  []gin.H{},
	})

}
