package main

import (
	"vehicle-api/configs"
	"vehicle-api/middlewares"
	"vehicle-api/routes"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	//"github.com/gofiber/template/html/v2"
)

func main() {
	//create html engine
	//engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		BodyLimit: 1024 * 1024 * 10,
		//Views:       engine,
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	//run database
	configs.ConnectDB()

	//middlewares
	app.Use(logger.New())
	app.Get("/metrics", monitor.New())

	//create groups for middleware

	//api/v1 = public api routes for customers
	api := app.Group("/api/v1")

	//autofill routes (years, makes, models, trims) should use autofill middleware
	autofillRoutes := api.Group("/autofill")
	autofillRoutes.Use(middlewares.AutofillMiddleware)
	routes.AutofillRoutes(app)

	//admin-api = admin api routes for staff admins
	adminApi := app.Group("/admin-api")
	adminApi.Use(middlewares.AdminMiddleware)
	routes.AdminRoutes(app)

	app.Listen(":" + configs.RetrieveEnv("PORT"))
}
