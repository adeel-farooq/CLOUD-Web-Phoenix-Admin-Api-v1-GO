package routes

import (
    "cloud-web-phoenix-customer-v1-go/controllers/approvals"

    "github.com/gin-gonic/gin"
)

// RegisterApprovalRoutes sets up all endpoints for Approvals module
func RegisterApprovalRoutes(r *gin.RouterGroup) {
    // List endpoint
    r.GET("/approvalsmodule/list", approvals.GetApprovalsList)

    // View endpoint
    r.GET("/approvalsmodule/view", approvals.ViewApprovalDetails)

    // Process endpoint
    r.POST("/approvalsmodule/process", approvals.ProcessApproval)
}
