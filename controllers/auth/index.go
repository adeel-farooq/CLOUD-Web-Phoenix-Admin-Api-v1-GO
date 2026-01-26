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

	// ---------- Bind ----------
	var req TfaLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeApiResult(c, ObjectResult[any]{
			Details: nil,
			Errors: []ErrorResult{
				{ErrorType: ErrorBadRequest, FieldName: "", MessageCode: "Invalid_JSON"},
			},
		})
		return
	}

	// ---------- Validate (match .NET DefaultRequired + validator) ----------
	if errs := validateTfaLoginRequest(&req); len(errs) > 0 {
		writeApiResult(c, ObjectResult[any]{Details: nil, Errors: errs})
		return
	}

	// ---------- Load auth settings (match .NET) ----------
	authSettings := loadAuthSettings()

	// ---------- Step 1: SP GetSiteUsersAuthData ----------
	res, err := ExecSP(
		db.DB,
		"v1_PublicRole_AuthModule_GetSiteUsersAuthData",
		map[string]any{
			"Username":    req.Username,
			"AccountType": "Admin", // .NET fixed
			// Domain not passed in .NET for Admin -> keep omitted
		},
		1,
	)
	if err != nil {
		// .NET would transform dbResult normally; keep safe generic:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	row, ok := AsSingleRow(res)
	if !ok || row == nil {
		writeApiResult(c, ObjectResult[LoginSuccessDetails]{Details: LoginSuccessDetails{}, Errors: errUserNotFound()})
		return
	}

	authData, mapErr := mapRowToAuthData(row)
	if mapErr != nil || authData.SiteUsersId <= 0 {
		writeApiResult(c, ObjectResult[LoginSuccessDetails]{Details: LoginSuccessDetails{}, Errors: errUserNotFound()})
		return
	}

	// ---------- Step 2: Account lock check (exact .NET) ----------
	if isAccountLocked(authSettings, authData.TryLoginCount, authData.DateLastFailedLogin) {
		writeApiResult(c, ObjectResult[LoginSuccessDetails]{Details: LoginSuccessDetails{}, Errors: errAccountLocked()})
		return
	}

	// ---------- Step 3: Password verify ----------
	okPass, err := VerifyPassword(req.Password, authData.PasswordHash)
	if err != nil || !okPass {
		// record failed login SP (like .NET)
		_, _ = ExecSP(db.DB, "v2_PublicRole_AuthModule_FailLogin", buildFailLoginParams(authData, c), 0)

		writeApiResult(c, ObjectResult[LoginSuccessDetails]{Details: LoginSuccessDetails{}, Errors: errInvalidPassword()})
		return
	}

	// ---------- Step 4: Email not verified => success WITHOUT token ----------
	// .NET: if !bEmailVerified => success with flags only
	if !authData.BEmailVerified {
		details := LoginSuccessDetails{
			AccessToken:              nil,
			ExpiresIn:                0,
			RefreshToken:             nil,
			RefreshTokenExpiresIn:    0,
			BTwoFactorAppAuthEnabled: authData.BTwoFactorAppAuthEnabled,
			BTwoFactorSMSAuthEnabled: authData.BTwoFactorSMSAuthEnabled,
			BEmailVerified:           authData.BEmailVerified,
		}

		writeApiResult(c, ObjectResult[LoginSuccessDetails]{Id: authData.SiteUsersId, Details: details, Errors: []ErrorResult{}})
		return
	}

	// ---------- Step 5: Validate TFA (match .NET ValidationHelper.ValidateTfaCodeAsync behavior) ----------
	now := time.Now().UTC()

	if req.TfaType == "AuthenticatorApp" {
		// if disabled => treat as "NoTfaEnabled" in service flow, but in tfalogin command they are forcing required.
		// In your .NET flow, if user chooses AuthenticatorApp but user doesn't have it enabled, VerifyTotp won't pass => invalid.
		// However ValidationHelper has a TfaType_Disabled check when called in other flows, not in TfaLoginAsync (it uses ValidateTfaLoginAttempt).
		// Here, service uses ValidateTfaLoginAttempt which checks enabled.
		if !authData.BTwoFactorAppAuthEnabled {
			// result => BadRequest TfaCode_Invalid (because service returns NoTfaEnabled => treated as invalid for this login attempt)
			_, _ = ExecSP(db.DB, "v2_PublicRole_AuthModule_FailLogin", buildFailLoginParams(authData, c), 0)
			writeApiResult(c, ObjectResult[LoginSuccessDetails]{Details: LoginSuccessDetails{}, Errors: errTfaInvalid()})
			return
		}

		if authData.TotpSharedSecret == nil || *authData.TotpSharedSecret == "" {
			_, _ = ExecSP(db.DB, "v2_PublicRole_AuthModule_FailLogin", buildFailLoginParams(authData, c), 0)
			writeApiResult(c, ObjectResult[LoginSuccessDetails]{Details: LoginSuccessDetails{}, Errors: errTfaInvalid()})
			return
		}

		if !verifyTotp(req.TfaCode, *authData.TotpSharedSecret, authSettings.TotpAllowedPreviousEpochs, authSettings.TotpAllowedFutureEpochs, now) {
			_, _ = ExecSP(db.DB, "v2_PublicRole_AuthModule_FailLogin", buildFailLoginParams(authData, c), 0)
			writeApiResult(c, ObjectResult[LoginSuccessDetails]{Details: LoginSuccessDetails{}, Errors: errTfaInvalid()})
			return
		}
	}

	if req.TfaType == "SMS" {
		if !authData.BTwoFactorSMSAuthEnabled {
			_, _ = ExecSP(db.DB, "v2_PublicRole_AuthModule_FailLogin", buildFailLoginParams(authData, c), 0)
			writeApiResult(c, ObjectResult[LoginSuccessDetails]{Details: LoginSuccessDetails{}, Errors: errTfaInvalid()})
			return
		}

		if authData.TfaCodeExpiry != nil && authData.TfaCodeExpiry.UTC().Before(now) {
			_, _ = ExecSP(db.DB, "v2_PublicRole_AuthModule_FailLogin", buildFailLoginParams(authData, c), 0)
			writeApiResult(c, ObjectResult[LoginSuccessDetails]{Details: LoginSuccessDetails{}, Errors: errTfaExpired()})
			return
		}

		if authData.TfaCode == nil || *authData.TfaCode != req.TfaCode {
			_, _ = ExecSP(db.DB, "v2_PublicRole_AuthModule_FailLogin", buildFailLoginParams(authData, c), 0)
			writeApiResult(c, ObjectResult[LoginSuccessDetails]{Details: LoginSuccessDetails{}, Errors: errTfaInvalid()})
			return
		}
	}

	// ---------- Step 6: SUCCESS => Generate tokens (match .NET TokenHelper) ----------
	firstName, lastName, userCode := "", "", ""
	if authData.FirstName != nil {
		firstName = *authData.FirstName
	}
	if authData.LastName != nil {
		lastName = *authData.LastName
	}
	if authData.UserCode != nil {
		userCode = *authData.UserCode
	}

	allowedCdl := ""
	if authData.AllowedApiEndpointsCdl != nil {
		allowedCdl = *authData.AllowedApiEndpointsCdl
	}

	accessToken, expiresIn, refreshToken, refreshExpiresIn, tokErr :=
		generateJwtAndRefresh(authSettings, authData.SiteUsersId, req.RememberMe, firstName, lastName, userCode, allowedCdl)

	if tokErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	// ---------- Step 7: SP SucceedLogin (IMPORTANT: no RefreshTokenExpiry passed) ----------
	_, _ = ExecSP(
		db.DB,
		"v2_PublicRole_AuthModule_SucceedLogin",
		buildSucceedLoginParams(authData, refreshToken, c),
		0,
	)

	// ---------- Step 8: Response details (match .NET LoginSuccessResultDetails) ----------
	details := LoginSuccessDetails{
		AccessToken:              accessToken,
		ExpiresIn:                expiresIn,
		RefreshToken:             refreshToken,
		RefreshTokenExpiresIn:    refreshExpiresIn,
		BTwoFactorAppAuthEnabled: authData.BTwoFactorAppAuthEnabled,
		BTwoFactorSMSAuthEnabled: authData.BTwoFactorSMSAuthEnabled,
		BEmailVerified:           authData.BEmailVerified,
		Scopes:                   parseScopesJson(authData.NavigationScopesJson),
	}

	writeApiResult(c, ObjectResult[LoginSuccessDetails]{
		Id:      authData.SiteUsersId,
		Details: details,
		Errors:  []ErrorResult{},
	})
}
