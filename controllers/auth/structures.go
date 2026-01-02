package auth

type SignInRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	RememberMe  bool   `json:"rememberMe" binding:"required"`
	AccountType int    `json:"accountType" `
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
