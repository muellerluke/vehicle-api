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

func ValuationMiddleware(c *fiber.Ctx) error {
	host := c.Hostname()
	keyString := c.Query("key")
	// get X-RapidAPI-Proxy-Secret header from request
	rapidAPI := c.Get("X-RapidAPI-Proxy-Secret")

	//verify request has key
	if keyString == "" && rapidAPI == "" {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusUnauthorized, Message: "error", Data: &fiber.Map{"data": "Key is required"}})
	}

	//verify query params exist
	vin := c.Query("vin")
	if vin == "" {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": "VIN is required"}})
	}

	zipCode := c.Query("zip_code")
	if zipCode == "" {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": "Zip code is required"}})
	}

	radius := c.Query("radius")
	if radius == "" {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": "Radius is required"}})
	}

	if rapidAPI == configs.RetrieveEnv("RAPID_API_SECRET") {
		return c.Next()
	}

	//verify key for each host
	//don't need to verify organization is active bc keys are set to inactive when organization is set to inactive
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var key models.Key
	defer cancel()

	err := keyCollection.FindOne(ctx, bson.M{"key": keyString, "is_active": true, "routes": "valuation"}).Decode(&key)

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
	go utils.LogCall(key, c.OriginalURL(), "valuation")

	return c.Next()
}
