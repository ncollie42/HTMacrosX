package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	db "myapp/DB"
	"myapp/view"
	"net/http"
	"strconv"
	"strings"
	"time"

	"myapp/auth"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed js/htmx.min.js
var htmxJS []byte

//go:embed css/daisy.min.css
var daisyCSS []byte

//go:embed js/html5-qrcode.min.js
var html5QrcodeJS []byte

//go:embed pwa/manifest.json
var manifestJSON []byte

//go:embed pwa/sw.js
var swJS []byte

//go:embed icons/icon.svg
var iconSVG []byte

var presets = map[string]db.Macro{
	"1750": {Calories: 1750, Fat: 50, Carb: 195, Fiber: 28, Protein: 130},
	"2000": {Calories: 2000, Fat: 60, Carb: 220, Fiber: 30, Protein: 145},
	"2250": {Calories: 2250, Fat: 65, Carb: 250, Fiber: 32, Protein: 165},
}

const ctxUserID = "userID"

var httpClient = &http.Client{Timeout: 10 * time.Second}

var errProductNotFound = errors.New("product not found")

// GET       -> SELECT
// POST      -> INSERT -> New
// PUT|PATCH -> UPDATE -> Edit
// DELETE    -> DELETE

func main() {
	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status} latency=${latency_human} Error=${error}\n",
	}))

	e.GET("/", overview, validate)
	e.GET("/:date", overview, validate)

	e.POST("/meal", createMeal, validate)
	e.DELETE("/meal/:id/", deleteMeal, validate)
	e.GET("/meal/:id/", findMealOrTemplate, validate)
	e.GET("/meal/:id/food_search", foodSearch, validate)
	e.POST("/food", createFood, validate)
	e.DELETE("/food/:id", deleteFood, validate)

	// Shared meal/template handlers
	e.POST("/meal/:id/join/:foodID", addFood, validate)
	e.DELETE("/meal/:id/join/:joinID", removeFood, validate)
	e.PUT("/meal/:id/join/:joinID", updateGrams, validate)
	e.PUT("/meal/:id/name", updateName, validate)

	e.GET("/template/", findAllTemplates, validate)
	e.POST("/template/", createTemplate, validate)
	e.GET("/template/:id/", findMealOrTemplate, validate)
	e.GET("/template/:id/food_search", foodSearch, validate)
	e.DELETE("/template/:id/", deleteTemplate, validate)
	e.POST("/template/:id/", templateToMeal, validate)

	e.POST("/template/:id/join/:foodID", addFood, validate)
	e.DELETE("/template/:id/join/:joinID", removeFood, validate)
	e.PUT("/template/:id/join/:joinID", updateGrams, validate)
	e.PUT("/template/:id/name", updateName, validate)

	e.GET("/scan", scanView, validate)
	e.POST("/scan/:barcode", scanBarcode, validate)

	e.GET("/favicon.ico", fav)
	e.GET("/htmx", htmx)
	e.GET("/daisy", daisy)
	e.GET("/html5qrcode", html5qrcode)
	e.GET("/manifest.json", manifest)
	e.GET("/sw.js", sw)
	e.GET("/icon", icon)
	e.GET("/signin", signinView)
	e.POST("/signin", signin)
	e.GET("/signup", signupView)
	e.POST("/signup", signup)
	e.POST("/signout", signout, validate)

	e.GET("/onboarding", onboardingView, validate)
	e.POST("/onboarding", saveOnboarding, validate)

	e.GET("/settings", settings, validate)
	e.PUT("/settings", updateSettings, validate)

	e.Logger.Fatal(e.Start(":8080"))
}

// ---------- middleware -------------------

func validate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// NOTE: For Dev
		// userID, err := 2, error(nil)
		userID, err := auth.GetUserFromCookie(c)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/signin")
		}
		c.Set(ctxUserID, userID)
		return next(c)
	}
}

func handleDBErr(c echo.Context, err error) error {
	if errors.Is(err, db.ErrNotOwned) {
		return c.NoContent(http.StatusNotFound)
	}
	return c.NoContent(http.StatusInternalServerError)
}

// ---------- Handlers ---------------------

func fav(c echo.Context) error {
	return c.NoContent(http.StatusNotFound)
}

func htmx(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/javascript")
	_, err := c.Response().Writer.Write(htmxJS)
	return err
}

func daisy(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/css")
	_, err := c.Response().Writer.Write(daisyCSS)
	return err
}

func html5qrcode(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/javascript")
	_, err := c.Response().Writer.Write(html5QrcodeJS)
	return err
}

func manifest(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/manifest+json")
	c.Response().Header().Set("Cache-Control", "public, max-age=86400")
	_, err := c.Response().Writer.Write(manifestJSON)
	return err
}

func sw(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/javascript")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Service-Worker-Allowed", "/")
	_, err := c.Response().Writer.Write(swJS)
	return err
}

func icon(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "image/svg+xml")
	c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	_, err := c.Response().Writer.Write(iconSVG)
	return err
}

func scanView(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	mealID, _ := strconv.Atoi(c.QueryParam("meal"))
	nav := view.NavBack(userID, "/", "Scan")
	scan := view.ScanPage(mealID)
	component := view.Full(nav, scan)
	return component.Render(context.Background(), c.Response().Writer)
}

type openFoodFactsProduct struct {
	Name    string
	Fat     float64
	Carb    float64
	Fiber   float64
	Protein float64
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
		mealID, err = db.CreateMeal(time.Now().Format("3:04 PM"), userID, false)
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

func overview(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)

	target := db.GetUserTargets(userID)

	timeStr := c.Param("date")
	date := strconvTime(timeStr)
	macros := db.GetMealItemsByDate(userID, date)

	totalMacros := db.SumMacros(macros)
	macrosByID := db.SumMacrosByID(macros)

	nav := view.Nav(userID)
	overview := view.DayOverview(date, totalMacros, target)
	quickview := view.DayQuickview(macrosByID, date)
	bottomNav := view.BottomNav()
	component := view.Full(nav, overview, quickview, bottomNav)
	return component.Render(context.Background(), c.Response().Writer)
}

// --------  Food Search ----------------------------
// NOTE: 1 template 2 endpoints, 1 function

func foodSearch(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	search := c.FormValue("search")
	// TODO: Anitize input
	foods := db.FoodSearch(search, userID)

	id, _ := strconv.Atoi(c.Param("id"))
	nav := view.NavBack(userID, editPath(c, id), "Add Food")
	searchResult := view.FoodSearch(foods)
	component := view.Full(nav, searchResult)
	return component.Render(context.Background(), c.Response().Writer)
}

func createFood(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	name := c.FormValue("name")
	fat, _ := strconv.ParseFloat(c.FormValue("fat"), 64)
	carb, _ := strconv.ParseFloat(c.FormValue("carb"), 64)
	fiber, _ := strconv.ParseFloat(c.FormValue("fiber"), 64)
	protein, _ := strconv.ParseFloat(c.FormValue("protein"), 64)
	grams, _ := strconv.ParseFloat(c.FormValue("grams"), 64)

	if _, err := db.CreateFood(name, fat, carb, fiber, protein, grams, userID); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	foods := db.FoodSearch("", userID)
	component := view.FoodSearchResults(foods)
	return component.Render(context.Background(), c.Response().Writer)
}

func deleteFood(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	foodID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if err := db.DeleteFood(foodID, userID); err != nil {
		return handleDBErr(c, err)
	}
	return c.NoContent(http.StatusOK)
}

// --------  Templates ----------------------------
func templateToMeal(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	templateID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	token := c.FormValue("token")
	if ok := auth.ValidateDupToken(c, token); !ok {
		return fmt.Errorf("Invalid token")
	}
	if _, err := db.TemplateToMeal(templateID, userID); err != nil {
		return handleDBErr(c, err)
	}

	auth.ClearDupToken(c)

	c.Response().Header().Set("HX-Location", "/")
	return c.NoContent(http.StatusOK)
}

func findAllTemplates(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	macros := db.GetTemplates(userID)

	token := auth.GenToken()
	auth.SetDupToken(c, token)
	macrosByID := db.SumMacrosByID(macros)

	nav := view.NavBack(userID, "/", "Templates")
	overview := view.TemplateOverview(macrosByID, token)
	component := view.Full(nav, overview)
	return component.Render(context.Background(), c.Response().Writer)
}

func createTemplate(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)

	time := time.Now().Format("3:04 PM")
	templateID, err := db.CreateMeal(time, userID, true)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	c.Response().Header().Set("HX-Location", fmt.Sprint("/template/", templateID, "/food_search"))
	return c.NoContent(http.StatusOK)
}

func deleteTemplate(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if err := db.DeleteMeal(id, userID); err != nil {
		return handleDBErr(c, err)
	}
	return c.NoContent(http.StatusOK)
}

// --------  Shared Meal/Template Handlers ----------------------------
func findMealOrTemplate(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	meal, err := db.GetMealByID(id, userID)
	if err != nil {
		return handleDBErr(c, err)
	}

	currentURL := c.Request().Header.Get("HX-Current-URL")
	if strings.Contains(currentURL, "/food_search") {
		c.Response().Header().Set("HX-Replace-Url", c.Request().URL.Path)
	}

	var backURL, title string
	if isTemplate(c) {
		backURL = "/template/"
		title = "Edit Template"
	} else {
		backURL = "/"
		title = "Edit Meal"
	}
	nav := view.NavBack(userID, backURL, title)
	mealEdit := view.MealEdit(meal)
	mealNav := view.MealEditNav(id)
	component := view.Full(nav, mealEdit, mealNav)
	return component.Render(context.Background(), c.Response().Writer)
}

func addFood(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	mealID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	foodID, err := strconv.Atoi(c.Param("foodID"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if err := db.CreateMealItem(mealID, foodID, 100, userID); err != nil {
		return handleDBErr(c, err)
	}
	c.Response().Header().Set("HX-Location", editPath(c, mealID))
	return c.NoContent(http.StatusOK)
}

func removeFood(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	mealID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	joinID, err := strconv.Atoi(c.Param("joinID"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if err := db.DeleteMealItem(joinID, userID); err != nil {
		return handleDBErr(c, err)
	}
	meal, err := db.GetMealByID(mealID, userID)
	if err != nil {
		return handleDBErr(c, err)
	}
	return view.MealTotalsOOB(meal.Items).Render(context.Background(), c.Response().Writer)
}

func updateGrams(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	mealID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	joinID, err := strconv.Atoi(c.Param("joinID"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	grams, err := strconv.ParseFloat(c.FormValue("grams"), 64)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if err := db.UpdateMealItem(joinID, grams, userID); err != nil {
		return handleDBErr(c, err)
	}
	meal, err := db.GetMealByID(mealID, userID)
	if err != nil {
		return handleDBErr(c, err)
	}
	for _, item := range meal.Items {
		if item.ItemID == joinID {
			if err := view.GramEdit(item).Render(context.Background(), c.Response().Writer); err != nil {
				return err
			}
			break
		}
	}
	return view.MealTotalsOOB(meal.Items).Render(context.Background(), c.Response().Writer)
}

func updateName(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	name := c.FormValue("name")
	if err := db.UpdateMealName(id, userID, name); err != nil {
		return handleDBErr(c, err)
	}
	return nil
}

// --------  Sign in / up -------------------------
func signin(c echo.Context) error {
	login := c.FormValue("login")
	password := c.FormValue("password")
	err := auth.Signin(c, login, password)

	if err != nil {
		component := view.SignError(err.Error())
		return component.Render(context.Background(), c.Response().Writer)
	}

	c.Response().Header().Set("HX-Location", "/")
	return c.NoContent(http.StatusOK)
}

func settings(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	targets := db.GetUserTargets(userID)

	if p := c.QueryParam("preset"); p != "" {
		if preset, ok := presets[p]; ok {
			targets = preset
		}
	}

	if c.Request().Header.Get("HX-Request") != "" {
		return view.Settings(targets).Render(context.Background(), c.Response().Writer)
	}

	nav := view.NavBack(userID, "/", "Settings")
	settingsForm := view.Settings(targets)
	component := view.Full(nav, settingsForm)
	return component.Render(context.Background(), c.Response().Writer)
}

func updateSettings(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)

	fat, _ := strconv.ParseFloat(c.FormValue("fat"), 32)
	carb, _ := strconv.ParseFloat(c.FormValue("carb"), 32)
	fiber, _ := strconv.ParseFloat(c.FormValue("fiber"), 32)
	protein, _ := strconv.ParseFloat(c.FormValue("protein"), 32)

	targets := db.Macro{
		Calories: db.CaloriesFromGrams(fat, carb, protein),
		Fat:      float32(fat),
		Carb:     float32(carb),
		Fiber:    float32(fiber),
		Protein:  float32(protein),
	}
	if err := db.UpdateUserTargets(userID, targets); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	component := view.Settings(targets)
	return component.Render(context.Background(), c.Response().Writer)
}

func onboardingView(c echo.Context) error {
	targets := presets["2000"]
	if p := c.QueryParam("preset"); p != "" {
		if preset, ok := presets[p]; ok {
			targets = preset
		}
	}

	if c.Request().Header.Get("HX-Request") != "" {
		return view.OnboardingForm(targets).Render(context.Background(), c.Response().Writer)
	}

	component := view.Full(view.Onboarding(targets))
	return component.Render(context.Background(), c.Response().Writer)
}

func saveOnboarding(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)

	fat, _ := strconv.ParseFloat(c.FormValue("fat"), 32)
	carb, _ := strconv.ParseFloat(c.FormValue("carb"), 32)
	fiber, _ := strconv.ParseFloat(c.FormValue("fiber"), 32)
	protein, _ := strconv.ParseFloat(c.FormValue("protein"), 32)

	targets := db.Macro{
		Calories: db.CaloriesFromGrams(fat, carb, protein),
		Fat:      float32(fat),
		Carb:     float32(carb),
		Fiber:    float32(fiber),
		Protein:  float32(protein),
	}
	if err := db.UpdateUserTargets(userID, targets); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	c.Response().Header().Set("HX-Location", "/")
	return c.NoContent(http.StatusOK)
}

func signout(c echo.Context) error {
	auth.ClearCookie(c)
	c.Response().Header().Set("HX-Location", "/signin")
	return c.NoContent(http.StatusOK)
}

func signupView(c echo.Context) error {
	component := view.Full(view.Signup())
	return component.Render(context.Background(), c.Response().Writer)
}

func signup(c echo.Context) error {
	login := c.FormValue("login")
	password := c.FormValue("password")
	confirm := c.FormValue("confirm")

	if password != confirm {
		component := view.SignError("Passwords do not match")
		return component.Render(context.Background(), c.Response().Writer)
	}

	err := auth.Signup(c, login, password)
	if err != nil {
		component := view.SignError(err.Error())
		return component.Render(context.Background(), c.Response().Writer)
	}

	c.Response().Header().Set("HX-Location", "/onboarding")
	return c.NoContent(http.StatusOK)
}

func signinView(c echo.Context) error {
	signin := view.Signin()
	component := view.Full(signin)
	return component.Render(context.Background(), c.Response().Writer)
}

// --------  Meals --------------------------------
func createMeal(c echo.Context) error {
	// NOTE: this will create a empty meal entries, will probably want a way to clean it up in the future.
	userID := c.Get(ctxUserID).(int)

	time := time.Now().Format("3:04 PM")
	mealID, err := db.CreateMeal(time, userID, false)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	c.Response().Header().Set("HX-Location", fmt.Sprint("/meal/", mealID, "/food_search"))
	return c.NoContent(http.StatusOK)
}

func deleteMeal(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	if err := db.DeleteMeal(id, userID); err != nil {
		return handleDBErr(c, err)
	}

	c.Response().Header().Set("HX-Location", "/")
	return c.NoContent(http.StatusOK)
}

// --------  Util ---------------------------------

func isTemplate(c echo.Context) bool {
	return strings.HasPrefix(c.Path(), "/template")
}

func editPath(c echo.Context, id int) string {
	if isTemplate(c) {
		return fmt.Sprintf("/template/%d/", id)
	}
	return fmt.Sprintf("/meal/%d/", id)
}

func strconvTime(num string) time.Time {
	if num == "" {
		return time.Now()
	}
	ii, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return time.Now()
	}
	return time.Unix(ii, 0)
}
