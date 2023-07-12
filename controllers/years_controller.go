package controllers

import (
	"net/http"
	"vehicle-api/responses"

	"github.com/gofiber/fiber/v2"
)

func YearsController(c *fiber.Ctx) error {

	return c.Status(http.StatusOK).JSON(responses.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"signed_url": "hello"}})
}
