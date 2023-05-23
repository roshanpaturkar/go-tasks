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
	route.Get("/sign/out", middleware.Auth(), middleware.ValidateJwt(), controllers.UserSignOut)
	route.Get("/sign/out/all", middleware.Auth(), middleware.ValidateJwt(), controllers.UserSignOutAll)
	route.Get("/profile", middleware.Auth(), middleware.ValidateJwt(), controllers.UserProfile)
	route.Post("/avatar", middleware.Auth(), middleware.ValidateJwt(), controllers.UploadUserAvatar)
	route.Get("/avatar", middleware.Auth(), middleware.ValidateJwt(), controllers.GetUserAvatar)
	route.Get("/avatar/:id", controllers.GetAvatarById)
	route.Delete("/avatar", middleware.Auth(), middleware.ValidateJwt(), controllers.DeleteUserAvatar)
	route.Post("/change/password", middleware.Auth(), middleware.ValidateJwt(), controllers.ChangeUserPassword)
}
