package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/roshanpaturkar/go-tasks/controllers"
	"github.com/roshanpaturkar/go-tasks/middleware"
)

func TaskRoutes(app *fiber.App) {
	route := app.Group("/api/v1/task")

	route.Post("/", middleware.Auth(), middleware.ValidateJwt(), controllers.CreateTask)
}