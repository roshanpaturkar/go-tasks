package controllers

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/roshanpaturkar/go-tasks/models"
	"github.com/roshanpaturkar/go-tasks/utils"
)

func UserSignUp(c *fiber.Ctx) error {
	validate := validator.New()
	signUp := new(models.SignUp)
	c.BodyParser(&signUp)

	if err := validate.Struct(signUp); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":	true,
			"message":	err.Error(),
		})
	}

	passwdHash, err := utils.HashPassword(signUp.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":	true,
			"message":  err.Error(),
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
			"error":	true,
			"message":	err.Error(),
		})
	}

	db := c.Locals("db").(*mongo.Database)

	// Check if the user already exists
	if err := db.Collection(os.Getenv("USER_COLLECTION")).FindOne(c.Context(), fiber.Map{"email": user.Email}).Decode(&user); err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":	true,
			"message":	"User already exists",
		})
	}

	// Insert the user
	if _, err := db.Collection(os.Getenv("USER_COLLECTION")).InsertOne(c.Context(), user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":	true,
			"message":	err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error":	false,
		"message":	"User created successfully",
	})
}

func UserSignIn(c *fiber.Ctx) error {
	validate := validator.New()
	signIn := new(models.SignIn)
	c.BodyParser(&signIn)

	if err := validate.Struct(signIn); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":	true,
			"message":	err.Error(),
		})
	}

	// Find the user
	user := &models.User{}
	db := c.Locals("db").(*mongo.Database)

	if err := db.Collection(os.Getenv("USER_COLLECTION")).FindOne(c.Context(), fiber.Map{"email": signIn.Email}).Decode(&user); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":	true,
			"message":	"Incorrect email or password",
		})
	}

	// Check if the password is correct
	if match := utils.CheckPasswordHash(signIn.Password, user.PasswordHash); !match {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":	true,
			"message":	"Incorrect email or password",
		})
	}

	// Create a token
	token, err := utils.GenerateNewToken(user.ID.Hex())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":	true,
			"message":	err.Error(),
		})
	}

	// Save the token in the database
	if _, err := db.Collection(os.Getenv("USER_COLLECTION")).UpdateOne(c.Context(), fiber.Map{"_id": user.ID}, fiber.Map{"$set": fiber.Map{"tokens": []string(append(user.Tokens, token.Access)), "updated_at": time.Now().Unix()}}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":	true,
			"message":	err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":	false,
		"message":	"User signed in successfully",
		"tokens":	fiber.Map{
			"access": token.Access,
		},
	})
}

func UserSignOut(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	db := c.Locals("db").(*mongo.Database)

	bearToken := strings.Split(c.Get("Authorization"), " ")[1]
	tokens := user.Tokens

	for i, token := range tokens {
		if token == bearToken {
			tokens = append(tokens[:i], tokens[i+1:]...)
			break
		}
	}

	// Remove the token from the database
	if _, err := db.Collection(os.Getenv("USER_COLLECTION")).UpdateOne(c.Context(), fiber.Map{"_id": user.ID}, fiber.Map{"$set": fiber.Map{"tokens": tokens, "updated_at": time.Now().Unix()}}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":	true,
			"message":	"Internal server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":	false,
		"message":	"User signed out successfully",
	})
}

func UserSignOutAll(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	db := c.Locals("db").(*mongo.Database)

	// Remove the token from the database
	if _, err := db.Collection(os.Getenv("USER_COLLECTION")).UpdateOne(c.Context(), fiber.Map{"_id": user.ID}, fiber.Map{"$set": fiber.Map{"tokens": []string{}, "updated_at": time.Now().Unix()}}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":	true,
			"message":	"Internal server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":	false,
		"message":	"User signed out successfully",
	})
}

func UserProfile(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":	false,
		"message":	"User profile",
		"user":		models.UserProfileResponse{
			ID:			user.ID.Hex(),
			FirstName:	user.FirstName,
			LastName:	user.LastName,
			Email:		user.Email,
			Mobile:		user.Mobile,
			Avatar:		"/api/v1/user/avatar/" + user.ID.Hex(),
			CreatedAt:	user.CreatedAt,
			UpdatedAt:	user.UpdatedAt,
		},
	})
}

func UploadUserAvatar(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	db := c.Locals("db").(*mongo.Database)

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":	true,
			"message":	err.Error(),
		})
	}

	if (fileHeader.Size) > 1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":	true,
			"message":	"File size too large, max 1MB allowed",
		})
	}

	fileExtension := strings.ToLower(fileHeader.Filename[strings.LastIndex(fileHeader.Filename, "."):])

	if fileExtension != ".jpg" && fileExtension != ".jpeg" && fileExtension != ".png" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":	true,
			"message":	"Invalid file type",
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":	true,
			"message":	"Internal server error",
		})
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":	true,
			"message":	"Internal server error",
		})
	}

	bucket, err := gridfs.NewBucket(db, options.GridFSBucket().SetName(os.Getenv("AVATAR_BUCKET")))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":	true,
			"message":	"Internal server error",
		})
	}

	var avatarMetadata bson.M

	if err := db.Collection(os.Getenv("AVATAR_COLLECTION")).FindOne(c.Context(), fiber.Map{"metadata.user_id": user.ID}).Decode(&avatarMetadata); err == nil {
		// Delete existing avatar file
		if err:= bucket.Delete(user.ID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":	true,
				"message":	"Internal server error",
			})
		}
	}

	uploadStream, err := bucket.OpenUploadStream(fileHeader.Filename, options.GridFSUpload().SetMetadata(fiber.Map{
		"user_id": user.ID,
		"ext":     fileExtension,
	}))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":	true,
			"message":	"Internal server error",
		})
	}

	uploadStream.FileID = user.ID
	defer uploadStream.Close()

	fileSize, err := uploadStream.Write(content)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":	true,
			"message":	"Internal server error",
		})
	}

	log.Printf("Write file to DB was successful. File size: %d KB\n", fileSize/1024)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":	false,
		"message":	"Avatar uploaded successfully",
	})
}

func GetUserAvatar(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	db := c.Locals("db").(*mongo.Database)

	var avatarMetadata bson.M

	if err := db.Collection(os.Getenv("AVATAR_COLLECTION")).FindOne(c.Context(), fiber.Map{"metadata.user_id": user.ID}).Decode(&avatarMetadata); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":	true,
			"message":	"Avatar not found",
		})
	}

	bucket, _ := gridfs.NewBucket( db, options.GridFSBucket().SetName(os.Getenv("AVATAR_BUCKET")))

	var buffer bytes.Buffer
	bucket.DownloadToStream(user.ID, &buffer)

	utils.SetAvatarHeaders(c, buffer, avatarMetadata["metadata"].(bson.M)["ext"].(string))

	return c.Send(buffer.Bytes())
}

func GetAvatarById(c *fiber.Ctx) error {
	userID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":	true,
			"message":	"Invalid user ID",
		})
	}

	var avatarMetadata bson.M

	db := c.Locals("db").(*mongo.Database)

	if err := db.Collection(os.Getenv("AVATAR_COLLECTION")).FindOne(c.Context(), fiber.Map{"metadata.user_id": userID}).Decode(&avatarMetadata); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":	true,
			"message":	"Avatar not found",
		})
	}

	bucket, _ := gridfs.NewBucket( db, options.GridFSBucket().SetName(os.Getenv("AVATAR_BUCKET")))

	var buffer bytes.Buffer
	bucket.DownloadToStream(userID, &buffer)

	utils.SetAvatarHeaders(c, buffer, avatarMetadata["metadata"].(bson.M)["ext"].(string))

	return c.Send(buffer.Bytes())
}

func DeleteUserAvatar(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	db := c.Locals("db").(*mongo.Database)

	var avatarMetadata bson.M

	if err := db.Collection(os.Getenv("AVATAR_COLLECTION")).FindOne(c.Context(), fiber.Map{"metadata.user_id": user.ID}).Decode(&avatarMetadata); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":	true,
			"message":	"Avatar not found",
		})
	}

	bucket, _ := gridfs.NewBucket( db, options.GridFSBucket().SetName(os.Getenv("AVATAR_BUCKET")))

	if err := bucket.Delete(user.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":	true,
			"message":	"Internal server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":	false,
		"message":	"Avatar deleted successfully",
	})
}

func ChangeUserPassword(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	db := c.Locals("db").(*mongo.Database)
	validate := validator.New()

	userPasswords := new(models.ChangeUserPassword)

	if err := c.BodyParser(&userPasswords); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":	true,
			"message":	"Invalid request body",
		})
	}

	// Validate the user
	if err := validate.Struct(userPasswords); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":	true,
			"message":	err.Error(),
		})
	}

	// Check if the password is correct
	if match := utils.CheckPasswordHash(userPasswords.OldPassword, user.PasswordHash); !match {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":	true,
			"message":	"Incorrect password",
		})
	}

	// Hash the new password
	passwdHash, err := utils.HashPassword(userPasswords.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":	true,
			"message":  err.Error(),
		})
	}

	user.PasswordHash = passwdHash
	user.UpdatedAt = time.Now().Unix()

	if _, err := db.Collection(os.Getenv("USER_COLLECTION")).UpdateOne(c.Context(), fiber.Map{"_id": user.ID}, fiber.Map{"$set": user}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":	true,
			"message":	"Internal server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":	false,
		"message":	"Password changed successfully",
	})
}
