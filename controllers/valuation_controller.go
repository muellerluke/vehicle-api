package controllers

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"vehicle-api/utils"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"

	"github.com/sajari/regression"
	"golang.org/x/net/html"
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
	var zipCode = c.Query("zip_code")
	var radius = c.Query("radius")
	var mileage = c.Query("mileage")
	var multipleYears = c.Query("multiple_years")

	responseVin, err := http.Get("https://vpic.nhtsa.dot.gov/api/vehicles/decodevin/" + vin + "?format=json")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": "Something went wrong. Please try again later."}})
	}

	defer responseVin.Body.Close()

	// Read the response body
	responseBody, err := ioutil.ReadAll(responseVin.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	// Create a struct to unmarshal the JSON response
	var decodedVin struct {
		Results []struct {
			Variable   string `json:"Variable"`
			Value      string `json:"Value"`
			VariableId int    `json:"VariableId"`
			ValueId    string `json:"ValueId"`
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

	var year string
	var make string
	var model string

	for _, result := range decodedVin.Results {
		if result.Variable == "Model Year" {
			year = result.Value
		}
		if result.Variable == "Make" {
			make = result.Value
		}
		if result.Variable == "Model" {
			model = result.Value
		}
		/*if result.Variable == "Series" {
			series = result.Value
		}*/
	}

	if year == "" || make == "" || model == "" {
		fmt.Println("No results found for the VIN.")
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": "No results found for the VIN."}})
	}

	//get valuation
	var url string
	var multipleYearsString string = year + "/"
	if multipleYears == "true" {
		multipleYearsString = ""
	}

	if radius == "" {
		url = "https://www.autotrader.com/cars-for-sale/all-cars/" + multipleYearsString + strings.ReplaceAll(strings.ToLower(make), " ", "-") + "/" + strings.ReplaceAll(strings.ToLower(model), " ", "-") + "?numRecords=100&searchRadius=100&zip=" + zipCode
	} else {
		url = "https://www.autotrader.com/cars-for-sale/all-cars/" + multipleYearsString + strings.ReplaceAll(strings.ToLower(make), " ", "-") + "/" + strings.ReplaceAll(strings.ToLower(model), " ", "-") + "?numRecords=100&searchRadius=" + radius + "&zip=" + zipCode
	}

	fmt.Println(url)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	response, _ := client.Do(req)

	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": "Something went wrong. Please try again later."}})
	}
	defer response.Body.Close()

	//tokenize the response html
	z := html.NewTokenizer(response.Body)

	isProductElement := false
	isPriceElement := false
	isMileageElement := false
	isListingTitleElement := false
	prices := []int{}
	mileages := []int{}
	listingTitles := []string{}

	//loop through the tokens
	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done

			//remove all listings that are new vehicles
			years := []int{}

			for i, listingTitle := range listingTitles {
				if strings.Contains(listingTitle, "New") {
					prices = append(prices[:i], prices[i+1:]...)
					mileages = append(mileages[:i], mileages[i+1:]...)
				} else {
					// Define a regular expression pattern to match four consecutive numbers
					pattern := `[0-9]{4}`

					// Compile the regular expression
					regexpPattern := regexp.MustCompile(pattern)

					// Find all matches in the input string
					matches := regexpPattern.FindAllString(listingTitle, -1)

					if len(matches) == 0 {
						return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": "Something went wrong. Please try again later."}})
					}

					yearAsInt, err := strconv.Atoi(matches[0])
					if err != nil {
						fmt.Println("Error converting year to int")
						return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
					}

					years = append(years, yearAsInt)
				}
			}

			//check that all maps are the same length
			if len(prices) != len(mileages) || len(prices) != len(years) {
				fmt.Println("Error: prices, mileages, and years are not the same length")
				fmt.Println(strconv.Itoa(len(prices))+" prices: ", prices)
				fmt.Println(strconv.Itoa(len(mileages))+" mileages: ", mileages)
				fmt.Println(strconv.Itoa(len(years))+" years: ", years)
				return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": "Something went wrong. Please try again later."}})
			}

			if len(prices) < 2 {
				fmt.Println("Error: not enough results.")
				return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": "Not enough results. Please expand the search radius and try querying with multipleYears=true"}})
			}

			r := new(regression.Regression)
			r.SetObserved("Price")
			r.SetVar(0, "Mileage")
			r.SetVar(1, "Year")

			for i, price := range prices {
				r.Train(regression.DataPoint(float64(price), []float64{float64(mileages[i]), float64(years[i])}))
			}
			r.Run()

			fmt.Printf("Regression formula:\n%v\n", r.Formula)
			fmt.Printf("Regression:\n%s\n", r)

			//return valuation
			yearInt, err := strconv.Atoi(year)
			if err != nil {
				fmt.Println("Error converting year to int")
				return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
			}

			var mileageInt int

			if mileage == "" {
				//get average mileage for specific year
				var totalMileage int
				var totalRecords int

				for i, year := range years {
					if year == yearInt {
						totalMileage += mileages[i]
						totalRecords++
					}
				}

				if totalRecords == 0 {
					//get average mileage for all years
					for _, mileage := range mileages {
						totalMileage += mileage
					}
					totalRecords = len(mileages)
				}

				mileageInt = totalMileage / totalRecords

			} else {
				mileageInt, err = strconv.Atoi(mileage)
				if err != nil {
					fmt.Println("Error converting mileage to int")
					return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
				}
			}

			prediction, err := r.Predict([]float64{float64(mileageInt), float64(yearInt)})

			if err != nil {
				fmt.Println("Error predicting price")
			}

			return c.Status(http.StatusOK).JSON(utils.ApiResponse{Status: http.StatusOK, Message: "success", Data: &fiber.Map{
				"predicted_price": math.Floor(prediction*100) / 100,
				"based_on":        strconv.Itoa(len(prices)) + " results",
				"mileage":         strconv.Itoa(mileageInt),
				"year":            year,
				"make":            make,
				"model":           model,
			}})

		case tt == html.StartTagToken:
			t := z.Token()

			if t.Type == html.StartTagToken && t.Data == "div" {
				for _, a := range t.Attr {
					if a.Key == "class" && strings.Contains(a.Val, "item-card") {
						isProductElement = true
					}
				}
			}

			if t.Type == html.StartTagToken && t.Data == "span" {
				for _, a := range t.Attr {
					if a.Key == "class" && strings.Contains(a.Val, "first-price") {
						isPriceElement = true
					}
					if a.Key == "class" && strings.Contains(a.Val, "text-bold") {
						isMileageElement = true
					}
				}
			}

			if t.Type == html.StartTagToken && t.Data == "h3" {
				for _, a := range t.Attr {
					if a.Key == "class" && strings.Contains(a.Val, "text-bold") {
						isListingTitleElement = true
					}
				}
			}
		case tt == html.TextToken:
			t := z.Token()

			if isProductElement {
				if isMileageElement && strings.Contains(t.Data, " miles") && len(t.Data) < 15 && len(t.Data) > 0 {
					mileage, err := strconv.Atoi(strings.ReplaceAll(strings.ReplaceAll(t.Data, ",", ""), " miles", ""))
					if err != nil {
						fmt.Println("Error converting mileage to int")
						return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
					}
					mileages = append(mileages, mileage)
					isMileageElement = false
				}

				if isPriceElement {
					//check that current element has both mileage and title
					if len(mileages) < len(listingTitles) {
						//remove previous listing title
						listingTitles = listingTitles[:len(listingTitles)-1]
						isProductElement = false
					}

					//check that it's not dealer price - if it is, there will be an extra two titles and mileages
					//need to remove the second to last title and mileage
					if len(listingTitles) > len(prices)+1 && len(mileages) == len(listingTitles) {
						indexToRemove := len(listingTitles) - 2

						// Remove the second-to-last element by slicing the slice
						listingTitles = append(listingTitles[:indexToRemove], listingTitles[indexToRemove+1:]...)
						mileages = append(mileages[:indexToRemove], mileages[indexToRemove+1:]...)
					}

					price, err := strconv.Atoi(strings.ReplaceAll(t.Data, ",", ""))
					if err != nil {
						fmt.Println("Error converting price to int")
						return c.Status(http.StatusInternalServerError).JSON(utils.ApiResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
					}
					prices = append(prices, price)
					isProductElement = false
					isPriceElement = false
				}

				if isListingTitleElement && len(t.Data) > 0 && len(t.Data) < 250 {
					listingTitle := t.Data
					listingTitles = append(listingTitles, listingTitle)
					isListingTitleElement = false
				}
			}
		}
	}
}
