package middlewares

import (
	"context"
	"net/http"
	"time"
	"vehicle-api/configs"
	"vehicle-api/models"
	"vehicle-api/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/slices"
)

var keyCollection *mongo.Collection = configs.GetCollection(configs.DB, "keys")

func AutofillMiddleware(c *fiber.Ctx) error {
	host := c.Hostname()
	path := c.Path()
	keyString := c.Query("key")
	var route string = "years"

	//verify request has key
	if keyString == "" {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusUnauthorized, Message: "error", Data: &fiber.Map{"data": "Key is required"}})
	}

	//verify query params for each path
	switch path {
	case "/api/v1/makes":
		year := c.Query("year")
		if year == "" {
			return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": "Year is required"}})
		}
		route = "makes"
	case "/api/v1/models":
		year := c.Query("year")
		make := c.Query("make")
		if year == "" || make == "" {
			return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": "Year and make are required"}})
		}
		route = "models"
	case "/api/v1/trims":
		year := c.Query("year")
		make := c.Query("make")
		model := c.Query("model")
		if year == "" || make == "" || model == "" {
			return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": "Year, make, and model are required"}})
		}
		route = "trims"
	default:
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": "Invalid path"}})
	}

	//verify key for each host
	//don't need to verify organization is active bc keys are set to inactive when organization is set to inactive
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var key models.Key
	defer cancel()

	err := keyCollection.FindOne(ctx, bson.M{"key": keyString, "is_active": true, "routes": route}).Decode(&key)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusUnauthorized, Message: "error", Data: &fiber.Map{"data": "Invalid key"}})
		}
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": "Something went wrong. Please try again later."}})
	}

	// if key has a list of authorized domains and the host is not in the list, return unauthorized, else continue
	if len(key.AuthorizedDomains) > 0 && slices.Contains(key.AuthorizedDomains, host) == false {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusUnauthorized, Message: "error", Data: &fiber.Map{"data": "Invalid key"}})
	}

	// log call with logger function
	go utils.LogCall(key, c.OriginalURL(), route)

	return c.Next()
}
