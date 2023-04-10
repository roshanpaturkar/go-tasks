package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/roshanpaturkar/go-tasks/controllers"
	"github.com/roshanpaturkar/go-tasks/middleware"
)

func UserRoutes(app *fiber.App) {
	route := app.Group("/api/v1/user")

	route.Post("/sign/up", controllers.UserSignUp)
	route.Post("/sign/in", controllers.UserSignIn)
	route.Get("/sign/out", middleware.Auth(), controllers.UserSignOut)
	route.Get("/sign/out/all", middleware.Auth(), controllers.UserSignOutAll)
	route.Get("/profile", middleware.Auth(), controllers.UserProfile)
}
