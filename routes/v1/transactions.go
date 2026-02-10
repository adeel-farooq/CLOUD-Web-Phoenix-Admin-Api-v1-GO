package routes

import (
	"cloud-web-phoenix-customer-v1-go/controllers/transactions"

	"github.com/gin-gonic/gin"
)

func RegisterTransactionRoutes(router *gin.RouterGroup) {
	// Matches .NET controller route:
	// [Route("api/v1/adminrole/transactionsmodule")]
	router.GET("/transactionsmodule/list", transactions.List)
	router.GET("/transactionsmodule/listall", transactions.ListAll)
	router.GET("/transactionsmodule/transaction-details", transactions.TransactionDetails)
	router.GET("/pendingtransactionsmodule/list", transactions.ListPending)
	router.GET("/pendingtransactionsmodule/listproducts", transactions.ListProducts)
	router.GET("/pendingtransactionstreasurymodule/list", transactions.ListPendingTreasury)
	router.GET("/pendingtransactionstreasurymodule/listproducts", transactions.ListPendingTreasuryProducts)
	router.GET("/transactionsmodule/view-alerts", transactions.ViewAlerts)

	// We'll add next endpoints here one-by-one:
	// router.POST("/transactionsmodule/downloadall", transactions.DownloadAll)
	// router.GET("/transactionsmodule/list-documents", transactions.ListDocuments)
	// ...
}
