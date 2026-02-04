package routes

import (
	"cloud-web-phoenix-customer-v1-go/controllers/transactions"

	"github.com/gin-gonic/gin"
)

func RegisterTransactionRoutes(router *gin.RouterGroup) {
	// ==========================================================================
	// TransactionsModule - /api/v1/adminrole/transactionsmodule
	// ==========================================================================

	// List endpoints
	router.GET("/transactionsmodule/list", transactions.List)
	router.GET("/transactionsmodule/listall", transactions.ListAll)
	router.POST("/transactionsmodule/downloadall", transactions.DownloadAll)

	// Transaction details
	router.GET("/transactionsmodule/transaction-details", transactions.TransactionDetails)
	router.GET("/transactionsmodule/transaction-json", transactions.TransactionJSON)
	router.GET("/transactionsmodule/view-alerts", transactions.ViewAlerts)

	// Notes
	router.GET("/transactionsmodule/list-notes", transactions.ListNotes)
	router.GET("/transactionsmodule/edit-note", transactions.GetEditNote)
	router.POST("/transactionsmodule/addnote", transactions.AddNote)
	router.POST("/transactionsmodule/edit-note", transactions.EditNote)
	router.POST("/transactionsmodule/pin-note", transactions.PinNote)

	// Documents
	router.GET("/transactionsmodule/list-documents", transactions.ListDocuments)
	router.GET("/transactionsmodule/edit-document", transactions.GetEditDocument)
	router.POST("/transactionsmodule/adddocument", transactions.AddDocument)
	router.POST("/transactionsmodule/edit-document", transactions.EditDocument)
	router.POST("/transactionsmodule/pin-document", transactions.PinDocument)

	// Cancellation
	router.GET("/transactionsmodule/cancel", transactions.GetCancellationEligibility)
	router.POST("/transactionsmodule/cancel", transactions.CancelTransaction)

	// Fee waiver
	router.GET("/transactionsmodule/waive-fee", transactions.GetWaiveFee)
	router.POST("/transactionsmodule/waive-fee", transactions.WaiveFee)

	// ==========================================================================
	// PendingTransactionsModule - /api/v1/adminrole/pendingtransactionsmodule
	// ==========================================================================

	router.GET("/pendingtransactionsmodule/list", transactions.ListPending)
	router.GET("/pendingtransactionsmodule/listproducts", transactions.ListProducts)
	router.GET("/pendingtransactionsmodule/assign", transactions.GetAssignDetails)
	router.POST("/pendingtransactionsmodule/assign", transactions.Assign)
	router.POST("/pendingtransactionsmodule/complete", transactions.Complete)
	router.POST("/pendingtransactionsmodule/cancel", transactions.PendingCancel)
	router.POST("/pendingtransactionsmodule/markaspending", transactions.MarkAsPending)

	// ==========================================================================
	// PendingTransactionsTreasuryModule - /api/v1/adminrole/pendingtransactionstreasurymodule
	// ==========================================================================

	router.GET("/pendingtransactionstreasurymodule/list", transactions.ListPendingTreasury)
	router.GET("/pendingtransactionstreasurymodule/listproducts", transactions.ListPendingTreasuryProducts)
	router.POST("/pendingtransactionstreasurymodule/release", transactions.Release)
	router.POST("/pendingtransactionstreasurymodule/cancel", transactions.TreasuryCancel)

	// ==========================================================================
	// FrozenTransactionsModule - /api/v1/adminrole/frozentransactionsmodule
	// ==========================================================================

	router.GET("/frozentransactionsmodule/list", transactions.ListFrozen)
	router.GET("/frozentransactionsmodule/view", transactions.FrozenView)
	router.POST("/frozentransactionsmodule/unfreeze", transactions.Unfreeze)
}
