package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

func IngestDb(db *mongo.Database) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Locals("db", db)
		return c.Next()
	}
}