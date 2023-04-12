package utils

import (
	"bytes"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func SetAvatarHeaders(c *fiber.Ctx, buff bytes.Buffer, ext string) error {

	switch ext {
		case ".png": c.Set("Content-Type", "image/png")
		case ".jpg": c.Set("Content-Type", "image/jpeg")
		case ".jpeg": c.Set("Content-Type", "image/jpeg")
	}
	
	c.Set("Cache-Control", "public, max-age=31536000")
	c.Set("Content-Length", strconv.Itoa(len(buff.Bytes())))

	return c.Next()
}