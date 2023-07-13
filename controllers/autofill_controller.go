package controllers

import (
	"net/http"
	"vehicle-api/utils"

	"github.com/gofiber/fiber/v2"
)

func MakesController(c *fiber.Ctx) error {

	return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"signed_url": "hello"}})
}

func ModelsController(c *fiber.Ctx) error {

	return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"signed_url": "hello"}})
}

func TrimsController(c *fiber.Ctx) error {

	return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"signed_url": "hello"}})
}

func YearsController(c *fiber.Ctx) error {

	return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"signed_url": "hello"}})
}
