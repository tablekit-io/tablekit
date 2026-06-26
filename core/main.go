package main

import (
	"log"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "tablekit",
		Short: "tablekit core service",
	}

	serve := &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := fiber.New(fiber.Config{
				JSONEncoder: sonic.ConfigFastest.Marshal,
				JSONDecoder: sonic.ConfigFastest.Unmarshal,
			})

			app.Get("/", func(c *fiber.Ctx) error {
				return c.SendString("hello world")
			})

			app.Get("/health", func(c *fiber.Ctx) error {
				return c.JSON(fiber.Map{
					"status":    "OK",
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				})
			})

			return app.Listen(":8080")
		},
	}

	root.AddCommand(serve)

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
