package auth

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/sha1" // PBKDF2 default in .NET Rfc2898DeriveBytes
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"database/sql"
	"errors"

	"cloud-web-phoenix-customer-v1-go/pkg"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/pbkdf2"
)

func orderedSPParamKeys(params map[string]interface{}) []string {
	// Prefer a stable, .NET-like order for common list stored procedures.
	preferred := []string{
		"PageSize",
		"ListKey",
		"RawFilterString",
		"Filters",
		"SortBy",
		"TrackingID",
		"RawSortString",
		"RawSearchString",
		"User_SiteUsersID",
		"PageNumber",
		"SearchString",
	}

	seen := make(map[string]bool, len(params))
	keys := make([]string, 0, len(params))
	for _, k := range preferred {
		if _, ok := params[k]; ok {
			keys = append(keys, k)
			seen[k] = true
		}
	}

	rest := make([]string, 0, len(params))
	for k := range params {
		if !seen[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	keys = append(keys, rest...)
	return keys
}

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
	} else if tfaType != "SMS" && tfaType != "AuthenticatorApp" {
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

// VerifyTotpWithSkew validates TOTP allowing asymmetric previous/future epochs (30s steps).
func VerifyTotpWithSkew(code string, secret string, prevEpochs int, futureEpochs int) bool {
	if strings.TrimSpace(secret) == "" || strings.TrimSpace(code) == "" {
		return false
	}
	cleanCode := strings.ReplaceAll(code, " ", "")
	if prevEpochs < 0 {
		prevEpochs = 0
	}
	if futureEpochs < 0 {
		futureEpochs = 0
	}

	opts := totp.ValidateOpts{
		Period:    30,
		Skew:      0,
		Digits:    6,
		Algorithm: otp.AlgorithmSHA1,
	}

	now := time.Now().UTC()
	for step := -prevEpochs; step <= futureEpochs; step++ {
		t := now.Add(time.Duration(step*30) * time.Second)
		valid, _ := totp.ValidateCustom(cleanCode, secret, t, opts)
		if valid {
			return true
		}
	}
	return false
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
		// var prepareAccessElements map[string]interface{}
		switch v := scopes.(type) {
		case string:
			if v != "" {
				if err := json.Unmarshal([]byte(v), &parsedScopes); err == nil {
					if len(parsedScopes) > 0 {
						details["scopes"] = normalizeScopesAny(parsedScopes)
					}
				}
			}
		case []map[string]interface{}:
			if len(v) > 0 {
				details["scopes"] = normalizeScopesAny(v)
			}
		case []interface{}:
			if len(v) > 0 {
				details["scopes"] = normalizeScopesAny(v)
			}
		}
	}
	return details
}

func normalizeScopesAny(raw interface{}) interface{} {
	if raw == nil {
		return nil
	}

	// The API expects: []{displayName,path,position,usageType,childElements}
	// and nested childElements with the same shape.

	switch t := raw.(type) {
	case []map[string]interface{}:
		out := make([]interface{}, 0, len(t))
		for _, node := range t {
			out = append(out, normalizeScopeNode(node))
		}
		return out
	case []interface{}:
		out := make([]interface{}, 0, len(t))
		for _, item := range t {
			switch node := item.(type) {
			case map[string]interface{}:
				out = append(out, normalizeScopeNode(node))
			default:
				// ignore unknown shapes
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	default:
		return nil
	}
}

func normalizeScopeNode(node map[string]interface{}) map[string]interface{} {
	if node == nil {
		return map[string]interface{}{
			"displayName":   "",
			"path":          "",
			"position":      nil,
			"usageType":     nil,
			"childElements": nil,
		}
	}

	displayName := getStringFromMapAny(node, []string{"displayName", "DisplayName"})
	path := getStringFromMapAny(node, []string{"path", "Path"})
	position := getAnyFromMapAny(node, []string{"position", "Position"})
	usageType := getAnyFromMapAny(node, []string{"usageType", "UsageType"})
	childrenRaw := getAnyFromMapAny(node, []string{"childElements", "ChildElements"})
	children := normalizeScopesAny(childrenRaw)

	// Match your expected response: if no children => null
	if childrenSlice, ok := children.([]interface{}); ok {
		if len(childrenSlice) == 0 {
			children = nil
		}
	}

	return map[string]interface{}{
		"displayName":   displayName,
		"path":          path,
		"position":      position,
		"usageType":     usageType,
		"childElements": children,
	}
}

func getStringFromMapAny(m map[string]interface{}, keys []string) string {
	for _, k := range keys {
		if m == nil {
			return ""
		}
		v, ok := m[k]
		if !ok || v == nil {
			continue
		}
		switch s := v.(type) {
		case string:
			return s
		case []byte:
			return string(s)
		default:
			// if it's numeric/bool etc, stringify like fmt does? skip for now
		}
	}
	return ""
}

func getAnyFromMapAny(m map[string]interface{}, keys []string) interface{} {
	for _, k := range keys {
		if m == nil {
			return nil
		}
		if v, ok := m[k]; ok {
			return v
		}
	}
	return nil
}

func generate6DigitCode() (string, error) {
	// Cryptographically secure 6-digit code (000000-999999)
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	n := int(buf[0])<<24 | int(buf[1])<<16 | int(buf[2])<<8 | int(buf[3])
	if n < 0 {
		n = -n
	}
	code := n % 1000000
	return fmt.Sprintf("%06d", code), nil
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

// ------------------------ Env + row helpers ------------------------

func getIntEnvAny(keys []string, def int) int {
	for _, k := range keys {
		v := strings.TrimSpace(os.Getenv(k))
		if v == "" {
			continue
		}
		n, err := strconv.Atoi(v)
		if err == nil && n > 0 {
			return n
		}
	}
	return def
}

func getInt64(row map[string]interface{}, key string) int64 {
	if row == nil {
		return 0
	}
	v := row[key]
	if v == nil {
		return 0
	}
	switch t := v.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case float64:
		return int64(t)
	case string:
		if n, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64); err == nil {
			return n
		}
	case []byte:
		if n, err := strconv.ParseInt(strings.TrimSpace(string(t)), 10, 64); err == nil {
			return n
		}
	}
	return 0
}

func getBoolAny(row map[string]interface{}, keys []string) bool {
	if row == nil {
		return false
	}
	for _, k := range keys {
		v, ok := row[k]
		if !ok || v == nil {
			continue
		}
		switch t := v.(type) {
		case bool:
			return t
		case int64:
			return t != 0
		case int:
			return t != 0
		case float64:
			return t != 0
		case string:
			s := strings.ToLower(strings.TrimSpace(t))
			return s == "true" || s == "1" || s == "yes"
		}
	}
	return false
}

func getTimeAny(row map[string]interface{}, keys []string) (time.Time, bool) {
	if row == nil {
		return time.Time{}, false
	}
	for _, k := range keys {
		v := row[k]
		if v == nil {
			continue
		}
		switch t := v.(type) {
		case time.Time:
			return t, true
		case string:
			s := strings.TrimSpace(t)
			if s == "" {
				continue
			}
			if tt, err := time.Parse(time.RFC3339Nano, s); err == nil {
				return tt, true
			}
			if tt, err := time.Parse(time.RFC3339, s); err == nil {
				return tt, true
			}
		case []byte:
			s := strings.TrimSpace(string(t))
			if s == "" {
				continue
			}
			if tt, err := time.Parse(time.RFC3339Nano, s); err == nil {
				return tt, true
			}
			if tt, err := time.Parse(time.RFC3339, s); err == nil {
				return tt, true
			}
		}
	}
	return time.Time{}, false
}

func getStringAny(result map[string]interface{}, keys []string) string {
	for _, k := range keys {
		if s := getString(result, k); s != "" {
			return s
		}
	}
	return ""
}

func getAny(row map[string]interface{}, keys []string) interface{} {
	if row == nil {
		return nil
	}
	for _, k := range keys {
		if v, ok := row[k]; ok && v != nil {
			return v
		}
	}
	return nil
}

func getLocationJSONSafe(ip string) string {
	if strings.TrimSpace(ip) == "" {
		return ""
	}
	// Placeholder; the v2 SP expects NVARCHAR(100) Location.
	return "Unknown Location"
}

func userAgentSafe(c *gin.Context) string {
	if c == nil {
		return ""
	}
	ua := strings.TrimSpace(c.GetHeader("User-Agent"))
	if ua == "" {
		return ""
	}
	if len(ua) > 1000 {
		ua = ua[:1000]
	}
	return ua
}

// ------------------------ Login audit SP calls ------------------------

// ExecSPNonQuery executes a stored procedure that returns no resultset.
func ExecSPNonQuery(db *sql.DB, spName string, params map[string]interface{}) error {
	if db == nil {
		return errors.New("db not initialized")
	}

	query := "exec " + spName
	args := []interface{}{}

	keys := orderedSPParamKeys(params)
	for i, k := range keys {
		if i == 0 {
			query += " "
		} else {
			query += ", "
		}
		query += "@" + k + " = @" + k
		args = append(args, sql.Named(k, params[k]))
	}

	finalQuery := "exec " + spName
	for i, k := range keys {
		v := params[k]
		if i == 0 {
			finalQuery += " "
		} else {
			finalQuery += ","
		}
		finalQuery += "@" + k + "="
		if v == nil {
			finalQuery += "null"
			continue
		}
		switch val := v.(type) {
		case string:
			finalQuery += "'" + val + "'"
		case int, int64, float64:
			finalQuery += fmt.Sprintf("%v", val)
		case bool:
			if val {
				finalQuery += "1"
			} else {
				finalQuery += "0"
			}
		case time.Time:
			// .NET-style: 2026-01-26T10:53:41.982
			tt := val.UTC().Truncate(time.Millisecond)
			finalQuery += "'" + tt.Format("2006-01-02T15:04:05.000") + "'"
		case *time.Time:
			if val == nil {
				finalQuery += "null"
			} else {
				tt := val.UTC().Truncate(time.Millisecond)
				finalQuery += "'" + tt.Format("2006-01-02T15:04:05.000") + "'"
			}
		default:
			finalQuery += fmt.Sprintf("'%v'", val)
		}
	}

	pkg.Log("[SP CALL]", finalQuery)

	_, err := db.Exec(query, args...)
	if err != nil {
		pkg.Log("[SP ERROR]", err)
	}
	return err
}

// ---- FailLogin: same SP as .NET ----
func FailLogin(
	db *sql.DB,
	siteUsersId int64,
	tryLoginCount int64,
	dateLastSuccessful interface{},
	now time.Time,
	browserSummary, userAgent, ip, locationJSON string,
) {
	if db == nil || siteUsersId <= 0 {
		return
	}
	_ = ExecSPNonQuery(db, "v2_PublicRole_AuthModule_FailLogin", map[string]interface{}{
		"SiteUsersId":             siteUsersId,
		"TryLoginCount":           tryLoginCount,
		"DateLastSuccessfulLogin": dateLastSuccessful,
		"DateLastFailedLogin":     now,
		"Browser":                 browserSummary,
		"UserAgent":               userAgent,
		"IP":                      ip,
		"Location":                locationJSON,
	})
}

// ---- V1 FailLogin: /api/v1/publicrole/authmodule/login ----
func V1FailLogin(
	db *sql.DB,
	userId int64,
	tryLoginCount int64,
	dateLastSuccessful interface{},
	now time.Time,
) {
	if db == nil || userId <= 0 {
		return
	}
	_ = ExecSPNonQuery(db, "v1_PublicRole_AuthModule_FailLogin", map[string]interface{}{
		"UserId":                  userId,
		"TryLoginCount":           tryLoginCount,
		"DateLastFailedLogin":     now,
		"DateLastSuccessfulLogin": dateLastSuccessful,
	})
}

// ---- SucceedLogin: same SP as .NET ----
func SucceedLogin(
	db *sql.DB,
	siteUsersId int64,
	refreshToken string,
	refreshExpiry time.Time,
	browserSummary, userAgent, ip, locationJSON string,
	dateLastSuccessfulLogin time.Time,
	dateLastFailedLogin time.Time,
) {
	if db == nil || siteUsersId <= 0 {
		return
	}
	_ = ExecSPNonQuery(db, "v2_PublicRole_AuthModule_SucceedLogin", map[string]interface{}{
		"SiteUsersId":             siteUsersId,
		"RefreshToken":            refreshToken,
		"RefreshTokenExpiry":      refreshExpiry,
		"DateLastSuccessfulLogin": dateLastSuccessfulLogin,
		"TryLoginCount":           0,
		"Browser":                 browserSummary,
		"UserAgent":               userAgent,
		"IP":                      ip,
		"Location":                locationJSON,
		"DateLastFailedLogin":     nil,
	})
}

// ---- V1 SucceedTfaPhaseOne: SMS code + expiry, no tokens ----
func V1SucceedTfaPhaseOne(
	db *sql.DB,
	userId int64,
	dateLastSuccessful interface{},
	tfaCode string,
	tfaCodeExpiry time.Time,
) {
	if db == nil || userId <= 0 {
		return
	}
	_ = ExecSPNonQuery(db, "v1_PublicRole_AuthModule_SucceedTfaPhaseOne", map[string]interface{}{
		"UserId":                  userId,
		"TryLoginCount":           0,
		"DateLastSuccessfulLogin": dateLastSuccessful,
		"DateLastFailedLogin":     nil,
		"TfaCode":                 tfaCode,
		"TfaCodeExpiry":           tfaCodeExpiry,
	})
}

// ---- V1 SucceedLogin: issue refresh token + audit fields ----
func V1SucceedLogin(
	db *sql.DB,
	userId int64,
	refreshToken string,
	refreshExpiry time.Time,
	browserSummary, userAgent, ip, location string,
	now time.Time,
) {
	if db == nil || userId <= 0 {
		return
	}
	_ = ExecSPNonQuery(db, "v1_PublicRole_AuthModule_SucceedLogin", map[string]interface{}{
		"UserId":                  userId,
		"TryLoginCount":           0,
		"DateLastSuccessfulLogin": now,
		"DateLastFailedLogin":     nil,
		"RefreshToken":            refreshToken,
		"RefreshTokenExpiry":      refreshExpiry,
		"BrowserSummary":          browserSummary,
		"UserAgent":               userAgent,
		"IpAddress":               ip,
		"Location":                location,
	})
}

func clientIPSafe(c *gin.Context) string {
	if c == nil {
		return ""
	}
	ip := strings.TrimSpace(c.ClientIP())
	if ip == "" {
		return ""
	}
	if net.ParseIP(ip) == nil {
		return ""
	}
	return ip
}

func getBrowserSummaryDotNetStyle(userAgent string) string {
	ua := strings.TrimSpace(userAgent)
	if ua == "" {
		return ""
	}
	if len(ua) > 500 {
		ua = ua[:500]
	}
	return ua
}

func getLocationSafe(ip string) string {
	if strings.TrimSpace(ip) == "" {
		return ""
	}
	return "Unknown Location"
}

// func GenerateTokens(userID int, rememberMe bool, firstName string, lastName string, accountType string, userCode string) (string, string, int, int, error) {
// 	now := time.Now().UTC()

// 	// Get JWT secret and expiry from environment
// 	jwtSecret := os.Getenv("JWT_SECRET")
// 	if jwtSecret == "" {
// 		jwtSecret = "your_default_jwt_secret_here"
// 	}
// 	jwtExpiryStr := os.Getenv("JWT_EXPIRY_MINUTES")
// 	jwtExpiryMinutes := 60
// 	if jwtExpiryStr != "" {
// 		if v, err := strconv.Atoi(jwtExpiryStr); err == nil {
// 			jwtExpiryMinutes = v
// 		}
// 	}
// 	jwtExpirySeconds := jwtExpiryMinutes * 60

// 	claims := jwt.MapClaims{
// 		"sub":         strconv.Itoa(userID),
// 		"exp":         now.Add(time.Minute * time.Duration(jwtExpiryMinutes)).Unix(),
// 		"RememberMe":  rememberMe,
// 		"id":          userID,
// 		"firstName":   firstName,
// 		"lastName":    lastName,
// 		"accountType": accountType,
// 		"userCode":    userCode,
// 	}
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	accessToken, err := token.SignedString([]byte(jwtSecret))
// 	if err != nil {
// 		return "", "", 0, 0, err
// 	}

// 	refreshToken := uuid.NewString()
// 	refreshExpiry := now.Add(time.Hour * 24 * 30) // 30 days
// 	refreshTokenExpiresIn := int(refreshExpiry.Sub(now).Seconds())

// 	return accessToken, refreshToken, jwtExpirySeconds, refreshTokenExpiresIn, nil
// }

// ExtractUserID extracts user ID from JWT claims
func ExtractUser(c *gin.Context) map[string]interface{} {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		return nil
	}

	// Handle Bearer token
	token := tokenString
	if len(tokenString) > 7 && strings.ToLower(tokenString[:7]) == "bearer " {
		token = strings.TrimSpace(tokenString[7:])
	}

	jwtToken, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return nil
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}

	user := make(map[string]interface{})

	// ---- ID handling: support old (id) and .NET-like (sub) ----
	if idVal, ok := claims["id"]; ok {
		switch v := idVal.(type) {
		case float64:
			user["id"] = int(v)
		case int:
			user["id"] = v
		case string:
			if idInt, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
				user["id"] = idInt
			}
		}
	}
	if _, ok := user["id"]; !ok {
		if sub, ok := claims["sub"]; ok {
			switch v := sub.(type) {
			case string:
				if idInt, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
					user["id"] = idInt
				}
			case float64:
				user["id"] = int(v)
			}
		}
	}

	// Accept both camelCase and PascalCase claim keys
	if v, ok := claims["FirstName"].(string); ok {
		user["FirstName"] = v
	} else if v, ok := claims["firstName"].(string); ok {
		user["FirstName"] = v
	}
	if v, ok := claims["LastName"].(string); ok {
		user["LastName"] = v
	} else if v, ok := claims["lastName"].(string); ok {
		user["LastName"] = v
	}
	if v, ok := claims["AccountType"].(string); ok {
		user["AccountType"] = v
	} else if v, ok := claims["accountType"].(string); ok {
		user["AccountType"] = v
	}
	if v, ok := claims["UserCode"].(string); ok {
		user["UserCode"] = v
	} else if v, ok := claims["userCode"].(string); ok {
		user["UserCode"] = v
	}
	if len(user) == 0 {
		return nil
	}
	return user
}
func ExecSP(
	db *sql.DB,
	spName string,
	params map[string]interface{},
	mode int, // 1 = single row, 2 = multi row
) (interface{}, error) {

	if db == nil {
		return nil, errors.New("db not initialized")
	}

	// ---- Build query ----
	query := "exec " + spName
	args := []interface{}{}

	keys := orderedSPParamKeys(params)
	for i, k := range keys {
		if i == 0 {
			query += " "
		} else {
			query += ", "
		}
		query += "@" + k + " = @" + k
		args = append(args, sql.Named(k, params[k]))
	}

	// ---------- LOG (Before Execution) ----------
	finalQuery := "exec " + spName
	for i, k := range keys {
		v := params[k]
		if i == 0 {
			finalQuery += " "
		} else {
			finalQuery += ","
		}
		finalQuery += "@" + k + "="
		if v == nil {
			finalQuery += "null"
			continue
		}
		switch val := v.(type) {
		case string:
			finalQuery += "'" + val + "'"
		case int, int64, float64:
			finalQuery += fmt.Sprintf("%v", val)
		case bool:
			if val {
				finalQuery += "1"
			} else {
				finalQuery += "0"
			}
		case time.Time:
			// .NET-style: 2026-01-26T10:53:41.982
			tt := val.UTC().Truncate(time.Millisecond)
			finalQuery += "'" + tt.Format("2006-01-02T15:04:05.000") + "'"
		case *time.Time:
			if val == nil {
				finalQuery += "null"
			} else {
				tt := val.UTC().Truncate(time.Millisecond)
				finalQuery += "'" + tt.Format("2006-01-02T15:04:05.000") + "'"
			}
		default:
			finalQuery += fmt.Sprintf("'%v'", val)
		}
	}

	pkg.Log("[SP CALL]", finalQuery)

	rows, err := db.Query(query, args...)
	if err != nil {
		pkg.Log("[SP ERROR]", err)
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// ---------- MODE 1 ----------
	if mode == 1 {
		if !rows.Next() {
			return nil, sql.ErrNoRows
		}

		row, err := scanRow(columns, rows)
		if err != nil {
			return nil, err
		}

		return row, nil
	}

	// ---------- MODE 2 ----------
	if mode == 2 {
		results := []map[string]interface{}{}

		for rows.Next() {
			row, err := scanRow(columns, rows)
			if err != nil {
				return nil, err
			}
			results = append(results, row)
		}

		if len(results) == 0 {
			return nil, sql.ErrNoRows
		}

		return results, nil
	}

	return nil, errors.New("invalid mode")
}

// AsSingleRow normalizes a stored-procedure result into a single row.
// Some call-sites historically assumed mode=1 could return a slice; this keeps them resilient.
func AsSingleRow(res interface{}) (map[string]interface{}, bool) {
	switch v := res.(type) {
	case map[string]interface{}:
		return v, true
	case []map[string]interface{}:
		if len(v) == 0 {
			return nil, false
		}
		return v[0], true
	case []interface{}:
		if len(v) == 0 {
			return nil, false
		}
		row, ok := v[0].(map[string]interface{})
		return row, ok
	default:
		return nil, false
	}
}

func scanRow(columns []string, rows *sql.Rows) (map[string]interface{}, error) {
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	row := make(map[string]interface{})
	for i, col := range columns {
		val := values[i]
		if b, ok := val.([]byte); ok {
			row[col] = string(b)
		} else {
			row[col] = val
		}
	}
	return row, nil
}

func sendError(c *gin.Context, httpCode int, field string, code string) {
	c.JSON(httpCode, gin.H{
		"id":      0,
		"details": nil,
		"status":  "0",
		"errors": []gin.H{
			{"fieldName": field, "messageCode": code},
		},
	})
}

func sendUserPassError(c *gin.Context, httpCode int, code string) {
	c.JSON(httpCode, gin.H{
		"id":      0,
		"details": nil,
		"status":  "0",
		"errors": []gin.H{
			{"fieldName": "Username", "messageCode": code},
			{"fieldName": "Password", "messageCode": code},
		},
	})
}

func compressToBase64String(input string) (string, bool) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", false
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, _ = gz.Write([]byte(input))
	_ = gz.Close()

	return base64.StdEncoding.EncodeToString(buf.Bytes()), true
}

func GenerateTokens(
	userID int,
	rememberMe bool,
	firstName string,
	lastName string,
	accountType string,
	userCode string,
	allowedApiEndpointsCdl string,
) (string, string, int, int, error) {

	now := time.Now().UTC()

	jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if jwtSecret == "" {
		jwtSecret = strings.TrimSpace(os.Getenv("JWT_SIGNING_KEY"))
	}
	if jwtSecret == "" {
		return "", "", 0, 0, errors.New("JWT_SECRET/JWT_SIGNING_KEY missing")
	}

	jwtExpiryMinutes := 30
	if v := os.Getenv("JWT_EXPIRY_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			jwtExpiryMinutes = n
		}
	}
	jwtExpirySeconds := jwtExpiryMinutes * 60

	refreshExpiryMinutes := getIntEnvAny([]string{"REFRESH_TOKEN_EXPIRY_MINUTES"}, 60)
	if rememberMe {
		refreshExpiryMinutes = getIntEnvAny(
			[]string{"REFRESH_TOKEN_REMEMBERME_EXPIRY_MINUTES", "REFRESH_TOKEN_REMEMBER_ME_EXPIRY_MINUTES"},
			refreshExpiryMinutes,
		)
	}
	refreshTokenExpiresIn := refreshExpiryMinutes * 60

	rememberMeStr := "False"
	if rememberMe {
		rememberMeStr = "True"
	}

	// Scopes = gzip+base64(AllowedApiEndpointsCdl)
	scopes := ""
	if s, ok := compressToBase64String(allowedApiEndpointsCdl); ok {
		scopes = s
	}

	claims := jwt.MapClaims{
		"sub": strconv.Itoa(userID),
		"jti": uuid.New().String(),

		"iat": strconv.FormatInt(now.Unix(), 10),
		"nbf": now.Unix(),
		"exp": now.Add(time.Minute * time.Duration(jwtExpiryMinutes)).Unix(),

		"AccountType": accountType,
		"FirstName":   firstName,
		"LastName":    lastName,
		"UserCode":    userCode,
		"RememberMe":  rememberMeStr,
		"Scopes":      scopes,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", 0, 0, err
	}

	refreshToken := uuid.NewString()
	return accessToken, refreshToken, jwtExpirySeconds, refreshTokenExpiresIn, nil
}

func getString(result map[string]interface{}, key string) string {
	v, ok := result[key]
	if !ok || v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return ""
	}
}
