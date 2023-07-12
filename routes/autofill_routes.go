package routes

import (
	"vehicle-api/controllers"

	"github.com/gofiber/fiber/v2"
)

func AutofillRoutes(app *fiber.App) {
	app.Get("/api/v1/autofill/years", controllers.YearsController)
	app.Get("/api/v1/autofill/makes", controllers.MakesController)
	app.Get("/api/v1/autofill/models", controllers.ModelsController)
	app.Get("/api/v1/autofill/trims", controllers.TrimsController)
}
