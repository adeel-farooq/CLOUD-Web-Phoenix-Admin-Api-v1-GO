package routes

import (
	"cloud-web-phoenix-customer-v1-go/controllers/auth"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(router *gin.RouterGroup) {
	router.POST("/login", auth.SignIn)
	router.POST("/tfalogin", auth.SignIn)

}
