package router

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/shashank-mugiwara/joyboy/pkg/logging"
)

func New() *echo.Echo {
	e := echo.New()
	e.Logger.SetLevel(log.DEBUG)

	// Initialize structured logger
	structuredLogger := logging.GetDefaultLogger()

	e.Pre(middleware.RemoveTrailingSlash())

	// Add request ID middleware first to ensure all requests have IDs
	e.Use(logging.RequestIDMiddleware())

	// Add structured logging middleware
	e.Use(logging.StructuredLoggingMiddleware(structuredLogger))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))
	e.Validator = NewValidator()
	return e
}
