package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	db "myapp/DB"
	"myapp/view"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

var errProductNotFound = errors.New("product not found")

type openFoodFactsProduct struct {
	Name    string
	Fat     float64
	Carb    float64
	Fiber   float64
	Protein float64
}

func registerScanRoutes(e *echo.Echo) {
	e.GET("/scan", scanView, validate)
	e.POST("/scan/:barcode", scanBarcode, validate)
}

func scanView(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	mealID, _ := strconv.Atoi(c.QueryParam("meal"))
	backURL := "/"
	if mealID != 0 {
		backURL = fmt.Sprintf("/meal/%d/", mealID)
	}
	nav := view.NavBack(userID, backURL, "Scan")
	scan := view.ScanPage(mealID)
	component := view.Full(nav, scan)
	return component.Render(context.Background(), c.Response().Writer)
}

func fetchOpenFoodFacts(barcode string) (openFoodFactsProduct, error) {
	resp, err := httpClient.Get("https://world.openfoodfacts.org/api/v2/product/" + barcode + ".json")
	if err != nil {
		return openFoodFactsProduct{}, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return openFoodFactsProduct{}, fmt.Errorf("API returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return openFoodFactsProduct{}, fmt.Errorf("failed to read response: %w", err)
	}
	var result struct {
		Status  int `json:"status"`
		Product struct {
			ProductName string `json:"product_name"`
			Nutriments  struct {
				Fat100g   float64 `json:"fat_100g"`
				Carbs100g float64 `json:"carbohydrates_100g"`
				Fiber100g float64 `json:"fiber_100g"`
				Prot100g  float64 `json:"proteins_100g"`
			} `json:"nutriments"`
		} `json:"product"`
	}
	if err := json.Unmarshal(body, &result); err != nil || result.Status != 1 {
		return openFoodFactsProduct{}, errProductNotFound
	}

	p := result.Product
	name := p.ProductName
	if name == "" {
		name = "Unknown (" + barcode + ")"
	}
	return openFoodFactsProduct{
		Name:    name,
		Fat:     p.Nutriments.Fat100g,
		Carb:    p.Nutriments.Carbs100g,
		Fiber:   p.Nutriments.Fiber100g,
		Protein: p.Nutriments.Prot100g,
	}, nil
}

func resolveBarcodedFood(barcode string, userID int) (int, error) {
	if f := db.FindFoodByBarcode(barcode); f != nil {
		return f.ID, nil
	}
	p, err := fetchOpenFoodFacts(barcode)
	if err != nil {
		return 0, err
	}
	return db.CreateFoodWithBarcode(p.Name, p.Fat, p.Carb, p.Fiber, p.Protein, 100, userID, barcode)
}

func scanBarcode(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	foodID, err := resolveBarcodedFood(c.Param("barcode"), userID)
	if err != nil {
		if errors.Is(err, errProductNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Product not found"})
		}
		return c.JSON(http.StatusBadGateway, map[string]string{"error": "Failed to fetch product"})
	}

	mealID, _ := strconv.Atoi(c.QueryParam("meal"))
	if mealID == 0 {
		mealID, err = db.CreateMeal(defaultMealName, userID, false)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
	}
	if err := db.CreateMealItem(mealID, foodID, 100, userID); err != nil {
		return handleDBErr(c, err)
	}

	c.Response().Header().Set("HX-Location", fmt.Sprint("/meal/", mealID, "/"))
	return c.NoContent(http.StatusOK)
}
