package routes

import (
	v1 "cloud-web-phoenix-customer-v1-go/routes/v1"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup) {
	// Public auth routes
	auth := api.Group("/v1/publicrole/authmodule")
	v1.RegisterAuthRoutes(auth)

	// Admin profile routes
	profile := api.Group("/v1/adminrole/profilemodule")
	v1.RegisterAdminRoutes(profile)

	// Dashboard routes
	dashboardGroup := api.Group("/v1/adminrole/dashboardmodule")
	v1.RegisterDashboardRoutes(dashboardGroup)

	// Admin users routes
	adminGroup := api.Group("/v1/adminrole/adminusersmodule")
	v1.RegisterAdminRoutes(adminGroup)
}
