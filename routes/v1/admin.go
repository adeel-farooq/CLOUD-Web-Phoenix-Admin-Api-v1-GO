package routes

import (
	"cloud-web-phoenix-customer-v1-go/controllers/admin"
	"cloud-web-phoenix-customer-v1-go/controllers/auth"

	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(router *gin.RouterGroup) {

	router.GET("/profilemodule/userinfo", auth.UserFromToken)

	router.GET("/adminusersmodule/list", admin.GetAdminUsersData)
	router.GET("/licenseeadminusersmodule/list", admin.GetLicenseAdminUsersData)
	router.GET("/adminrolesmodule/list", admin.GetAdminRolesData)

}
