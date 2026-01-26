package auth

import (
	"cloud-web-phoenix-customer-v1-go/global"
	"net/http"
	"time"
)

func LoginHandler(d Deps) func(c GinCtx) {
	return func(c GinCtx) {

		var req CredentialsDto
		if err := c.ShouldBindJSON(&req); err != nil {
			sendLoginFail(c, "Username", "Password", "Required", http.StatusBadRequest)
			return
		}
		if req.Username == "" || req.Password == "" {
			sendLoginFail(c, "Username", "Password", "Required", http.StatusBadRequest)
			return
		}

		// 1) Get auth data
		res, err := d.ExecSP(d.DB, "v1_PublicRole_AuthModule_GetSiteUsersAuthData",
			map[string]interface{}{
				"Username":    req.Username,
				"AccountType": "Admin",
				"Domain":      nil, // .NET optional; agar aap chaho to origin host parse karke bhej do
			}, 1,
		)
		if err != nil {
			// .NET: UserNotFound -> invalid credentials
			sendLoginFail(c, "Username", "Password", "Username_Or_Password_Incorrect", http.StatusBadRequest)
			return
		}
		row, ok := d.AsSingleRow(res)
		if !ok {
			sendLoginFail(c, "Username", "Password", "Username_Or_Password_Incorrect", http.StatusBadRequest)
			return
		}

		userId := global.GetInt(row, "SiteUsersId")
		if userId <= 0 {
			sendLoginFail(c, "Username", "Password", "Username_Or_Password_Incorrect", http.StatusBadRequest)
			return
		}

		// 2) AccountDisabled (bSuppressed)
		if global.GetBool(row, "bSuppressed") {
			sendLoginFail(c, "Username", "Password", "Account_Disabled", http.StatusUnauthorized)
			return
		}

		// 3) AccountLocked by TryLoginCount
		tryCount := global.GetInt(row, "TryLoginCount")
		if d.MaxTryLoginCount > 0 && tryCount >= d.MaxTryLoginCount {
			sendLoginFail(c, "Username", "Password", "Account_Locked", http.StatusUnauthorized)
			return
		}

		// 4) Password verify
		passHash := global.GetString(row, "PasswordHash")
		okPass, verr := d.VerifyPassword(req.Password, passHash)
		if verr != nil || !okPass {

			// RecordFailedLogin (.NET same)
			newTry := tryCount + 1
			_, _ = d.ExecSP(d.DB, "v1_PublicRole_AuthModule_FailLogin",
				map[string]interface{}{
					"UserId":                  userId,
					"TryLoginCount":           newTry,
					"DateLastFailedLogin":     time.Now().UTC(),
					"DateLastSuccessfulLogin": global.GetTimePtr(row, "DateLastSuccessfulLogin"),
				}, 1,
			)

			// .NET: PasswordInvalid => unauthenticated
			sendLoginFail(c, "Username", "Password", "Username_Or_Password_Incorrect", http.StatusUnauthorized)
			return
		}

		// 5) Password ok: TFA decision
		bSms := global.GetBool(row, "bTwoFactorSMSAuthEnabled")
		bApp := global.GetBool(row, "bTwoFactorAppAuthEnabled")

		// 5A) SMS TFA enabled => generate code + expiry + SucceedTfaPhaseOne
		if bSms {
			code, e := generate6DigitCode()
			if e != nil {
				sendLoginFail(c, "Username", "Password", "Username_Or_Password_Incorrect", http.StatusUnauthorized)
				return
			}
			expMin := d.TfaCodeExpiryMinutes
			if expMin <= 0 {
				expMin = 10 // safe default
			}
			expiry := time.Now().UTC().Add(time.Minute * time.Duration(expMin))

			_, _ = d.ExecSP(d.DB, "v1_PublicRole_AuthModule_SucceedTfaPhaseOne",
				map[string]interface{}{
					"UserId":                  userId,
					"TryLoginCount":           0,
					"DateLastSuccessfulLogin": global.GetTimePtr(row, "DateLastSuccessfulLogin"),
					"DateLastFailedLogin":     nil,
					"TfaCode":                 code,
					"TfaCodeExpiry":           expiry,
				}, 1,
			)

			// Response without token (exact .NET)
			details := LoginSuccessDetails{
				BTwoFactorSMSAuthEnabled: bSms,
				BTwoFactorAppAuthEnabled: bApp,
				BEmailVerified:           global.GetBool(row, "bEmailVerified"),
			}
			sendLoginSuccess(c, userId, details)
			return
		}

		// 5B) App TFA enabled => success but NO token (exact .NET)
		if bApp {
			details := LoginSuccessDetails{
				BTwoFactorSMSAuthEnabled: bSms,
				BTwoFactorAppAuthEnabled: bApp,
				BEmailVerified:           global.GetBool(row, "bEmailVerified"),
			}
			sendLoginSuccess(c, userId, details)
			return
		}

		// 6) NO TFA => issue JWT + refresh + succeed login SP
		// NOTE: yahan aap apna GenerateTokens (JWT+Refresh) use karo (jo aap ne token match kar liya).
		// Also: scopes in response should be parsed from NavigationScopesJson (like .NET)
		accessToken, refreshToken, expiresIn, refreshTokenExpiresIn, err := GenerateTokens(userId, req.RememberMe, global.GetString(row, "FirstName"), global.GetString(row, "LastName"), global.GetString(row, "AccountType"), global.GetString(row, "UserCode"), global.GetString(row, "AllowedAPIEndpointsCdl"))
		if err != nil {
			sendLoginFail(c, "Username", "Password", "Username_Or_Password_Incorrect", http.StatusUnauthorized)
			return
		}

		// SucceedLogin SP (exact .NET)
		refreshExpiry := time.Now().UTC().Add(time.Second * time.Duration(refreshTokenExpiresIn))
		_, _ = d.ExecSP(d.DB, "v1_PublicRole_AuthModule_SucceedLogin",
			map[string]interface{}{
				"UserId":                  userId,
				"TryLoginCount":           0,
				"DateLastSuccessfulLogin": time.Now().UTC(),
				"DateLastFailedLogin":     nil,
				"RefreshToken":            refreshToken,
				"RefreshTokenExpiry":      refreshExpiry,
				"Browser":                 "", // aap parse kar lo
				"UserAgent":               c.GetHeader("User-Agent"),
				"IP":                      c.ClientIP(),
				"Location":                "", // aap geo service se fill kar lo
			}, 1,
		)

		token := accessToken
		rt := refreshToken
		details := LoginSuccessDetails{
			AccessToken:              &token,
			ExpiresIn:                &expiresIn,
			RefreshToken:             &rt,
			RefreshTokenExpiresIn:    &refreshTokenExpiresIn,
			BTwoFactorSMSAuthEnabled: bSms,
			BTwoFactorAppAuthEnabled: bApp,
			BEmailVerified:           global.GetBool(row, "bEmailVerified"),
			Scopes:                   parseScopesJson(row["NavigationScopesJson"]),
		}
		sendLoginSuccess(c, userId, details)
	}
}
