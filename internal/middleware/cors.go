package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware configures Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow requests from Vercel and localhost
		origin := c.Request.Header.Get("Origin")

		// List of allowed origins
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"https://*.vercel.app",
		}

		// Check if origin is allowed (simplified check for Vercel)
		if origin != "" {
			// Allow all Vercel domains and localhost
			if contains([]string{"localhost", "vercel.app"}, origin) {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Helper function to check if origin contains allowed domain
func contains(allowed []string, origin string) bool {
	for _, domain := range allowed {
		if len(origin) >= len(domain) {
			// Simple substring check
			for i := 0; i <= len(origin)-len(domain); i++ {
				if origin[i:i+len(domain)] == domain {
					return true
				}
			}
		}
	}
	return false
}
