package main

import (
	"github.com/gofiber/fiber/v2"
	_ "github.com/joho/godotenv/autoload"

	"github.com/roshanpaturkar/go-tasks/middleware"
	"github.com/roshanpaturkar/go-tasks/routes"
)

func main() {
	app := fiber.New()
	
	middleware.FiberMiddleware(app)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// Routes
	routes.UserRoutes(app)
	routes.TaskRoutes(app)

	app.Listen(":3000")
}
