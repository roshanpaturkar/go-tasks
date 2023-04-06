package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/roshanpaturkar/go-tasks/controllers"
)

func UserRoutes(app *fiber.App) {
	route := app.Group("/api/v1/user")

	route.Post("/sign/up", controllers.UserSignUp)
	route.Post("/sign/in", controllers.UserSignIn)
}
