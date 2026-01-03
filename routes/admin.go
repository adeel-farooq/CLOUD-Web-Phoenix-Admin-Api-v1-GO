package routes

import (
	"cloud-web-phoenix-customer-v1-go/controllers/admin"

	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(router *gin.RouterGroup) {

	router.GET("/v1/adminrole/dashboardmodule/listproducts", admin.ProductsList)
	router.GET("/v1/adminrole/dashboardmodule/depositschart", admin.GetChartsData)
	router.GET("/v1/adminrole/dashboardmodule/withdrawalschart", admin.GetChartsData)
	router.GET("/v1/adminrole/dashboardmodule/cardsissuedchart", admin.GetChartsData)
	router.GET("/v1/adminrole/dashboardmodule/cardspendchart", admin.GetChartsData)
	router.GET("/v1/adminrole/dashboardmodule/turnoverchart", admin.GetChartsData)
	router.GET("/v1/adminrole/dashboardmodule/revenuechart", admin.GetChartsData)
	router.GET("/v1/adminrole/dashboardmodule/usercounts", admin.GetChartsData)
	router.GET("/v1/adminrole/adminusersmodule/list", admin.GetAdminUsersList)

}
