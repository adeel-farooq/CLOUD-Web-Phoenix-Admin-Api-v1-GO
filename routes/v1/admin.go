package routes

import (
	"cloud-web-phoenix-customer-v1-go/controllers/admin"
	"cloud-web-phoenix-customer-v1-go/controllers/auth"

	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(router *gin.RouterGroup) {

	router.GET("/userinfo", auth.UserFromToken)

	router.GET("/list", admin.GetAdminUsersData)
}
