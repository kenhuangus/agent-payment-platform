package common

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware logs HTTP requests
func LoggerMiddleware() gin.HandlerFunc {
	return gin.Logger()
}

// CORSMiddleware handles CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequestIDMiddleware adds a request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = strconv.FormatInt(time.Now().UnixNano(), 36)
		}

		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// RecoveryMiddleware handles panics and recovers from them
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			log.Printf("Panic recovered: %s", err)
		} else {
			log.Printf("Panic recovered: %v", recovered)
		}

		c.JSON(http.StatusInternalServerError, NewErrorResponse("INTERNAL_ERROR", "Internal server error"))
		c.Abort()
	})
}

// HealthCheckMiddleware provides a health check endpoint
func HealthCheckMiddleware(healthChecker func() error) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/healthz" {
			if err := healthChecker(); err != nil {
				c.JSON(http.StatusServiceUnavailable, NewErrorResponse("HEALTH_CHECK_FAILED", err.Error()))
				return
			}
			c.JSON(http.StatusOK, NewSuccessResponse(map[string]string{"status": "healthy"}))
			return
		}
		c.Next()
	}
}

// ErrorHandlerMiddleware handles errors and converts them to standard responses
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			log.Printf("Request error: %v", err.Err)

			// Try to determine the appropriate status code
			statusCode := http.StatusInternalServerError
			if c.Writer.Status() != http.StatusOK {
				statusCode = c.Writer.Status()
			}

			c.JSON(statusCode, NewErrorResponse("REQUEST_ERROR", err.Error()))
			return
		}
	}
}

// RateLimitMiddleware provides basic rate limiting (simplified version)
func RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	// This is a simplified rate limiter - in production, use a proper rate limiting library
	type clientInfo struct {
		requests []time.Time
	}

	clients := make(map[string]*clientInfo)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		if _, exists := clients[clientIP]; !exists {
			clients[clientIP] = &clientInfo{
				requests: []time.Time{now},
			}
			c.Next()
			return
		}

		client := clients[clientIP]

		// Remove requests older than 1 minute
		var recentRequests []time.Time
		for _, reqTime := range client.requests {
			if now.Sub(reqTime) < time.Minute {
				recentRequests = append(recentRequests, reqTime)
			}
		}

		if len(recentRequests) >= requestsPerMinute {
			c.JSON(http.StatusTooManyRequests, NewErrorResponse("RATE_LIMIT_EXCEEDED", "Too many requests"))
			c.Abort()
			return
		}

		client.requests = append(recentRequests, now)
		clients[clientIP] = client

		c.Next()
	}
}

// SetupCommonMiddleware sets up all common middleware for a Gin router
func SetupCommonMiddleware(router *gin.Engine, healthChecker func() error) {
	router.Use(
		LoggerMiddleware(),
		CORSMiddleware(),
		RequestIDMiddleware(),
		RecoveryMiddleware(),
		ErrorHandlerMiddleware(),
		RateLimitMiddleware(100), // 100 requests per minute
	)

	// Add health check middleware
	router.Use(HealthCheckMiddleware(healthChecker))
}
