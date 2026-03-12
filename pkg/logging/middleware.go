package logging

import (
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	// RequestIDHeader is the HTTP header name for request ID
	RequestIDHeader = "X-Request-ID"
	// RequestIDContextKey is the context key for storing request ID
	RequestIDContextKey = "request_id"
)

// RequestIDMiddleware generates a unique request ID for each request
// and adds it to the response header and context
func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			// Check if request already has an ID (from upstream service)
			requestID := req.Header.Get(RequestIDHeader)
			if requestID == "" {
				// Generate new UUID for this request
				requestID = uuid.New().String()
			}

			// Set request ID in response header for downstream services
			res.Header().Set(RequestIDHeader, requestID)

			// Store request ID in context for handlers to access
			c.Set(RequestIDContextKey, requestID)

			return next(c)
		}
	}
}

// StructuredLoggingMiddleware logs HTTP requests in structured JSON format
func StructuredLoggingMiddleware(logger *Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			req := c.Request()

			// Get request ID from context
			requestID := GetRequestID(c)

			// Call the next handler
			err := next(c)

			// Calculate duration
			duration := time.Since(start)
			durationMs := float64(duration.Nanoseconds()) / 1e6

			// Get response details
			res := c.Response()
			statusCode := res.Status

			// Log the request details
			fields := map[string]interface{}{
				"request_id":   requestID,
				"method":       req.Method,
				"path":         req.URL.Path,
				"status_code":  statusCode,
				"duration_ms":  durationMs,
				"remote_addr":  req.RemoteAddr,
				"user_agent":   req.UserAgent(),
			}

			// Add query parameters if present
			if req.URL.RawQuery != "" {
				fields["query"] = req.URL.RawQuery
			}

			// Log at appropriate level based on status code
			message := "HTTP request completed"
			if statusCode >= 500 {
				logger.Error(message, fields)
			} else if statusCode >= 400 {
				logger.Warn(message, fields)
			} else {
				logger.Info(message, fields)
			}

			return err
		}
	}
}

// GetRequestID retrieves the request ID from the Echo context
func GetRequestID(c echo.Context) string {
	requestID, ok := c.Get(RequestIDContextKey).(string)
	if !ok {
		return ""
	}
	return requestID
}

// LogWithRequestID logs a message with the request ID from context
func LogWithRequestID(c echo.Context, logger *Logger, level LogLevel, message string, additionalFields map[string]interface{}) {
	requestID := GetRequestID(c)

	fields := map[string]interface{}{
		"request_id": requestID,
	}

	// Merge additional fields
	for k, v := range additionalFields {
		fields[k] = v
	}

	switch level {
	case DEBUG:
		logger.Debug(message, fields)
	case INFO:
		logger.Info(message, fields)
	case WARN:
		logger.Warn(message, fields)
	case ERROR:
		logger.Error(message, fields)
	case FATAL:
		logger.Fatal(message, fields)
	}
}
