package controllers

import (
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/roshanpaturkar/go-tasks/database"
	"github.com/roshanpaturkar/go-tasks/models"
	"github.com/roshanpaturkar/go-tasks/utils"
)

func CreateUser(c *fiber.Ctx) error {
	validate := validator.New()
	signUp := new(models.SignUp)
	c.BodyParser(&signUp)

	if err := validate.Struct(signUp); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	passwdHash, err := utils.HashPassword(signUp.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Create the user
	user := &models.User{}
	timeStamp := time.Now().Unix()

	user.FirstName = signUp.FirstName
	user.LastName = signUp.LastName
	user.Email = signUp.Email
	user.Mobile = signUp.Mobile
	user.PasswordHash = passwdHash
	user.CreatedAt = timeStamp
	user.UpdatedAt = timeStamp

	// Validate the user
	if err := validate.Struct(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	db := database.MongoClient()

	// Check if the user already exists
	if err := db.Collection(os.Getenv("USER_COLLECTION")).FindOne(c.Context(), fiber.Map{"email": user.Email}).Decode(&user); err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": true,
			"msg":   "User already exists",
		})
	}

	// Insert the user
	if _, err := db.Collection(os.Getenv("USER_COLLECTION")).InsertOne(c.Context(), user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error": false,
		"msg":   "User created successfully",
	})
}
