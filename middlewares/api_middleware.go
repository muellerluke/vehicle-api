package middlewares

import (
	"github.com/gofiber/fiber/v2"
)

func ApiMiddleware(c *fiber.Ctx) error {
	apiKey := c.Get("X-API-KEY")

	// check if apiKey is not undefined
	// if not undefined, check if apiKey is valid
	if apiKey == "" {
		return c.Status(401).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	return c.Next()
}
