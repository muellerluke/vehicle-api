package controllers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"vehicle-api/utils"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
)

//redis keys
/*
	years = all years
	[a year] = all makes for a year ex: 2021
	[a year]:[a make] = all models for a year and make ex: 2021:acura
	[a year]:[a make]:[a model] = all trims for a year, make, and model ex: 2021:acura:ilx
*/

func ValuationController(c *fiber.Ctx) error {

	var vin = c.Query("vin")
	//var zipCode = c.Query("zip_code")
	//var radius = c.Query("radius")

	response, err := http.Get("https://vpic.nhtsa.dot.gov/api/vehicles/decodevin/" + vin + "?format=json")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": "Something went wrong. Please try again later."}})
	}

	defer response.Body.Close()

	// Read the response body
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	// Create a struct to unmarshal the JSON response
	var decodedVin struct {
		Results []struct {
			Year   string `json:"ModelYear"`
			Make   string `json:"Make"`
			Model  string `json:"Model"`
			Series string `json:"Series"`
		} `json:"Results"`
	}

	// Unmarshal the JSON response into the struct
	if err := json.Unmarshal(responseBody, &decodedVin); err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": "Something went wrong. Please try again later."}})
	}

	// Check if there are results
	if len(decodedVin.Results) <= 0 {
		fmt.Println("No results found for the VIN.")
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": "No results found for the VIN."}})
	}

	result := decodedVin.Results[0]
	year := result.Year
	make := result.Make
	model := result.Model
	series := result.Series

	return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{"year": year, "make": make, "model": model, "series": series}})

}
