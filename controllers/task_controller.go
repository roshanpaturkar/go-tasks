package controllers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/roshanpaturkar/go-tasks/database"
	"github.com/roshanpaturkar/go-tasks/models"
)

func CreateTask(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	db := database.MongoClient()

	task := new(models.Task)
	if err := c.BodyParser(task); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
			"error":   err.Error(),
		})
	}

	timestamp := time.Now().Unix()

	task.UserId = user.ID
	task.CreatedAt = timestamp	
	task.UpdatedAt = timestamp

	collection := db.Collection("tasks")
	res, err := collection.InsertOne(c.Context(), task)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error": false,
		"message": "Task created successfully",
		"task":    res.InsertedID,
	})
}