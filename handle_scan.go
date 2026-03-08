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
	"net/url"
	"strconv"
	"strings"
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
	e.GET("/scan/confirm", scanConfirmView, validate)
	e.POST("/scan/confirm", scanConfirmAdd, validate)
	e.POST("/scan", scanBarcodeManual, validate)
	e.POST("/scan/:barcode", scanBarcode, validate)
}

func scanView(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	mealID, _ := strconv.Atoi(c.QueryParam("meal"))
	mealType := c.QueryParam("type")
	backURL := "/"
	if mealID != 0 {
		backURL = editPathForType(mealID, mealType)
	} else {
		backURL = overviewPath(requestedMealDate(c))
	}
	nav := view.NavBack(userID, backURL, "Scan")
	scan := view.ScanPage(mealID, mealType, queryDateUnix(c), scanSearchPath(mealID, mealType, queryDateUnix(c)), c.QueryParam("error"))
	component := view.Full(nav, scan)
	return component.Render(context.Background(), c.Response().Writer)
}

func scanConfirmView(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	mealID, _ := strconv.Atoi(c.QueryParam("meal"))
	mealType := c.QueryParam("type")
	dateUnix := queryDateUnix(c)
	barcode := c.QueryParam("barcode")
	if barcode == "" {
		return c.Redirect(http.StatusSeeOther, "/scan")
	}

	food, err := previewBarcodedFood(barcode, userID)
	if err != nil {
		if errors.Is(err, errProductNotFound) {
			return c.Redirect(http.StatusSeeOther, scanPathWithQuery(mealID, mealType, dateUnix, "Product not found"))
		}
		return c.Redirect(http.StatusSeeOther, scanPathWithQuery(mealID, mealType, dateUnix, "Lookup failed. Please try again"))
	}

	backURL := "/"
	if mealID != 0 {
		backURL = editPathForType(mealID, mealType)
	} else {
		backURL = overviewPath(requestedMealDate(c))
	}
	nav := view.NavBack(userID, backURL, "Confirm Scan")
	component := view.Full(nav, view.ScanConfirm(food, barcode, mealID, mealType, dateUnix, scanSearchPath(mealID, mealType, dateUnix)))
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
	if f := db.FindFoodByBarcode(barcode, userID); f != nil {
		return f.ID, nil
	}
	p, err := fetchOpenFoodFacts(barcode)
	if err != nil {
		return 0, err
	}
	return db.CreateFoodWithBarcode(p.Name, p.Fat, p.Carb, p.Fiber, p.Protein, 100, userID, barcode)
}

func scanBarcode(c echo.Context) error {
	if strings.TrimSpace(c.Param("barcode")) == "" {
		return c.NoContent(http.StatusBadRequest)
	}
	mealID, _ := strconv.Atoi(c.QueryParam("meal"))
	mealType := c.QueryParam("type")
	dateUnix := queryDateUnix(c)
	c.Response().Header().Set("HX-Location", scanConfirmPath(c.Param("barcode"), mealID, mealType, dateUnix))
	return c.NoContent(http.StatusOK)
}

func scanBarcodeManual(c echo.Context) error {
	barcode := strings.TrimSpace(c.FormValue("barcode"))
	if barcode == "" {
		mealID, _ := strconv.Atoi(c.FormValue("meal"))
		mealType := c.FormValue("type")
		dateUnix := parseDateUnixValue(c.FormValue("date"))
		return c.Redirect(http.StatusSeeOther, scanPathWithQuery(mealID, mealType, dateUnix, "Enter a barcode to continue"))
	}
	mealID, _ := strconv.Atoi(c.FormValue("meal"))
	mealType := c.FormValue("type")
	dateUnix := parseDateUnixValue(c.FormValue("date"))
	return c.Redirect(http.StatusSeeOther, scanConfirmPath(barcode, mealID, mealType, dateUnix))
}

func scanSearchPath(mealID int, mealType string, dateUnix int64) string {
	if mealID == 0 {
		if dateUnix != 0 {
			return fmt.Sprintf("/meal/new/food_search?date=%d", dateUnix)
		}
		return "/meal/new/food_search"
	}
	if mealType == "template" {
		return fmt.Sprintf("/template/%d/food_search", mealID)
	}
	return fmt.Sprintf("/meal/%d/food_search", mealID)
}

func scanConfirmPath(barcode string, mealID int, mealType string, dateUnix int64) string {
	path := fmt.Sprintf("/scan/confirm?barcode=%s", url.QueryEscape(barcode))
	if mealID != 0 {
		path += fmt.Sprintf("&meal=%d", mealID)
	}
	if mealType != "" {
		path += "&type=" + url.QueryEscape(mealType)
	}
	if dateUnix != 0 {
		path += fmt.Sprintf("&date=%d", dateUnix)
	}
	return path
}

func scanPathWithQuery(mealID int, mealType string, dateUnix int64, errMsg string) string {
	path := "/scan"
	var params []string
	if mealID != 0 {
		params = append(params, fmt.Sprintf("meal=%d", mealID))
	}
	if mealType != "" {
		params = append(params, "type="+mealType)
	}
	if dateUnix != 0 {
		params = append(params, fmt.Sprintf("date=%d", dateUnix))
	}
	if errMsg != "" {
		params = append(params, "error="+url.QueryEscape(errMsg))
	}
	if len(params) > 0 {
		path += "?" + strings.Join(params, "&")
	}
	return path
}

func previewBarcodedFood(barcode string, userID int) (db.Food, error) {
	if f := db.FindFoodByBarcode(barcode, userID); f != nil {
		return db.Food{
			ID:    f.ID,
			Name:  f.Name,
			Grams: 100,
			Macros: db.Macro{
				Protein:  f.ProteinPerGram * 100,
				Fat:      f.FatPerGram * 100,
				Carb:     f.CarbPerGram * 100,
				Fiber:    f.FiberPerGram * 100,
				Calories: f.ProteinPerGram*100*db.ProteinKcalPerGram + f.FatPerGram*100*db.FatKcalPerGram + f.CarbPerGram*100*db.CarbKcalPerGram,
			},
		}, nil
	}
	p, err := fetchOpenFoodFacts(barcode)
	if err != nil {
		return db.Food{}, err
	}
	return db.Food{
		Name:  p.Name,
		Grams: 100,
		Macros: db.Macro{
			Calories: db.CaloriesFromGrams(p.Fat, p.Carb, p.Protein),
			Fat:      float32(p.Fat),
			Carb:     float32(p.Carb),
			Fiber:    float32(p.Fiber),
			Protein:  float32(p.Protein),
		},
	}, nil
}

func scanConfirmAdd(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	barcode := strings.TrimSpace(c.FormValue("barcode"))
	if barcode == "" {
		return c.NoContent(http.StatusBadRequest)
	}
	mealID, _ := strconv.Atoi(c.FormValue("meal"))
	mealType := c.FormValue("type")
	dateUnix := parseDateUnixValue(c.FormValue("date"))
	if mealID != 0 {
		if err := db.ValidateMealAccess(mealID, userID, mealType == "template"); err != nil {
			return handleDBErr(c, err)
		}
	}
	foodID, err := resolveBarcodedFood(barcode, userID)
	if err != nil {
		if errors.Is(err, errProductNotFound) {
			return c.Redirect(http.StatusSeeOther, scanPathWithQuery(mealID, mealType, dateUnix, "Product not found"))
		}
		return c.Redirect(http.StatusSeeOther, scanPathWithQuery(mealID, mealType, dateUnix, "Lookup failed. Please try again"))
	}
	if mealID == 0 {
		mealDate := requestedMealDate(c)
		if dateUnix != 0 {
			mealDate = time.Unix(dateUnix, 0)
		}
		mealID, err = db.CreateMeal(defaultMealName, userID, false, mealDate)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
	}
	if err := db.CreateMealItem(mealID, foodID, 100, userID, mealType == "template"); err != nil {
		return handleDBErr(c, err)
	}
	return c.Redirect(http.StatusSeeOther, editPathForType(mealID, mealType))
}
