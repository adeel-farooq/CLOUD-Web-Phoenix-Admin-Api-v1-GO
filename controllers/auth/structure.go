package auth

import "time"

type SignInRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	RememberMe  bool   `json:"rememberMe" `
	AccountType int    `json:"accountType" `
	TFACode     string `json:"tfaCode"`
	TFAType     string `json:"tfaType"`
}
type UserResponse struct {
	CustomersId                   *int
	CustomerUsersCustomersId      *int
	CustomerAccountType           *string
	EmailVerificationCode         *string
	SiteName                      *string
	CultureInfo                   *string
	BFrozen                       *bool
	EmailVerificationCodeExpiry   *string
	BDocumentVerified             *bool
	TwoFactorSMSRequestCounter    *int
	DateLastTwoFactorSMSRequested *string
	SiteUsersId                   int
	TryLoginCount                 int
	PasswordHash                  string
	DateLastFailedLogin           *string
	DateLastSuccessfulLogin       *string
	BTwoFactorAppAuthEnabled      bool
	BTwoFactorSMSAuthEnabled      bool
	BEmailVerified                bool
	PhoneNumber                   *string
	EmailAddress                  *string
	BSuppressed                   bool
	TfaCode                       *string
	TfaCodeExpiry                 *string
	TotpSharedSecret              *string
	FirstName                     *string
	LastName                      *string
	UserCode                      *string
	AllowedAPIEndpointsCDL        *string
	NavigationScopesJson          *string
}

// LoginStatusTypes enum
const (
	Success           = 0
	AccountLocked     = 1
	PasswordInvalid   = 2
	TfaCodeInvalid    = 3
	TfaCodeExpired    = 4
	SMSTfaEnabled     = 5
	AppTfaEnabled     = 6
	UserNotFound      = 7
	AccountSuppressed = 8
	NoTfaEnabled      = 9
	EmailNotVerified  = 10
	TfaTypeInvalid    = 11
)

// TfaTypes enum
const (
	TfaAuthenticatorApp = "AuthenticatorApp"
	TfaSMS              = "SMS"
)

type loginMeta struct {
	Browser   string
	UserAgent string
	IP        string
	Location  string
}

// ---------- Request DTO ----------
type TfaLoginRequest struct {
	// .NET me AccountType JSON ignore hota hai, but user bhej de to ignore ho jaye
	AccountType *int   `json:"accountType,omitempty"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	RememberMe  bool   `json:"rememberMe"`
	TfaCode     string `json:"tfaCode"`
	TfaType     string `json:"tfaType"` // "SMS" | "AuthenticatorApp"
}

// ---------- DB Row (from v1_PublicRole_AuthModule_GetSiteUsersAuthData) ----------
type SiteUsersAuthData struct {
	SiteUsersId              int64      `db:"SiteUsersId"`
	TryLoginCount            int        `db:"TryLoginCount"`
	PasswordHash             string     `db:"PasswordHash"`
	DateLastFailedLogin      *time.Time `db:"DateLastFailedLogin"`
	DateLastSuccessfulLogin  *time.Time `db:"DateLastSuccessfulLogin"`
	BTwoFactorAppAuthEnabled bool       `db:"bTwoFactorAppAuthEnabled"`
	BTwoFactorSMSAuthEnabled bool       `db:"bTwoFactorSMSAuthEnabled"`
	BEmailVerified           bool       `db:"bEmailVerified"`
	PhoneNumber              *string    `db:"PhoneNumber"`
	EmailAddress             *string    `db:"EmailAddress"`
	BSuppressed              bool       `db:"bSuppressed"`
	TfaCode                  *string    `db:"TfaCode"`
	TfaCodeExpiry            *time.Time `db:"TfaCodeExpiry"`
	TotpSharedSecret         *string    `db:"TotpSharedSecret"`
	FirstName                *string    `db:"FirstName"`
	LastName                 *string    `db:"LastName"`
	UserCode                 *string    `db:"UserCode"`
	AllowedApiEndpointsCdl   *string    `db:"AllowedApiEndpointsCdl"`
	NavigationScopesJson     *string    `db:"NavigationScopesJson"`
}

// ---------- Response Models (match .NET JSON) ----------
type ErrorResult struct {
	ErrorType   string `json:"-"` // internal only for HTTP status selection
	FieldName   string `json:"fieldName"`
	MessageCode string `json:"messageCode"`
}

type ObjectResult[T any] struct {
	Id      int64         `json:"id,omitempty"`
	Details T             `json:"details"`
	Status  string        `json:"status"`
	Errors  []ErrorResult `json:"errors"`
}

type LoginSuccessDetails struct {
	AccessToken              any  `json:"accessToken"`
	ExpiresIn                int  `json:"expiresIn"`
	RefreshToken             any  `json:"refreshToken"`
	RefreshTokenExpiresIn    int  `json:"refreshTokenExpiresIn"`
	BTwoFactorAppAuthEnabled bool `json:"bTwoFactorAppAuthEnabled"`
	BTwoFactorSMSAuthEnabled bool `json:"bTwoFactorSMSAuthEnabled"`
	BEmailVerified           bool `json:"bEmailVerified"`
	Scopes                   any  `json:"scopes,omitempty"`
}

type NavigationScope struct {
	DisplayName   string            `json:"displayName"`
	Path          string            `json:"path"`
	Position      *string           `json:"position"`
	UsageType     *string           `json:"usageType"`
	ChildElements []NavigationScope `json:"childElements"`
}

const (
	ErrorBadRequest      = "BadRequest"
	ErrorUnauthenticated = "Unauthenticated"
)

// ------------------- config (match .NET AuthSettings) -------------------
type AuthSettings struct {
	JwtSigningKey                       string
	JwtExpiryMinutes                    int
	JwtIssuer                           string
	JwtAudience                         string
	RefreshTokenExpiryMinutes           int
	RefreshTokenRememberMeExpiryMinutes int
	TryLoginCounterMax                  int
	LoginFailedAccountLockTimeMinutes   int
	TotpAllowedPreviousEpochs           int
	TotpAllowedFutureEpochs             int
}
