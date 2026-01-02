package auth

import (
	"crypto/sha1" // PBKDF2 default in .NET Rfc2898DeriveBytes
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/pbkdf2"
)

const KeySize = 32 // 256-bit

// VerifyPassword compares a plain password with the saved hash string
func VerifyPassword(password string, savedHash string) (bool, error) {
	fmt.Println("Verifying password...", savedHash, "normal", password)
	// Split "iterations.salt.key"
	parts := strings.Split(savedHash, ".")
	if len(parts) != 3 {
		return false, fmt.Errorf("invalid hash format")
	}

	// Parse iterations
	iterations, err := strconv.Atoi(parts[0])
	if err != nil {
		return false, err
	}

	// Decode salt
	salt, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, err
	}

	// Decode original hash
	origHash, err := base64.StdEncoding.DecodeString(parts[2])
	if err != nil {
		return false, err
	}

	// Derive key from password using PBKDF2 with SHA1 (same as .NET Rfc2898DeriveBytes default)
	testHash := pbkdf2.Key([]byte(password), salt, iterations, KeySize, sha1.New)

	// Compare
	if string(testHash) == string(origHash) {
		return true, nil
	}
	return false, nil
}

// ValidateTfaCode checks OTP validity
func ValidateTfaCode(tfaType string, tfaCode string, authData map[string]interface{}, bLogin bool, totpSharedSecret string) int {

	if tfaType == "AuthenticatorApp" {
		if bLogin && !authData["bTwoFactorAppAuthEnabled"].(bool) {
			return NoTfaEnabled
		}
		secret := ""
		if totpSharedSecret != "" {
			secret = totpSharedSecret
		}
		if VerifyTotp(tfaCode, secret) {
			return Success
		}
		return TfaCodeInvalid
	}

	if tfaType == "SMS" {
		if bLogin && !authData["bTwoFactorSMSAuthEnabled"].(bool) {
			return NoTfaEnabled
		}
		if time.Now().UTC().After(authData["TfaCodeExpiry"].(time.Time)) {
			return TfaCodeExpired
		}
		if authData["TfaCode"].(string) == tfaCode {
			return Success
		}
		return TfaCodeInvalid
	} else if tfaType != "SMS" || tfaType != "AuthenticatorApp" {
		return TfaTypeInvalid

	}

	return TfaCodeInvalid
}

// VerifyTotp validates TOTP code using secret
func VerifyTotp(totpCode string, secretKey string) bool {
	if secretKey == "" || totpCode == "" {
		return false
	}
	// Remove spaces from the code
	cleanCode := strings.ReplaceAll(totpCode, " ", "")
	opts := totp.ValidateOpts{
		Period:    30,
		Skew:      1, // allow 1 step before/after
		Digits:    6,
		Algorithm: otp.AlgorithmSHA1,
	}
	// Validate returns (bool, error), but we only care about bool
	valid, _ := totp.ValidateCustom(cleanCode, secretKey, time.Now().UTC(), opts)
	return valid
}
func PrepareUserDetails(result map[string]interface{}, accessToken interface{}, refreshToken interface{}, expiresIn int, refreshTokenExpiresIn int) gin.H {
	details := gin.H{
		"accessToken":              accessToken,
		"expiresIn":                expiresIn,
		"refreshToken":             refreshToken,
		"refreshTokenExpiresIn":    refreshTokenExpiresIn,
		"bTwoFactorAppAuthEnabled": result["bTwoFactorAppAuthEnabled"],
		"bTwoFactorSMSAuthEnabled": result["bTwoFactorSMSAuthEnabled"],
		"bEmailVerified":           result["bEmailVerified"],
		// "bSuppressed":              result["bSuppressed"],
		// "accountType":              result["AccountType"],
	}
	if scopes, ok := result["NavigationScopesJson"]; ok && scopes != nil {
		// fmt.Println("Parsing scopes:", scopes)
		// Try to parse as JSON array of objects
		var parsedScopes []map[string]interface{}
		switch v := scopes.(type) {
		case string:
			if v != "" {
				if err := json.Unmarshal([]byte(v), &parsedScopes); err == nil {
					if len(parsedScopes) > 0 {
						details["scopes"] = parsedScopes
					}
				}
			}
		case []map[string]interface{}:
			if len(v) > 0 {
				details["scopes"] = v
			}
		}
	}
	return details
}
func SendAuthError(c *gin.Context, errorType int) {
	errorMsg := []gin.H{}
	if errorType == 0 {

		errorMsg = []gin.H{
			{"fieldName": "Username", "messageCode": "Username_Or_Password_Incorrect"},
			{"fieldName": "Password", "messageCode": "Username_Or_Password_Incorrect"},
		}
	} else if errorType == 1 {
		errorMsg = []gin.H{
			{"fieldName": "TfaCode", "messageCode": "Invalid"},
		}
	} else if errorType == 2 {
		errorMsg = []gin.H{
			{"fieldName": "TfaType", "messageCode": "TfaType_Invalid"},
		}

	}
	c.JSON(http.StatusUnauthorized, gin.H{
		"id":      0,
		"details": nil,
		"status":  "0",
		"errors":  errorMsg,
	})
}
func GenerateTokens(userID int, rememberMe bool, firstName string, lastName string, accountType string) (string, string, int, int, error) {
	now := time.Now().UTC()

	// Get JWT secret and expiry from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your_default_jwt_secret_here"
	}
	jwtExpiryStr := os.Getenv("JWT_EXPIRY_MINUTES")
	jwtExpiryMinutes := 60
	if jwtExpiryStr != "" {
		if v, err := strconv.Atoi(jwtExpiryStr); err == nil {
			jwtExpiryMinutes = v
		}
	}
	jwtExpirySeconds := jwtExpiryMinutes * 60

	claims := jwt.MapClaims{
		"sub":         strconv.Itoa(userID),
		"exp":         now.Add(time.Minute * time.Duration(jwtExpiryMinutes)).Unix(),
		"RememberMe":  rememberMe,
		"id":          userID,
		"firstName":   firstName,
		"lastName":    lastName,
		"accountType": accountType,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", 0, 0, err
	}

	refreshToken := uuid.NewString()
	refreshExpiry := now.Add(time.Hour * 24 * 30) // 30 days
	refreshTokenExpiresIn := int(refreshExpiry.Sub(now).Seconds())

	return accessToken, refreshToken, jwtExpirySeconds, refreshTokenExpiresIn, nil
}

// ExtractUserID extracts user ID from JWT claims
func ExtractUser(token *jwt.Token) map[string]interface{} {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}

	user := make(map[string]interface{})

	if idVal, ok := claims["id"]; ok {
		switch v := idVal.(type) {
		case float64:
			user["id"] = int(v)
		case int:
			user["id"] = v
		case string:
			if idInt, err := strconv.Atoi(v); err == nil {
				user["id"] = idInt
			}
		}
	}
	if firstName, ok := claims["firstName"].(string); ok {
		user["firstName"] = firstName
	}
	if lastName, ok := claims["lastName"].(string); ok {
		user["lastName"] = lastName
	}
	if accountType, ok := claims["accountType"].(string); ok {
		user["accountType"] = accountType
	}

	if len(user) == 0 {
		return nil
	}
	return user
}

// JsonSuccessResponse sends a standard success response
func JsonSuccessResponse(c *gin.Context, id int, details interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"details": details,
		"status":  "1",
		"errors":  []string{},
	})
}
