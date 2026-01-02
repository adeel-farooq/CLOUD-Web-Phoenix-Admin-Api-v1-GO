package routes

import (
	"cloud-web-phoenix-customer-v1-go/controllers/auth"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/v1/publicrole/authmodule/login", auth.SignIn)
	router.POST("/v1/publicrole/authmodule/tfalogin", auth.SignIn)
	router.GET("/v1/adminrole/profilemodule/userinfo", auth.UserFromToken)

}
