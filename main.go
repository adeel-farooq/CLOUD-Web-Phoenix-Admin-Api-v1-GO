package main

import (
	"cloud-web-phoenix-customer-v1-go/db"
	"cloud-web-phoenix-customer-v1-go/routes"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	_ = godotenv.Load()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	url := os.Getenv("APP_URL")
	if url == "" {
		url = "http://localhost"
	}

	// Initialize DB
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(404, "Not Found")
	})
	api := r.Group("/api")
	routes.RegisterRoutes(api)

	fmt.Printf("Server running at %s:%s/\n", url, port)
	log.Fatal(r.Run(":" + port))
}
