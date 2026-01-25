package routes

import (
	v1 "cloud-web-phoenix-customer-v1-go/routes/v1"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup) {
	// Public auth routes
	auth := api.Group("/v1/publicrole")
	v1.RegisterAuthRoutes(auth)

	// Admin routes
	adminrole := api.Group("/v1/adminrole")
	v1.RegisterAdminRoutes(adminrole)
	v1.RegisterDashboardRoutes(adminrole)
	v1.RegisterBusinessRoutes(adminrole)

}
