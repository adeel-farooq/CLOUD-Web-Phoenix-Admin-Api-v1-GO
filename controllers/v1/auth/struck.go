package auth

type CredentialsDto struct {
	Username   string `json:"Username"`
	Password   string `json:"Password"`
	RememberMe bool   `json:"RememberMe"`
}

type ErrorItem struct {
	FieldName   string `json:"fieldName"`
	MessageCode string `json:"messageCode"`
}

type LoginSuccessDetails struct {
	AccessToken           *string `json:"accessToken,omitempty"`
	ExpiresIn             *int    `json:"expiresIn,omitempty"`
	RefreshToken          *string `json:"refreshToken,omitempty"`
	RefreshTokenExpiresIn *int    `json:"refreshTokenExpiresIn,omitempty"`

	BTwoFactorAppAuthEnabled bool `json:"bTwoFactorAppAuthEnabled"`
	BTwoFactorSMSAuthEnabled bool `json:"bTwoFactorSMSAuthEnabled"`
	BEmailVerified           bool `json:"bEmailVerified"`

	Scopes interface{} `json:"scopes,omitempty"` // parsed NavigationScopesJson (array)
}

type LoginResponse struct {
	Id      int                  `json:"id"`
	Details *LoginSuccessDetails `json:"details"`
	Status  string               `json:"status"` // "1" success, "0" fail
	Errors  []ErrorItem          `json:"errors"`
}
type GinCtx interface {
	ShouldBindJSON(any) error
	JSON(int, any)
	GetHeader(string) string
	ClientIP() string
}
type Deps struct {
	DB          any
	ExecSP      func(db any, sp string, params map[string]interface{}, mode int) (interface{}, error)
	AsSingleRow func(res interface{}) (map[string]interface{}, bool)

	VerifyPassword func(plain, hash string) (bool, error)

	MaxTryLoginCount     int
	TfaCodeExpiryMinutes int
}
