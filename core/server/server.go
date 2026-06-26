package server

import (
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
)

// New builds the Fiber app with sonic JSON config and registered routes.
func New() *fiber.App {
	app := fiber.New(fiber.Config{
		JSONEncoder: sonic.ConfigFastest.Marshal,
		JSONDecoder: sonic.ConfigFastest.Unmarshal,
	})

	RegisterRoutes(app)

	return app
}

// Start builds the app and listens on addr.
func Start(addr string) error {
	return New().Listen(addr)
}
