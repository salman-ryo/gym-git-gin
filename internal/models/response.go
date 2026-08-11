package models

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// SuccessResponse defines standard HTTP 200/201 JSON envelope
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// ErrorDetails defines structured error information
type ErrorDetails struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details"`
}

// ErrorResponse defines standard HTTP 4xx/5xx JSON envelope
type ErrorResponse struct {
	Success   bool         `json:"success"`
	Error     ErrorDetails `json:"error"`
	Timestamp string       `json:"timestamp"`
}

// SendSuccess sends a standardized success JSON response
func SendSuccess(c *gin.Context, statusCode int, data interface{}, message string) {
	if data == nil {
		data = gin.H{}
	}
	c.JSON(statusCode, SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// SendError sends a standardized error JSON response
func SendError(c *gin.Context, statusCode int, code string, message string, details interface{}) {
	if details == nil {
		details = []string{}
	}

	log.Printf("[API ERROR] %s %s | Status: %d | Code: %s | Message: %s | Details: %v",
		c.Request.Method,
		c.Request.URL.RequestURI(),
		statusCode,
		code,
		message,
		details,
	)

	c.JSON(statusCode, ErrorResponse{
		Success: false,
		Error: ErrorDetails{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
