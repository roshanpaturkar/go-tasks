package main

import (
	"github.com/gofiber/fiber/v2"
	_ "github.com/joho/godotenv/autoload"

	"github.com/roshanpaturkar/go-tasks/routes"
)

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// Routes
	routes.UserRoutes(app)

	app.Listen(":3000")
}
