package routes

import (
	"cloud-web-phoenix-customer-v1-go/controllers/admin"
	"cloud-web-phoenix-customer-v1-go/controllers/dashboard"

	"github.com/gin-gonic/gin"
)

func RegisterDashboardRoutes(router *gin.RouterGroup) {

	router.GET("/listproducts", dashboard.ProductsList)
	router.GET("/depositschart", dashboard.GetChartsData)
	router.GET("/withdrawalschart", dashboard.GetChartsData)
	router.GET("/cardsissuedchart", dashboard.GetChartsData)
	router.GET("/cardspendchart", dashboard.GetChartsData)
	router.GET("/turnoverchart", dashboard.GetChartsData)
	router.GET("/revenuechart", dashboard.GetChartsData)
	router.GET("/usercounts", admin.GetUserCounts)

	// router.GET("/v1/adminrole/licenseeadminusersmodule/list", dashboard.GetAdminUsersData)
	// router.GET("/v1/adminrole/adminrolesmodule/list", dashboard.GetAdminUsersData)

}
