package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware configures Cross-Origin Resource Sharing based on allowed environment URLs
func CORSMiddleware(allowedOriginsStr string) gin.HandlerFunc {
	origins := parseOrigins(allowedOriginsStr)

	return func(c *gin.Context) {
		reqOrigin := c.GetHeader("Origin")

		allowedOrigin := isOriginAllowed(reqOrigin, origins)
		if allowedOrigin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, x-timezone, X-Timezone")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func parseOrigins(originsStr string) []string {
	if originsStr == "" {
		return []string{"http://localhost:3000", "http://127.0.0.1:3000"}
	}

	rawOrigins := strings.Split(originsStr, ",")
	var origins []string
	for _, o := range rawOrigins {
		trimmed := strings.TrimSpace(o)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

func isOriginAllowed(reqOrigin string, allowedOrigins []string) string {
	if reqOrigin == "" {
		if len(allowedOrigins) > 0 {
			return allowedOrigins[0]
		}
		return "*"
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" || strings.EqualFold(allowed, reqOrigin) {
			return reqOrigin
		}
	}

	// Fallback to request origin for development ease if localhost
	if strings.HasPrefix(reqOrigin, "http://localhost:") || strings.HasPrefix(reqOrigin, "http://127.0.0.1:") {
		return reqOrigin
	}

	return ""
}
