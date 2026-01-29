package routes

import (
	"github.com/gin-gonic/gin"

	"cloud-web-phoenix-customer-v1-go/controllers/approval"
	"cloud-web-phoenix-customer-v1-go/pkg"
)

// RegisterApprovalRoutes registers all approval module routes
func RegisterApprovalRoutes(router *gin.RouterGroup) {
	approvalGroup := router.Group("/approvaltypesmodule")

	// Apply JWT middleware
	approvalGroup.Use(pkg.JWTMiddleware())

	// List approval types
	approvalGroup.GET("/list", approval.GetApprovalTypesListAsync)

	// Get approval type for editing
	approvalGroup.GET("/edit", approval.GetEditApprovalTypesAsync)

	// Update approval type
	approvalGroup.POST("/edit", approval.EditApprovalTypesAsync)

	// View approval type
	approvalGroup.GET("/view", approval.ViewApprovalTypesAsync)
}
