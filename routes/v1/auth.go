package routes

import (
	auth2 "cloud-web-phoenix-customer-v1-go/controllers/auth"
	v1auth "cloud-web-phoenix-customer-v1-go/controllers/v1/auth"
	"cloud-web-phoenix-customer-v1-go/db"
	"cloud-web-phoenix-customer-v1-go/global"
	"database/sql"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(router *gin.RouterGroup) {
	// router.POST("/authmodule/login", auth.SignIn)
	// router.POST("/authmodule/tfalogin", auth.TfaLogin)
	deps := v1auth.Deps{
		DB: db.DB,
		ExecSP: func(dbAny any, sp string, params map[string]interface{}, mode int) (interface{}, error) {
			realDB, _ := dbAny.(*sql.DB)
			return auth2.ExecSP(realDB, sp, params, mode)
		},
		AsSingleRow: auth2.AsSingleRow,
		VerifyPassword: func(plain, hash string) (bool, error) {
			return auth2.VerifyPassword(plain, hash)
		},
		MaxTryLoginCount:     global.GetIntEnvAny([]string{"TRY_LOGIN_COUNTER_MAX"}, 5),
		TfaCodeExpiryMinutes: global.GetIntEnvAny([]string{"TFA_CODE_EXPIRY_MINUTES"}, 10),
	}

	h := v1auth.LoginHandler(deps)
	router.POST("/authmodule/login", func(c *gin.Context) {
		h(c)
	})

}
