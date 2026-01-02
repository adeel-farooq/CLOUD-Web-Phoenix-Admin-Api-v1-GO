package pkg

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// RecoveryWithLogger is a middleware that recovers from panics and logs the error with stack trace
func RecoveryWithLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()
				msg := fmt.Sprintf("PANIC: %v\n%s", err, stack)
				Log(msg)
				// Only return a simple message to the user
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Internal Server Error",
					"error":   true,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
