package server

import (
	"core/server/handlers"

	"github.com/gofiber/fiber/v2"
)

// RegisterRoutes wires all routes to their handlers.
func RegisterRoutes(app *fiber.App) {
	app.Get("/", handlers.Root)
	app.Get("/health", handlers.Health)
}
