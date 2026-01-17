package routes

import (
	"cloud-web-phoenix-customer-v1-go/controllers/admin"
	"cloud-web-phoenix-customer-v1-go/controllers/dashboard"

	"github.com/gin-gonic/gin"
)

func RegisterDashboardRoutes(router *gin.RouterGroup) {

	router.GET("/dashboardmodule/listproducts", dashboard.ProductsList)
	router.GET("/dashboardmodule/depositschart", dashboard.GetChartsData)
	router.GET("/dashboardmodule/withdrawalschart", dashboard.GetChartsData)
	router.GET("/dashboardmodule/cardsissuedchart", dashboard.GetChartsData)
	router.GET("/dashboardmodule/cardspendchart", dashboard.GetChartsData)
	router.GET("/dashboardmodule/turnoverchart", dashboard.GetChartsData)
	router.GET("/dashboardmodule/revenuechart", dashboard.GetChartsData)
	router.GET("/dashboardmodule/usercounts", admin.GetUserCounts)

	// router.GET("/v1/adminrole/licenseeadminusersmodule/list", dashboard.GetAdminUsersData)
	// router.GET("/v1/adminrole/adminrolesmodule/list", dashboard.GetAdminUsersData)

}
