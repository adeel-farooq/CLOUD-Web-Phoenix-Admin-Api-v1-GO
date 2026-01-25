package pkg

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware enables browser access to the API.
//
// Env vars:
//   - CORS_ALLOW_ORIGINS: comma-separated list of allowed origins (e.g. "http://localhost:3000,https://admin.example.com")
//     Use "*" to allow all origins.
//   - CORS_ALLOW_CREDENTIALS: "true"/"false" (default false)
func CORSMiddleware() gin.HandlerFunc {
	allowCredentials := strings.EqualFold(strings.TrimSpace(os.Getenv("CORS_ALLOW_CREDENTIALS")), "true")

	cfg := cors.Config{
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"sentry-trace",
			"baggage",
		},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: allowCredentials,
		MaxAge:           12 * time.Hour,
	}

	// TEMP: allow all origins (wide open). This matches your request for now.
	// Using AllowOriginFunc (instead of AllowAllOrigins='*') keeps credentials support working.
	cfg.AllowOriginFunc = func(origin string) bool { return true }
	return cors.New(cfg)
}
