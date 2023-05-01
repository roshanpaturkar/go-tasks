package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/roshanpaturkar/go-tasks/database"
	"github.com/roshanpaturkar/go-tasks/models"
	"github.com/roshanpaturkar/go-tasks/utils"
)

func ValidateJwt() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		user := &models.User{}
		
		claims, err := utils.ExtractTokenMetadata(c)
		if err != nil {
			// Return status 500 and JWT parse error.
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}

		if claims.Expires < time.Now().Unix() {
			// Return status 401 and JWT expired error.
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": true,
				"msg":   "Token expired",
			})
		}

		db := database.MongoClient()

		if err := db.Collection(os.Getenv("USER_COLLECTION")).FindOne(c.Context(), fiber.Map{"_id": claims.UserID}).Decode(&user); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": true,
				"msg":   "Invalid token",
			})
		}

		bearToken := strings.Split(c.Get("Authorization"), " ")[1]
		tokens := user.Tokens
		tokenExists := false

		for _, token := range tokens {
			if token == bearToken {
				tokenExists = true
				break
			}
		}

		if !tokenExists {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": true,
				"msg":   "Token does not exist",
			})
		}

		c.Locals("user", user)
		return c.Next()
	}
}
