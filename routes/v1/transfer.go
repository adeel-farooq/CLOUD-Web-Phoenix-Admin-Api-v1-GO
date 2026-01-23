package routes

import (
	"cloud-web-phoenix-customer-v1-go/controllers/transfer"

	"github.com/gin-gonic/gin"
)

func RegisterTransferRoutes(router *gin.RouterGroup) {

	router.GET("/transfersmodule/outboundtransferlist", transfer.GetOutboundTransferList)

}
