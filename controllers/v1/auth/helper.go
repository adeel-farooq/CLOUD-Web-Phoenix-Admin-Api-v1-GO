package auth

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"

	"cloud-web-phoenix-customer-v1-go/global"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"strings"
	"time"
)

func sendLoginFail(c GinCtx, field1, field2, code string, httpStatus int) {
	resp := LoginResponse{
		Id:      0,
		Details: nil,
		Status:  "0",
		Errors: []ErrorItem{
			{FieldName: field1, MessageCode: code},
		},
	}
	if field2 != "" {
		resp.Errors = append(resp.Errors, ErrorItem{FieldName: field2, MessageCode: code})
	}
	c.JSON(httpStatus, resp)
}

func sendLoginSuccess(c GinCtx, id int, details LoginSuccessDetails) {
	resp := LoginResponse{
		Id:      id,
		Details: &details,
		Status:  "1",
		Errors:  []ErrorItem{},
	}
	c.JSON(200, resp)
}

// .NET style SMS code: 6 digits (100000-999999)
func generate6DigitCode() (string, error) {
	nBig, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	code := 100000 + nBig.Int64()
	return fmt.Sprintf("%d", code), nil
}

func parseScopesJson(raw interface{}) interface{} {
	// .NET: NavigationScopesJson -> IList<NavigationScopesDto>
	// yahan generic parse
	s := ""
	switch v := raw.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var out interface{}
	if err := json.Unmarshal([]byte(s), &out); err == nil {
		return out
	}
	return nil
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

	refreshExpiryMinutes := global.GetIntEnvAny([]string{"REFRESH_TOKEN_EXPIRY_MINUTES"}, 60)
	if rememberMe {
		refreshExpiryMinutes = global.GetIntEnvAny(
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
