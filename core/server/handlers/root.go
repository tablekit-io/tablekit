package handlers

import "github.com/gofiber/fiber/v2"

// Root responds with a plain-text greeting.
func Root(c *fiber.Ctx) error {
	return c.SendString("hello world")
}
