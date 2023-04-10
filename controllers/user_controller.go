package controllers

import (
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/roshanpaturkar/go-tasks/database"
	"github.com/roshanpaturkar/go-tasks/models"
	"github.com/roshanpaturkar/go-tasks/utils"
)

func UserSignUp(c *fiber.Ctx) error {
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

func UserSignIn(c *fiber.Ctx) error {
	validate := validator.New()
	signIn := new(models.SignIn)
	c.BodyParser(&signIn)

	if err := validate.Struct(signIn); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Find the user
	user := &models.User{}
	db := database.MongoClient()

	if err := db.Collection(os.Getenv("USER_COLLECTION")).FindOne(c.Context(), fiber.Map{"email": signIn.Email}).Decode(&user); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": true,
			"msg":   "Incorrect email or password",
		})
	}

	// Check if the password is correct
	if match := utils.CheckPasswordHash(signIn.Password, user.PasswordHash); !match {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": true,
			"msg":   "Incorrect email or password",
		})
	}

	// Create a token
	token, err := utils.GenerateNewToken(user.ID.Hex())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	// Save the token in the database
	if _, err := db.Collection(os.Getenv("USER_COLLECTION")).UpdateOne(c.Context(), fiber.Map{"_id": user.ID}, fiber.Map{"$set": fiber.Map{"tokens": []string(append(user.Tokens, token.Access)), "updated_at": time.Now().Unix()}}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
		"msg":   "User signed in successfully",
		"tokens": fiber.Map{
			"access": token.Access,
		},
	})
}

func UserSignOut(c *fiber.Ctx) error {
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

	for i, token := range tokens {
		if token == bearToken {
			tokens = append(tokens[:i], tokens[i+1:]...)
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

	// Remove the token from the database
	if _, err := db.Collection(os.Getenv("USER_COLLECTION")).UpdateOne(c.Context(), fiber.Map{"_id": user.ID}, fiber.Map{"$set": fiber.Map{"tokens": tokens, "updated_at": time.Now().Unix()}}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
		"msg":   "User signed out successfully",
	})
}

func UserSignOutAll(c *fiber.Ctx) error {
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

	// Remove the token from the database
	if _, err := db.Collection(os.Getenv("USER_COLLECTION")).UpdateOne(c.Context(), fiber.Map{"_id": user.ID}, fiber.Map{"$set": fiber.Map{"tokens": []string{}, "updated_at": time.Now().Unix()}}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": true,
			"msg":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
		"msg":   "User signed out successfully",
	})
}

func UserProfile(c *fiber.Ctx) error {
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

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
		"msg":   "User profile",
		"user": models.UserProfileResponse{
			ID:        user.ID.Hex(),
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Mobile:   user.Mobile,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	})
}
