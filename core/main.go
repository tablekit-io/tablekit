package main

import (
	"log"

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
			app := fiber.New()

			app.Get("/", func(c *fiber.Ctx) error {
				return c.SendString("hello world")
			})

			return app.Listen(":8080")
		},
	}

	root.AddCommand(serve)

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
