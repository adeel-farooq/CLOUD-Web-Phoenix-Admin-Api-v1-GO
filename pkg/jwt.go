package pkg

import (
    "net/http"
    "os"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

// JWTAuthMiddleware validates JWT and sets claims in context
func JWTAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
            c.Abort()
            return
        }

        // Extract Bearer token
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
            c.Abort()
            return
        }
        tokenString := parts[1]

        // Parse and validate token
        jwtSecret := os.Getenv("JWT_SECRET")
        if jwtSecret == "" {
            jwtSecret = "your_default_jwt_secret_here"
        }

        token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
            if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrSignatureInvalid
            }
            return []byte(jwtSecret), nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
            c.Abort()
            return
        }

        // Save claims in context
        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            c.Set("claims", claims)
        }

        c.Next()
    }
}


