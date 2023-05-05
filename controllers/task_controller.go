package controllers

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/roshanpaturkar/go-tasks/database"
	"github.com/roshanpaturkar/go-tasks/models"
	"github.com/roshanpaturkar/go-tasks/utils"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		"error":   false,
		"message": "Task created successfully",
		"task":    res.InsertedID,
	})
}

func GetTasks(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	db := database.MongoClient()

	var tasks []models.Task

	collection := db.Collection("tasks")
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := collection.Find(c.Context(), fiber.Map{"user_id": user.ID}, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"error":   err.Error(),
		})
	}

	if err := cursor.All(c.Context(), &tasks); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"error":   err.Error(),
		})
	}

	jsonTasks, err := json.Marshal(tasks)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"error":   err.Error(),
		})
	}

	var tasksResponse []models.GetTask
	if err := json.Unmarshal([]byte(jsonTasks), &tasksResponse); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
		"tasks": tasksResponse,
	})
}

func GetTask(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	db := database.MongoClient()

	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
			"error":   true,
		})
	}

	task := new(models.Task)

	collection := db.Collection("tasks")
	if err := collection.FindOne(c.Context(), fiber.Map{"_id": id, "user_id": user.ID}).Decode(&task); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Task not found",
			"error":   true,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
		"task":  models.GetTask{
			ID:        task.ID.Hex(),
			Title:     task.Title,
			Completed: task.Completed,
			Metadata:  task.Metadata,
			CreatedAt: task.CreatedAt,
			UpdatedAt: task.UpdatedAt,
		},
	})
}

func UpdateTask(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	db := database.MongoClient()
	var taskUpdate map[string]interface{}

	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
			"error":   true,
		})
	}

	if err := json.Unmarshal(c.Body(), &taskUpdate); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
			"error":   true,
		})
	}

	task := new(models.Task)
	collection := db.Collection("tasks")
	if err := collection.FindOne(c.Context(), fiber.Map{"_id": id, "user_id": user.ID}).Decode(&task); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Task not found",
			"error":   true,
		})
	}

	parsedTaskUpdate := utils.UpdateTaskParser(taskUpdate, task.Metadata)

	res, err := collection.UpdateOne(c.Context(), fiber.Map{"_id": id, "user_id": user.ID}, bson.M{"$set": parsedTaskUpdate})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"error":   true,
		})
	}

	if res.MatchedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Task not found",
			"error":   true,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
		"message": "Task updated successfully",
	})
}

func DeleteTask(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	db := database.MongoClient()

	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
			"error":   true,
		})
	}

	collection := db.Collection("tasks")
	res, err := collection.DeleteOne(c.Context(), fiber.Map{"_id": id, "user_id": user.ID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"error":   true,
		})
	}

	if res.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Task not found",
			"error":   true,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":   false,
		"message": "Task deleted successfully",
	})
}
