package routes

import (
	"vehicle-api/controllers"

	"github.com/gofiber/fiber/v2"
)

func ValuationRoutes(app *fiber.App) {
	app.Get("/api/v1/valuation", controllers.ValuationController)
}
