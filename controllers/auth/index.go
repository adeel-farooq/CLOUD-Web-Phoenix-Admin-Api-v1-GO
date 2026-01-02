package auth

import (
	"cloud-web-phoenix-customer-v1-go/db"
	"cloud-web-phoenix-customer-v1-go/pkg"
	"fmt"
	"net/http"

	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func SignIn(c *gin.Context) {
	var requestBody SignInRequest
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		// Try to extract detailed validation errors
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

	// Define struct for user data

	var user UserResponse

	// Map AccountType int to string for proc
	accountTypeStr := "Customer"
	if requestBody.AccountType == 0 {
		accountTypeStr = "Admin"
	}
	query := "exec v1_PublicRole_AuthModule_GetSiteUsersAuthData @Username=@Username, @AccountType=@AccountType"
	row := db.DB.QueryRow(query,
		sql.Named("Username", requestBody.Username),
		sql.Named("AccountType", accountTypeStr),
	)
	pkg.Log("[DB DEBUG] Would run:", fmt.Sprintf(
		"exec v1_PublicRole_AuthModule_GetSiteUsersAuthData @Username='%s',@AccountType='%s'",
		requestBody.Username, accountTypeStr,
	))
	// Scan only the fields you need (update UserResponse struct as needed)
	err := row.Scan(
		&user.SiteUsersId,
		&user.CustomersId,
		&user.CustomerUsersCustomersId,
		&user.CustomerAccountType,
		&user.TryLoginCount,
		&user.PasswordHash,
		&user.DateLastFailedLogin,
		&user.DateLastSuccessfulLogin,
		&user.BTwoFactorAppAuthEnabled,
		&user.BTwoFactorSMSAuthEnabled,
		&user.BEmailVerified,
		&user.EmailVerificationCode,
		&user.EmailVerificationCodeExpiry,
		&user.PhoneNumber,
		&user.SiteName,
		&user.EmailAddress,
		&user.CultureInfo,
		&user.BFrozen,
		&user.BSuppressed,
		&user.TfaCode,
		&user.TfaCodeExpiry,
		&user.TotpSharedSecret,
		&user.BDocumentVerified,
		&user.FirstName,
		&user.LastName,
		&user.UserCode,
		&user.AllowedAPIEndpointsCDL,
		&user.TwoFactorSMSRequestCounter,
		&user.DateLastTwoFactorSMSRequested,
	)
	if err != nil {
		pkg.Log("[DB SCAN ERROR]", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// ✅ Check password using VerifyPassword
	ok, err := VerifyPassword(requestBody.Password, user.PasswordHash)
	if err != nil || !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Success response (add token logic if needed)
	c.JSON(http.StatusOK, gin.H{
		"message": "Sign in successful",
		"user": gin.H{
			"id":       user.SiteUsersId,
			"username": requestBody.Username,
		},
	})
}
