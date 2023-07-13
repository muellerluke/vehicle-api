package controllers

import (
	"context"
	"net/http"
	"time"
	"vehicle-api/configs"
	"vehicle-api/models"
	"vehicle-api/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/customer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = configs.GetCollection(configs.DB, "users")
var validate = validator.New()

func Register(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var user models.User
	defer cancel()

	//validate the request body
	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	//use the validator library to validate required fields
	if validationErr := validate.Struct(&user); validationErr != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": validationErr.Error()}})
	}

	//verify user doesn't already exist
	var existingUser models.User
	if err := userCollection.FindOne(ctx, models.User{Email: user.Email}).Decode(&existingUser); err == nil {
		return c.Status(http.StatusBadRequest).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": "User already exists"}})
	}

	//verify password is at least 8 characters, has a number, a capital letter, and has a special character
	isValidPassword := utils.ValidatePassword(user.Password)
	if !isValidPassword {
		return c.Status(http.StatusBadRequest).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": "Password must be at least 8 characters, have a number, a capital letter, and a special character"}})
	}

	//hash the password
	bytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), 14)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	//all fields are entered correctly, create the user in stripe, then in our database
	stripe.Key = configs.RetrieveEnv("STRIPE_SECRET_KEY")

	params := &stripe.CustomerParams{
		Email: stripe.String(user.Email),
		Name:  stripe.String(user.FirstName + " " + user.LastName),
	}
	customer, _ := customer.New(params)

	newUser := models.User{
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		Email:            user.Email,
		Password:         string(bytes),
		Keys:             []primitive.ObjectID{},
		CreatedAt:        time.Now().Unix(),
		IsVerified:       false,
		IsActive:         false,
		StripeCustomerID: customer.ID,
	}

	result, err := userCollection.InsertOne(ctx, newUser)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	//retrieve session from fiber
	store, err := configs.GetSession().Get(c)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	//set the session values
	store.Set("user", newUser)

	//need to send user email to verify account

	return c.Status(http.StatusCreated).JSON(utils.ApiResponse{Status: http.StatusCreated, Message: "success", Data: &fiber.Map{"data": result}})
}

func Login(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var user models.User
	defer cancel()

	//validate the request body
	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	//validate that the request body includes email and password
	if user.Email == "" || user.Password == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": "Email and password are required"}})
	}

	//verify user exists
	var existingUser models.User
	if err := userCollection.FindOne(ctx, models.User{Email: user.Email}).Decode(&existingUser); err != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": "User does not exist"}})
	}

	//verify password is correct
	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password)); err != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": "Email and password do not match"}})
	}

	//retrieve session from fiber
	store, err := configs.GetSession().Get(c)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	//set the session values
	store.Set("user", existingUser)

	return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"data": existingUser}})
}

func Logout(c *fiber.Ctx) error {
	//retrieve session from fiber
	store, err := configs.GetSession().Get(c)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	//delete the session
	store.Destroy()

	return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"signed_url": "hello"}})
}

func ForgotPassword(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//validate the request body
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	//validate that the request body includes email
	if user.Email == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": "Email is required"}})
	}

	//verify user exists
	var existingUser models.User
	if err := userCollection.FindOne(ctx, models.User{Email: user.Email}).Decode(&existingUser); err != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": "User does not exist"}})
	}

	//generate a random string
	randomString := utils.GenerateRandomString(32)

	//update the user with the random string and expiry a day from now
	dayFromNow := time.Now().AddDate(0, 0, 1).Unix()
	if err := userCollection.FindOneAndUpdate(ctx, models.User{Email: user.Email}, bson.M{"$set": bson.M{"reset_token": randomString, "reset_token_expiry": dayFromNow}}).Decode(&existingUser); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"signed_url": "hello"}})
}

func ResetPassword(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//validate the request body
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	//validate that the request body includes email, password and reset token
	if user.Email == "" || user.Password == "" || user.ResetToken == "" {
		return c.Status(http.StatusBadRequest).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": "Email, password and reset token are required"}})
	}

	//verify user exists
	var existingUser models.User
	if err := userCollection.FindOne(ctx, models.User{Email: user.Email}).Decode(&existingUser); err != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": "User does not exist"}})
	}

	//verify reset token is valid
	if existingUser.ResetToken != user.ResetToken || existingUser.ResetTokenExpiry < time.Now().Unix() {
		return c.Status(http.StatusBadRequest).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": "Reset token is invalid"}})
	}

	//hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	//update the user with the new password and remove the reset token
	if err := userCollection.FindOneAndUpdate(ctx, models.User{Email: user.Email}, bson.M{"$set": bson.M{"password": string(hashedPassword), "reset_token": nil, "reset_token_expiry": nil}}).Decode(&existingUser); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"data": "Password reset successful"}})
}

func ChangePassword(c *fiber.Ctx) error {
	payload := struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//retrieve session from fiber
	store, err := configs.GetSession().Get(c)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	//retrieve user from session
	user := store.Get("user")
	if user == nil {
		return c.Status(http.StatusUnauthorized).JSON(utils.ApiResponse{Status: http.StatusUnauthorized, Message: "error", Data: &fiber.Map{"data": "Unauthorized"}})
	}

	var existingUser models.User
	if err := userCollection.FindOne(ctx, models.User{Email: user.(models.User).Email}).Decode(&existingUser); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	//validate the request body
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	//verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(payload.OldPassword)); err != nil {
		return c.Status(http.StatusBadRequest).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": "Old password is incorrect"}})
	}

	//validate new password
	passwordValid := utils.ValidatePassword(payload.NewPassword)
	if !passwordValid {
		return c.Status(http.StatusBadRequest).JSON(utils.ApiResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": "Password must be at least 8 characters and contain at least one uppercase letter, one lowercase letter, one number and one special character"}})
	}

	//hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), 10)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	//update the user with the new password
	if err := userCollection.FindOneAndUpdate(ctx, models.User{Email: user.(models.User).Email}, bson.M{"$set": bson.M{"password": string(hashedPassword)}}).Decode(&existingUser); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"signed_url": "hello"}})
}

func UpdateProfile(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"signed_url": "hello"}})

}

func GetProfile(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"signed_url": "hello"}})

}
