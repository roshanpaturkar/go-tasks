package controllers

import (
	"encoding/json"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/roshanpaturkar/go-tasks/models"
	"github.com/roshanpaturkar/go-tasks/utils"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateTask(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	validate := validator.New()

	createTask := new(models.CreateTask)
	if err := c.BodyParser(&createTask); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	if err := validate.Struct(createTask); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":	true,
			"message":	err.Error(),
		})
	}

	task := new(models.Task)

	timestamp := time.Now().Unix()

	task.UserId = user.ID
	task.Title = createTask.Title
	task.Completed = createTask.Completed
	task.Metadata = createTask.Metadata
	task.CreatedAt = timestamp
	task.UpdatedAt = timestamp

	db := c.Locals("db").(*mongo.Database)
	res, err := db.Collection(os.Getenv("TASKS_COLLECTION")).InsertOne(c.Context(), task)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Internal Server Error",
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

	var tasks []models.Task

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	db := c.Locals("db").(*mongo.Database)
	cursor, err := db.Collection(os.Getenv("TASKS_COLLECTION")).Find(c.Context(), fiber.Map{"user_id": user.ID}, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Internal Server Error",
		})
	}

	if err := cursor.All(c.Context(), &tasks); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Internal Server Error",
		})
	}

	var tasksResponse []models.GetTask
	for _, task := range tasks {
		tasksResponse = append(tasksResponse, models.GetTask{
			ID:        task.ID.Hex(),
			Title:     task.Title,
			Completed: task.Completed,
			Metadata:  task.Metadata,
			CreatedAt: task.CreatedAt,
			UpdatedAt: task.UpdatedAt,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error": false,
		"tasks": tasksResponse,
	})
}

func GetTask(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid task ID",
		})
	}

	task := new(models.Task)

	db := c.Locals("db").(*mongo.Database)
	if err := db.Collection(os.Getenv("TASKS_COLLECTION")).FindOne(c.Context(), fiber.Map{"_id": id, "user_id": user.ID}).Decode(&task); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Task not found",
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
	var taskUpdate map[string]interface{}

	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid task ID",
		})
	}

	if err := json.Unmarshal(c.Body(), &taskUpdate); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	task := new(models.Task)

	db := c.Locals("db").(*mongo.Database)
	if err := db.Collection(os.Getenv("TASKS_COLLECTION")).FindOne(c.Context(), fiber.Map{"_id": id, "user_id": user.ID}).Decode(&task); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Task not found",
		})
	}

	parsedTaskUpdate := utils.UpdateTaskParser(taskUpdate, task.Metadata)

	res, err := db.Collection(os.Getenv("TASKS_COLLECTION")).UpdateOne(c.Context(), fiber.Map{"_id": id, "user_id": user.ID}, bson.M{"$set": parsedTaskUpdate})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Internal Server Error",
		})
	}

	if res.MatchedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Task not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":   false,
		"message": "Task updated successfully",
	})
}

func DeleteTask(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid task ID",
		})
	}

	db := c.Locals("db").(*mongo.Database)
	res, err := db.Collection(os.Getenv("TASKS_COLLECTION")).DeleteOne(c.Context(), fiber.Map{"_id": id, "user_id": user.ID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Internal Server Error",
		})
	}

	if res.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Task not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error":   false,
		"message": "Task deleted successfully",
	})
}
