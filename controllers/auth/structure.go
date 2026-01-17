package auth

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
