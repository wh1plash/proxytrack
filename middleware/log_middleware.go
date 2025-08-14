package middleware

import (
	"log/slog"
	"proxytrack/api"
	"time"

	"github.com/gofiber/fiber/v2"
)

func LoggingHandlerDecorator(handler fiber.Handler) fiber.Handler {
	logger := slog.Default()
	return func(c *fiber.Ctx) error {
		status := fiber.StatusOK
		errorType := "none"
		start := time.Now()
		errors := make(map[string]string)
		err := handler(c)
		if err != nil {
			switch e := err.(type) {
			case api.Error:
				status = e.Code
				errors["error"] = e.Message
				errorType = "Api error"

			case api.ValidationError:
				status = e.Status
				errors = e.Errors
				errorType = "Validation error"
			default:
				status = fiber.StatusInternalServerError
				errors["error"] = err.Error()
				errorType = "Internal server error"
			}
		}
		duration := time.Since(start)
		method := c.Method()
		path := c.Path()

		logger.Info("New request:", "method", method, "path", path, "status", status, "errors", errors, "message", errorType, "duration", duration)
		// fmt.Println(string(c.Response().Body()))
		// fmt.Println("-----------------------------------------------------")
		return err
	}
}
