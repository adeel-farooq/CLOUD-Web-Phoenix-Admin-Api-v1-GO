package routes

import (
	"cloud-web-phoenix-customer-v1-go/controllers/business"

	"github.com/gin-gonic/gin"
)

func RegisterBusinessRoutes(router *gin.RouterGroup) {

	router.GET("/operationalassetaccountsmodule/list", business.GetOperationalAssetAccountsList)
	router.GET("/customerassetaccountsmodule/list", business.GetCustomerAssetAccountsList)
	router.GET("/businessmodule/list", business.GetBusinessList)

}
