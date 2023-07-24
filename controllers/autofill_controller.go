package controllers

import (
	"context"
	"net/http"
	"strings"
	"time"
	"vehicle-api/configs"
	"vehicle-api/models"
	"vehicle-api/utils"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//redis keys
/*
	years = all years
	[a year] = all makes for a year ex: 2021
	[a year]:[a make] = all models for a year and make ex: 2021:acura
	[a year]:[a make]:[a model] = all trims for a year, make, and model ex: 2021:acura:ilx
*/

var yearCollection *mongo.Collection = configs.GetCollection(configs.DB, "years")
var makeCollection *mongo.Collection = configs.GetCollection(configs.DB, "makes")
var modelCollection *mongo.Collection = configs.GetCollection(configs.DB, "models")
var trimCollection *mongo.Collection = configs.GetCollection(configs.DB, "trims")

func YearsController(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	val, err := configs.Redis.Get(ctx, "years").Result()

	if err == redis.Nil {
		//find within mongo db, save to redis, and return
		//find all years
		var years []models.Year
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		filter := bson.D{}
		opts := options.Find().SetProjection(bson.D{{"name", 1}})

		cursor, err := yearCollection.Find(ctx, filter, opts)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
		}

		if err = cursor.All(ctx, &years); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
		}

		yearsArray := make([]string, len(years))
		for i, year := range years {
			yearsArray[i] = year.Name
		}

		//marshal years
		marshalledYears, err := json.Marshal(yearsArray)

		//save to redis
		err = configs.Redis.Set(ctx, "years", marshalledYears, 0).Err()

		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
		}

		configs.Redis.Expire(ctx, "years", 24*7*time.Hour)

		return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"years": yearsArray}})
	}

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	var years []string
	err = json.Unmarshal([]byte(val), &years)

	return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"years": years}})

}

func MakesController(c *fiber.Ctx) error {
	var year string = c.Query("year")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	val, err := configs.Redis.Get(ctx, year).Result()

	if err == redis.Nil {
		//find within mongo db, save to redis, and return
		//find all makes for a year
		var makes []models.Make
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		filter := bson.D{{"year", year}}
		opts := options.Find().SetProjection(bson.D{{"name", 1}})
		cursor, err := makeCollection.Find(ctx, filter, opts)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
		}

		if err = cursor.All(ctx, &makes); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
		}

		makesArray := make([]string, len(makes))
		for i, make := range makes {
			makesArray[i] = make.Name
		}

		//marshal makes
		marshalledMakes, err := json.Marshal(makesArray)

		//save to redis
		err = configs.Redis.Set(ctx, year, marshalledMakes, 0).Err()

		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
		}

		configs.Redis.Expire(ctx, year, 24*7*time.Hour)

		return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"year": year, "makes": makesArray}})
	}

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	var makes []string
	err = json.Unmarshal([]byte(val), &makes)

	return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"year": year, "makes": makes}})
}

func ModelsController(c *fiber.Ctx) error {
	var foundMake models.Make

	var year string = c.Query("year")
	var makeQuery string = c.Query("make")

	makeQuery = strings.ToLower(strings.ReplaceAll(makeQuery, "-", " "))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	val, err := configs.Redis.Get(ctx, year+":"+makeQuery).Result()

	if err == redis.Nil {
		//find within mongo db, save to redis, and return
		//find all models for a year and make
		var foundModels []models.Model

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		filter := bson.D{{"year", year}, {"findable_name", makeQuery}}
		err := makeCollection.FindOne(ctx, filter).Decode(&foundMake)

		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
		}

		//create array of object ids
		var objectIds []primitive.ObjectID
		for _, model := range foundMake.Models {
			objectIds = append(objectIds, model)
		}

		opts := options.Find().SetProjection(bson.D{{"name", 1}})
		cursor, err := modelCollection.Find(ctx, bson.M{"_id": bson.M{"$in": objectIds}}, opts)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
		}

		if err = cursor.All(ctx, &foundModels); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
		}

		modelsArray := make([]string, len(foundModels))
		for i, model := range foundModels {
			modelsArray[i] = model.Name
		}

		//marshal models
		marshalledModels, err := json.Marshal(modelsArray)

		//save to redis
		err = configs.Redis.Set(ctx, year+":"+makeQuery, marshalledModels, 0).Err()

		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
		}

		configs.Redis.Expire(ctx, year+":"+makeQuery, 24*7*time.Hour)

		return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"year": year, "make": makeQuery, "models": modelsArray}})
	}

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	var models []string
	err = json.Unmarshal([]byte(val), &models)

	return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"year": year, "make": makeQuery, "models": models}})
}

func TrimsController(c *fiber.Ctx) error {
	var foundModel models.Model
	var foundMake models.Make

	var yearQuery string = c.Query("year")
	var makeQuery string = c.Query("make")
	var modelQuery string = c.Query("model")

	makeQuery = strings.ToLower(strings.ReplaceAll(makeQuery, "-", " "))
	modelQuery = strings.ToLower(strings.ReplaceAll(modelQuery, "-", " "))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	val, err := configs.Redis.Get(ctx, yearQuery+":"+makeQuery+":"+modelQuery).Result()

	if err == redis.Nil {
		//find within mongo db, save to redis, and return
		//find make for year first

		filter := bson.D{{"year", yearQuery}, {"findable_name", makeQuery}}
		err := makeCollection.FindOne(ctx, filter).Decode(&foundMake)

		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
		}

		//find model for make
		err = modelCollection.FindOne(ctx, bson.M{"_id": bson.M{"$in": foundMake.Models}, "findable_name": modelQuery}).Decode(&foundModel)

		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
		}

		//create array of object ids
		var objectIds []primitive.ObjectID
		for _, trim := range foundModel.Trims {
			objectIds = append(objectIds, trim)
		}

		opts := options.Find().SetProjection(bson.D{{"name", 1}})

		var foundTrims []models.Trim
		cursor, err := trimCollection.Find(ctx, bson.M{"_id": bson.M{"$in": objectIds}}, opts)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
		}

		if err = cursor.All(ctx, &foundTrims); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
		}

		trimsArray := make([]string, len(foundTrims))
		for i, trim := range foundTrims {
			trimsArray[i] = trim.Name
		}

		//marshal trims
		marshalledTrims, err := json.Marshal(trimsArray)

		//save to redis
		err = configs.Redis.Set(ctx, yearQuery+":"+makeQuery+":"+modelQuery, marshalledTrims, 0).Err()

		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
		}

		configs.Redis.Expire(ctx, yearQuery+":"+makeQuery+":"+modelQuery, 24*7*time.Hour)

		return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"year": yearQuery, "make": makeQuery, "model": modelQuery, "trims": trimsArray}})
	}

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	var trims []string
	err = json.Unmarshal([]byte(val), &trims)

	return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"year": yearQuery, "make": makeQuery, "model": modelQuery, "trims": trims}})
}
