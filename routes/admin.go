package routes

import (
	"cloud-web-phoenix-customer-v1-go/controllers/admin"

	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(router *gin.RouterGroup) {

	router.GET("/v1/adminrole/dashboardmodule/listproducts", admin.ProductsList)
	router.GET("/v1/adminrole/dashboardmodule/depositschart", admin.GetDeposits)

}
