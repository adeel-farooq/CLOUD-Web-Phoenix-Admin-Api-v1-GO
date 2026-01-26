package auth

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha1" // PBKDF2 default in .NET Rfc2898DeriveBytes
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
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

	"cloud-web-phoenix-customer-v1-go/db"
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
func GenerateTokens(
	userID int,
	rememberMe bool,
	firstName string,
	lastName string,
	accountType string,
	userCode string,
	allowedApiEndpointsCdl string, // ✅ add this
) (string, string, int, int, error) {

	now := time.Now().UTC()

	// .NET parity settings
	jwtSecret := os.Getenv("JWT_SIGNING_KEY") // .NET: AuthSettings.JwtSigningKey
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET") // fallback if you already use this
	}
	if jwtSecret == "" {
		jwtSecret = "your_default_jwt_secret_here"
	}

	jwtIssuer := os.Getenv("JWT_ISSUER")     // .NET: AuthSettings.JwtIssuer
	jwtAudience := os.Getenv("JWT_AUDIENCE") // .NET: AuthSettings.JwtAudience

	jwtExpiryMinutes := 60
	if v := os.Getenv("JWT_EXPIRY_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			jwtExpiryMinutes = n
		}
	}
	jwtExpirySeconds := jwtExpiryMinutes * 60

	// ✅ .NET: Scopes = gzip+base64(AllowedApiEndpointsCdl)
	scopes, err := CompressToBase64String(allowedApiEndpointsCdl)
	if err != nil {
		return "", "", 0, 0, err
	}

	// ✅ RememberMe string exactly like .NET ("True"/"False")
	rememberMeStr := "False"
	if rememberMe {
		rememberMeStr = "True"
	}

	claims := jwt.MapClaims{
		"sub": strconv.Itoa(userID),
		"jti": uuid.New().String(),
		"iat": strconv.FormatInt(now.Unix(), 10),

		// issuer/audience (still OK even if your validator ignores them)
		"iss": jwtIssuer,
		"aud": jwtAudience,

		// time window
		"nbf": now.Unix(),
		"exp": now.Add(time.Minute * time.Duration(jwtExpiryMinutes)).Unix(),

		// ✅ Additional claims same naming as .NET
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

	// ✅ Refresh token expiry minutes like .NET
	refreshExpiryMinutes := 0
	if rememberMe {
		refreshExpiryMinutes, _ = strconv.Atoi(os.Getenv("REFRESH_TOKEN_REMEMBER_ME_EXPIRY_MINUTES"))
	} else {
		refreshExpiryMinutes, _ = strconv.Atoi(os.Getenv("REFRESH_TOKEN_EXPIRY_MINUTES"))
	}
	if refreshExpiryMinutes <= 0 {
		// fallback (aap apni default policy rakh lo)
		refreshExpiryMinutes = 60 * 24 * 30
	}

	refreshToken := uuid.New().String() // .NET: Guid.NewGuid().ToString("D")
	refreshValidUntil := now.Add(time.Minute * time.Duration(refreshExpiryMinutes))
	refreshTokenExpiresIn := int(refreshValidUntil.Sub(now).Seconds())

	return accessToken, refreshToken, jwtExpirySeconds, refreshTokenExpiresIn, nil
}

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

	// ---- ID handling (supports both .NET-style 'id' and JWT 'sub') ----
	if idVal, ok := claims["id"]; ok {
		switch v := idVal.(type) {
		case float64:
			user["id"] = int(v)
		case int:
			user["id"] = v
		case int64:
			user["id"] = int(v)
		case string:
			if idInt, err := strconv.Atoi(v); err == nil {
				user["id"] = idInt
			}
		}
	} else if subVal, ok := claims["sub"]; ok {
		switch v := subVal.(type) {
		case float64:
			user["id"] = int(v)
		case int:
			user["id"] = v
		case int64:
			user["id"] = int(v)
		case string:
			if idInt, err := strconv.Atoi(v); err == nil {
				user["id"] = idInt
			}
		}
	}

	// names & misc (support both camel + Pascal)
	if v, ok := claims["firstName"].(string); ok {
		user["FirstName"] = v
	} else if v, ok := claims["FirstName"].(string); ok {
		user["FirstName"] = v
	}
	if v, ok := claims["lastName"].(string); ok {
		user["LastName"] = v
	} else if v, ok := claims["LastName"].(string); ok {
		user["LastName"] = v
	}
	if v, ok := claims["accountType"].(string); ok {
		user["AccountType"] = v
	} else if v, ok := claims["AccountType"].(string); ok {
		user["AccountType"] = v
	}
	if v, ok := claims["userCode"].(string); ok {
		user["UserCode"] = v
	} else if v, ok := claims["UserCode"].(string); ok {
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

func buildLoginMeta(c *gin.Context) loginMeta {
	ua := c.GetHeader("User-Agent")
	ip := c.ClientIP()
	if ip == "" {
		ip = "0.0.0.0"
	}
	// browser simple guess (optional)
	browser := detectBrowser(ua)

	// location aapki side pe dependency pe hai (geoip). abhi blank rakho.
	location := ""

	return loginMeta{
		Browser:   browser,
		UserAgent: truncate(ua, 500),
		IP:        normalizeIP(ip),
		Location:  truncate(location, 100),
	}
}

func detectBrowser(ua string) string {
	u := strings.ToLower(ua)
	switch {
	case strings.Contains(u, "chrome"):
		return "Chrome"
	case strings.Contains(u, "firefox"):
		return "Firefox"
	case strings.Contains(u, "safari"):
		return "Safari"
	case strings.Contains(u, "edge"):
		return "Edge"
	default:
		return ""
	}
}

func truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n]
}

func normalizeIP(ip string) string {
	// keep as-is if it's already OK
	if net.ParseIP(ip) != nil {
		return ip
	}
	// sometimes "ip:port"
	if host, _, err := net.SplitHostPort(ip); err == nil && net.ParseIP(host) != nil {
		return host
	}
	return "0.0.0.0"
}

func getInt64(m map[string]interface{}, key string) (int64, bool) {
	v, ok := m[key]
	if !ok || v == nil {
		return 0, false
	}
	switch t := v.(type) {
	case int64:
		return t, true
	case int:
		return int64(t), true
	case float64:
		return int64(t), true
	case []byte:
		// if DB returns numeric as bytes
		// try parse omitted for brevity
		return 0, false
	default:
		return 0, false
	}
}

func getInt(m map[string]interface{}, key string) int {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	default:
		return 0
	}
}

// ---------------- Helpers ----------------

// .NET SP: v2_PublicRole_AuthModule_FailLogin
// Params (common): SiteUsersID, TryLoginCount, DateLastSuccessfulLogin, DateLastFailedLogin, Browser, UserAgent, IP, Location
func spFailLogin(siteUsersID int, row map[string]interface{}, meta loginMeta) error {

	tryCount := getInt(row, "TryLoginCount") + 1
	now := time.Now().UTC()

	_, err := ExecSP(
		db.DB,
		"v2_PublicRole_AuthModule_FailLogin",
		map[string]interface{}{
			"SiteUsersID":             siteUsersID,
			"TryLoginCount":           tryCount,
			"DateLastSuccessfulLogin": row["DateLastSuccessfulLogin"],
			"DateLastFailedLogin":     now,
			"Browser":                 meta.Browser,
			"UserAgent":               meta.UserAgent,
			"IP":                      meta.IP,
			"Location":                meta.Location,
		},
		0,
	)
	return err
}

// .NET SP: v2_PublicRole_AuthModule_SucceedLogin
// Params (common): SiteUsersID, TryLoginCount, DateLastSuccessfulLogin, DateLastFailedLogin, RefreshToken, RefreshTokenExpiry, Browser, UserAgent, IP, Location
func spSucceedLogin(siteUsersID int, row map[string]interface{}, refreshToken string, refreshTokenExpiresIn int, meta loginMeta) error {

	now := time.Now().UTC()
	refreshExpiry := now.Add(time.Duration(refreshTokenExpiresIn) * time.Second)

	_, err := ExecSP(
		db.DB,
		"v2_PublicRole_AuthModule_SucceedLogin",
		map[string]interface{}{
			"SiteUsersID":             siteUsersID,
			"TryLoginCount":           0,
			"DateLastSuccessfulLogin": now,
			"DateLastFailedLogin":     nil,
			"RefreshToken":            refreshToken,
			"RefreshTokenExpiry":      refreshExpiry, // IMPORTANT (DB SP expects it)
			"Browser":                 meta.Browser,
			"UserAgent":               meta.UserAgent,
			"IP":                      meta.IP,
			"Location":                meta.Location,
		},
		0,
	)
	return err
}
func CompressToBase64String(input string) (string, error) {
	if strings.TrimSpace(input) == "" {
		return "", nil
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(input)); err != nil {
		_ = gz.Close()
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
func loadAuthSettings() AuthSettings {
	// defaults (match staging json values you shared from .NET)
	s := AuthSettings{
		JwtSigningKey:                       os.Getenv("AUTH_JWT_SIGNING_KEY"),
		JwtExpiryMinutes:                    envInt("AUTH_JWT_EXPIRY_MINUTES", 30),
		JwtIssuer:                           os.Getenv("AUTH_JWT_ISSUER"),
		JwtAudience:                         os.Getenv("AUTH_JWT_AUDIENCE"),
		RefreshTokenExpiryMinutes:           envInt("AUTH_REFRESH_TOKEN_EXPIRY_MINUTES", 60),
		RefreshTokenRememberMeExpiryMinutes: envInt("AUTH_REFRESH_TOKEN_REMEMBER_ME_EXPIRY_MINUTES", 43200),
		TryLoginCounterMax:                  envInt("AUTH_TRY_LOGIN_COUNTER_MAX", 5),
		LoginFailedAccountLockTimeMinutes:   envInt("AUTH_LOGIN_FAILED_ACCOUNT_LOCK_TIME_MINUTES", 10),
		TotpAllowedPreviousEpochs:           envInt("AUTH_TOTP_ALLOWED_PREVIOUS_EPOCHS", 2),
		TotpAllowedFutureEpochs:             envInt("AUTH_TOTP_ALLOWED_FUTURE_EPOCHS", 1),
	}
	// fallback for your old env names (if you used these before)
	if s.JwtSigningKey == "" {
		s.JwtSigningKey = os.Getenv("JWT_SECRET")
	}
	return s
}
func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
func validateTfaLoginRequest(req *TfaLoginRequest) []ErrorResult {
	var errs []ErrorResult

	if strings.TrimSpace(req.Username) == "" {
		errs = append(errs, ErrorResult{ErrorType: ErrorBadRequest, FieldName: "Username", MessageCode: "Required"})
	}
	if strings.TrimSpace(req.Password) == "" {
		errs = append(errs, ErrorResult{ErrorType: ErrorBadRequest, FieldName: "Password", MessageCode: "Required"})
	}
	if strings.TrimSpace(req.TfaCode) == "" {
		errs = append(errs, ErrorResult{ErrorType: ErrorBadRequest, FieldName: "TfaCode", MessageCode: "Required"})
	}
	if strings.TrimSpace(req.TfaType) == "" {
		errs = append(errs, ErrorResult{ErrorType: ErrorBadRequest, FieldName: "TfaType", MessageCode: "Required"})
	} else {
		tt := strings.TrimSpace(req.TfaType)
		if tt != "SMS" && tt != "AuthenticatorApp" {
			errs = append(errs, ErrorResult{ErrorType: ErrorBadRequest, FieldName: "TfaType", MessageCode: "TfaType_Invalid"})
		}
	}
	return errs
}

// ------------------- API Result (same behavior as .NET AvamaeApiController) -------------------
func writeApiResult[T any](c *gin.Context, res ObjectResult[T]) {
	if len(res.Errors) == 0 {
		res.Status = "1"
		c.JSON(200, res)
		return
	}

	res.Status = "0"
	// pick http status from FIRST error type (like .NET)
	switch res.Errors[0].ErrorType {
	case ErrorBadRequest:
		c.JSON(400, res)
	case ErrorUnauthenticated:
		c.JSON(401, res)
	default:
		c.JSON(400, res)
	}
}

// ------------------- CompressionHelper (GZip + Base64) -------------------
func compressToBase64String(input string) *string {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, _ = gz.Write([]byte(input))
	_ = gz.Close()
	out := base64.StdEncoding.EncodeToString(buf.Bytes())
	return &out
}

// ------------------- TOTP verify (match OtpNet behavior) -------------------
func verifyTotp(code string, base32Secret string, prev int, future int, now time.Time) bool {
	code = strings.ReplaceAll(code, " ", "")
	code = strings.TrimSpace(code)
	if code == "" {
		return false
	}

	secretBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(base32Secret)))
	if err != nil {
		// some secrets may include padding; try normal decode
		secretBytes, err = base32.StdEncoding.DecodeString(strings.ToUpper(strings.TrimSpace(base32Secret)))
		if err != nil {
			return false
		}
	}

	// OtpNet default: 30 sec period, 6 digits, SHA1
	period := int64(30)
	counterNow := now.UTC().Unix() / period

	for i := -prev; i <= future; i++ {
		if totpAt(secretBytes, counterNow+int64(i)) == code {
			return true
		}
	}
	return false
}

func totpAt(secret []byte, counter int64) string {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(counter))

	mac := hmac.New(sha1.New, secret)
	_, _ = mac.Write(b[:])
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	binCode := (int(sum[offset])&0x7f)<<24 |
		(int(sum[offset+1])&0xff)<<16 |
		(int(sum[offset+2])&0xff)<<8 |
		(int(sum[offset+3]) & 0xff)

	otp := binCode % 1000000
	return fmt.Sprintf("%06d", otp)
}

// ------------------- JWT generation (match TokenHelper.cs exactly) -------------------
func generateJwtAndRefresh(auth AuthSettings, userId int64, rememberMe bool, firstName, lastName, userCode, allowedCdl string) (accessToken string, expiresIn int, refreshToken string, refreshExpiresIn int, err error) {
	now := time.Now().UTC()

	jti := uuid.NewString()

	// .NET puts iat as string, and also sets nbf + exp
	iatStr := strconv.FormatInt(now.Unix(), 10)

	// Scopes claim is compressed AllowedApiEndpointsCdl (gzip+base64)
	var scopesClaim any = nil
	if strings.TrimSpace(allowedCdl) != "" {
		sc := compressToBase64String(allowedCdl)
		if sc != nil {
			scopesClaim = *sc
		}
	}

	claims := jwt.MapClaims{
		"sub": userIdStr(userId),
		"jti": jti,
		"iat": iatStr,

		"nbf": now.Unix(),
		"exp": now.Add(time.Minute * time.Duration(auth.JwtExpiryMinutes)).Unix(),

		"AccountType": "Admin",
		"FirstName":   firstName,
		"LastName":    lastName,
		"UserCode":    userCode,
		"RememberMe":  strconv.FormatBool(rememberMe), // "True/False" in .NET? actually .ToString() => "True"/"False"
		"Scopes":      scopesClaim,
	}

	// IMPORTANT: .NET RememberMe is "True"/"False" (capital)
	claims["RememberMe"] = strings.Title(strconv.FormatBool(rememberMe))

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err = tok.SignedString([]byte(auth.JwtSigningKey))
	if err != nil {
		return "", 0, "", 0, err
	}
	expiresIn = auth.JwtExpiryMinutes * 60

	refreshToken = uuid.NewString()
	var refreshMinutes int
	if rememberMe {
		refreshMinutes = auth.RefreshTokenRememberMeExpiryMinutes
	} else {
		refreshMinutes = auth.RefreshTokenExpiryMinutes
	}
	refreshExpiresIn = refreshMinutes * 60

	return accessToken, expiresIn, refreshToken, refreshExpiresIn, nil
}

func userIdStr(id int64) string { return strconv.FormatInt(id, 10) }

// ------------------- scopes parsing (NavigationScopesJson -> []objects) -------------------
func parseScopesJson(scopesJson *string) any {
	if scopesJson == nil || strings.TrimSpace(*scopesJson) == "" {
		return nil
	}
	var arr []map[string]any
	if err := json.Unmarshal([]byte(*scopesJson), &arr); err == nil && len(arr) > 0 {
		return arr
	}
	return nil
}

// ------------------- account locked logic (exact .NET) -------------------
func isAccountLocked(auth AuthSettings, tryLoginCount int, dateLastFailed *time.Time) bool {
	if tryLoginCount <= auth.TryLoginCounterMax {
		return false
	}
	if dateLastFailed == nil {
		return false
	}
	lockUntil := dateLastFailed.UTC().Add(time.Minute * time.Duration(auth.LoginFailedAccountLockTimeMinutes))
	return lockUntil.After(time.Now().UTC())
}

// ------------------- login status helpers (errors match .NET messages/fields) -------------------
func errUserNotFound() []ErrorResult {
	return []ErrorResult{
		{ErrorType: ErrorBadRequest, FieldName: "Username", MessageCode: "Username_Or_Password_Incorrect"},
		{ErrorType: ErrorBadRequest, FieldName: "Password", MessageCode: "Username_Or_Password_Incorrect"},
	}
}

func errAccountLocked() []ErrorResult {
	return []ErrorResult{
		{ErrorType: ErrorUnauthenticated, FieldName: "Username", MessageCode: "Account_Locked"},
		{ErrorType: ErrorUnauthenticated, FieldName: "Password", MessageCode: "Account_Locked"},
	}
}

func errInvalidPassword() []ErrorResult {
	return []ErrorResult{
		{ErrorType: ErrorUnauthenticated, FieldName: "Username", MessageCode: "Username_Or_Password_Incorrect"},
		{ErrorType: ErrorUnauthenticated, FieldName: "Password", MessageCode: "Username_Or_Password_Incorrect"},
	}
}

func errTfaExpired() []ErrorResult {
	return []ErrorResult{{ErrorType: ErrorBadRequest, FieldName: "TfaCode", MessageCode: "TfaCode_Expired"}}
}

func errTfaInvalid() []ErrorResult {
	return []ErrorResult{{ErrorType: ErrorBadRequest, FieldName: "TfaCode", MessageCode: "TfaCode_Invalid"}}
}

// ------------------- browser/ip helpers (used in SP logging) -------------------
func getUserAgent(c *gin.Context) string {
	return c.GetHeader("User-Agent")
}

func getBrowserSummary(userAgent string) string {
	// .NET has a real parser; here minimal safe behavior
	ua := strings.TrimSpace(userAgent)
	if ua == "" {
		return ""
	}
	if len(ua) > 100 {
		return ua[:100]
	}
	return ua
}

func getLocationInfo(_ *gin.Context) string {
	// .NET does IP->geo lookup. Keep placeholder if you haven’t implemented yet.
	return ""
}

// ------------------- Fail/Succeed login SP param builders -------------------
func buildFailLoginParams(authData SiteUsersAuthData, c *gin.Context) map[string]any {
	now := time.Now().UTC()
	tryCount := authData.TryLoginCount + 1

	var lastSuccess any = nil
	if authData.DateLastSuccessfulLogin != nil {
		lastSuccess = authData.DateLastSuccessfulLogin.UTC()
	}
	var lastFail any = now

	return map[string]any{
		"SiteUsersId":             authData.SiteUsersId,
		"TryLoginCount":           tryCount,
		"DateLastSuccessfulLogin": lastSuccess,
		"DateLastFailedLogin":     lastFail,
		"Browser":                 getBrowserSummary(getUserAgent(c)),
		"UserAgent":               getUserAgent(c),
		"IP":                      c.ClientIP(),
		"Location":                getLocationInfo(c),
	}
}

func buildSucceedLoginParams(authData SiteUsersAuthData, refreshToken string, c *gin.Context) map[string]any {
	now := time.Now().UTC()

	var lastFail any = nil
	if authData.DateLastFailedLogin != nil {
		lastFail = authData.DateLastFailedLogin.UTC()
	}

	return map[string]any{
		"SiteUsersId":             authData.SiteUsersId,
		"TryLoginCount":           0,
		"DateLastSuccessfulLogin": now,
		"DateLastFailedLogin":     lastFail,
		"RefreshToken":            refreshToken,
		// IMPORTANT: .NET does NOT pass RefreshTokenExpiry -> keep it NOT present here
		"Browser":   getBrowserSummary(getUserAgent(c)),
		"UserAgent": getUserAgent(c),
		"IP":        c.ClientIP(),
		"Location":  getLocationInfo(c),
	}
}

// ------------------- map row -> SiteUsersAuthData (safe) -------------------
func mapRowToAuthData(row map[string]any) (SiteUsersAuthData, error) {
	var a SiteUsersAuthData

	// required: SiteUsersId, PasswordHash
	id, ok := row["SiteUsersId"]
	if !ok {
		return a, errors.New("SiteUsersId missing")
	}
	switch v := id.(type) {
	case int64:
		a.SiteUsersId = v
	case int:
		a.SiteUsersId = int64(v)
	case float64:
		a.SiteUsersId = int64(v)
	default:
		return a, errors.New("SiteUsersId invalid type")
	}

	if ph, ok := row["PasswordHash"]; ok {
		switch v := ph.(type) {
		case string:
			a.PasswordHash = v
		case []byte:
			a.PasswordHash = string(v)
		}
	}

	// TryLoginCount
	if tlc, ok := row["TryLoginCount"]; ok {
		switch v := tlc.(type) {
		case int:
			a.TryLoginCount = v
		case int64:
			a.TryLoginCount = int(v)
		case float64:
			a.TryLoginCount = int(v)
		}
	}

	// booleans
	a.BTwoFactorAppAuthEnabled = asBool(row["bTwoFactorAppAuthEnabled"])
	a.BTwoFactorSMSAuthEnabled = asBool(row["bTwoFactorSMSAuthEnabled"])
	a.BEmailVerified = asBool(row["bEmailVerified"])
	a.BSuppressed = asBool(row["bSuppressed"])

	// strings
	a.FirstName = asStringPtr(row["FirstName"])
	a.LastName = asStringPtr(row["LastName"])
	a.UserCode = asStringPtr(row["UserCode"])
	a.AllowedApiEndpointsCdl = asStringPtr(row["AllowedApiEndpointsCdl"])
	a.NavigationScopesJson = asStringPtr(row["NavigationScopesJson"])
	a.TotpSharedSecret = asStringPtr(row["TotpSharedSecret"])
	a.TfaCode = asStringPtr(row["TfaCode"])

	// time
	a.DateLastFailedLogin = asTimePtr(row["DateLastFailedLogin"])
	a.DateLastSuccessfulLogin = asTimePtr(row["DateLastSuccessfulLogin"])
	a.TfaCodeExpiry = asTimePtr(row["TfaCodeExpiry"])

	return a, nil
}

func asBool(v any) bool {
	if v == nil {
		return false
	}
	switch x := v.(type) {
	case bool:
		return x
	case int:
		return x != 0
	case int64:
		return x != 0
	case float64:
		return x != 0
	case []byte:
		return string(x) == "1" || strings.EqualFold(string(x), "true")
	case string:
		return x == "1" || strings.EqualFold(x, "true")
	default:
		return false
	}
}

func asStringPtr(v any) *string {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case string:
		if x == "" {
			return nil
		}
		return &x
	case []byte:
		s := string(x)
		if s == "" {
			return nil
		}
		return &s
	default:
		return nil
	}
}

func asTimePtr(v any) *time.Time {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case time.Time:
		t := x
		return &t
	default:
		return nil
	}
}
