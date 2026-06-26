package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// Health responds with service status and the current UTC timestamp.
func Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "OK",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
