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
)

var userCollection *mongo.Collection = configs.GetCollection(configs.DB, "users")

func UserMiddleware(c *fiber.Ctx) error {
	//get user from session
	session, err := configs.GetSession().Get(c)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusUnauthorized, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	//retrieve user from session
	if session.Get("user") == nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusUnauthorized, Message: "error", Data: &fiber.Map{"data": "User not found"}})
	}

	var foundUser models.User = session.Get("user").(models.User)

	//get user from db
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var user models.User
	err = userCollection.FindOne(ctx, bson.M{"_id": foundUser.ID}).Decode(&user)
	defer cancel()

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusUnauthorized, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	return c.Next()
}
