package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	db "myapp/DB"
	"myapp/view"
	"net/http"
	"strconv"
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

var presets = map[string]db.Macro{
	"1750": {Calories: 1750, Fat: 50, Carb: 195, Fiber: 28, Protein: 130},
	"2000": {Calories: 2000, Fat: 60, Carb: 220, Fiber: 30, Protein: 145},
	"2250": {Calories: 2250, Fat: 65, Carb: 250, Fiber: 32, Protein: 165},
}

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
	// TODO: Group Validate + resources Auth
	e.POST("/meal", createMeal, validate)
	e.DELETE("/meal/:id/", deleteMeal)
	e.GET("/meal/:id/", findMeal, validate)

	e.GET("/meal/:id/food_search", foodSearch, validate)
	e.GET("/template/:id/food_search", foodSearch, validate)
	e.POST("/food", createFood, validate)

	e.POST("/meal/:mID/join/:jID", createMealJoin)
	e.DELETE("/meal/:mid/join/:id", deleteMealJoin)
	e.PUT("/meal/:mID/join/:id", updateMealJoin)

	e.PUT("/meal/:id/name", updateMealName)
	e.PUT("/template/:id/name", updateTemplateName)

	e.GET("/template/", findAllTemplates, validate)
	e.POST("/template/", createTemplate, validate)
	e.GET("/template/:id/", findTemplate, validate)
	e.DELETE("/template/:id/", deleteTemplate, validate)
	e.POST("/template/:id/", templateToMeal, validate)

	e.POST("/template/:tID/join/:jID", createTemplateJoin)
	e.DELETE("/template/:tID/join/:jID", deleteTemplateJoin)
	e.PUT("/template/:id/join", updateTemplateJoin)

	e.GET("/scan", scanView, validate)
	e.POST("/scan/:barcode", scanBarcode, validate)

	e.GET("/favicon.ico", fav)
	e.GET("/htmx", htmx)
	e.GET("/daisy", daisy)
	e.GET("/html5qrcode", html5qrcode)
	e.GET("/signin", signinView)
	e.POST("/signin", signin)
	e.GET("/signup", signupView)
	e.POST("/signup", signup)
	e.GET("/signout", signout)

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
			fmt.Println(err.Error())
			return c.Redirect(http.StatusSeeOther, "/signin")
		}
		c.Set("userID", userID)
		return next(c)
	}
}

// func timedLogger(next echo.HandlerFunc) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		start := time.Now()
// 		defer func() {
// 			log.Println(time.Since(start))
// 		}()

// 		if err := next(c); err != nil {
// 			c.Error(err)
// 		}
// 		return nil
// 	}
// }

// ---------- Handlers ---------------------

func fav(c echo.Context) error {
	return c.NoContent(http.StatusNotFound)
}

func htmx(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/javascript")
	fmt.Fprint(c.Response().Writer, string(htmxJS))
	return nil
}

func daisy(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/css")
	fmt.Fprint(c.Response().Writer, string(daisyCSS))
	return nil
}

func html5qrcode(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/javascript")
	fmt.Fprint(c.Response().Writer, string(html5QrcodeJS))
	return nil
}

func scanView(c echo.Context) error {
	userID := c.Get("userID").(int)
	nav := view.Nav(userID)
	scan := view.ScanPage()
	component := view.Full(nav, scan)
	return component.Render(context.Background(), c.Response().Writer)
}

func scanBarcode(c echo.Context) error {
	userID := c.Get("userID").(int)
	barcode := c.Param("barcode")

	// Check cache first
	existing := db.FindFoodByBarcode(barcode)
	var foodID int
	if existing != nil {
		foodID = existing.ID
	} else {
		// Fetch from Open Food Facts
		resp, err := http.Get("https://world.openfoodfacts.org/api/v2/product/" + barcode + ".json")
		if err != nil {
			return c.JSON(http.StatusBadGateway, map[string]string{"error": "Failed to fetch product"})
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
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
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Product not found"})
		}

		p := result.Product
		name := p.ProductName
		if name == "" {
			name = "Unknown (" + barcode + ")"
		}

		foodID = db.CreateFoodWithBarcode(
			name,
			p.Nutriments.Fat100g,
			p.Nutriments.Carbs100g,
			p.Nutriments.Fiber100g,
			p.Nutriments.Prot100g,
			100,
			userID,
			barcode,
		)
	}

	// Create meal and add the food
	mealTime := time.Now().Format("3:04 PM")
	mealID := db.CreateMeal(mealTime, userID, false)
	db.CreateMealJoin(strconv.Itoa(mealID), strconv.Itoa(foodID), "100")

	c.Response().Header().Set("HX-Location", fmt.Sprint("/meal/", mealID, "/"))
	return c.NoContent(http.StatusOK)
}

func overview(c echo.Context) error {
	userID := c.Get("userID").(int)

	target := db.GetUserTargets(userID)

	timeStr := c.Param("date")
	date := strconvTime(timeStr)
	macros := db.GetEntriessByDate(userID, date)

	totalMacros := db.SumMacros(macros)
	macrosByID := db.SumMacrosByID(macros)

	nav := view.Nav(userID)
	overview := view.DayOverview(date, totalMacros, target)
	quickview := view.DayQuickview(macrosByID)
	bottomNav := view.BottomNav()
	component := view.Full(nav, overview, quickview, bottomNav)
	return component.Render(context.Background(), c.Response().Writer)
}

// --------  Food Search ----------------------------
// NOTE: 1 template 2 endpoints, 1 function

func foodSearch(c echo.Context) error {
	userID := c.Get("userID").(int)
	search := c.FormValue("search")
	// TODO: Anitize input
	foods := db.FoodSearch(search, userID)

	nav := view.Nav(userID)
	searchResult := view.FoodSearch(foods)
	component := view.Full(nav, searchResult)
	return component.Render(context.Background(), c.Response().Writer)
}

func createFood(c echo.Context) error {
	userID := c.Get("userID").(int)
	name := c.FormValue("name")
	fat, _ := strconv.ParseFloat(c.FormValue("fat"), 64)
	carb, _ := strconv.ParseFloat(c.FormValue("carb"), 64)
	fiber, _ := strconv.ParseFloat(c.FormValue("fiber"), 64)
	protein, _ := strconv.ParseFloat(c.FormValue("protein"), 64)
	grams, _ := strconv.ParseFloat(c.FormValue("grams"), 64)

	db.CreateFood(name, fat, carb, fiber, protein, grams, userID)

	foods := db.FoodSearch("", userID)
	component := view.FoodSearchResults(foods)
	return component.Render(context.Background(), c.Response().Writer)
}

// --------  Templates ----------------------------
func templateToMeal(c echo.Context) error {
	templateID := c.Param("id")
	userID := c.Get("userID").(int)
	token := c.FormValue("token")
	if ok := auth.ValidateDupToken(c, token); !ok {
		return fmt.Errorf("Invalid token")
	}
	db.TemplateToMeal(templateID, userID)

	auth.ClearDupToken(c)

	c.Response().Header().Set("HX-Location", "/")
	return c.NoContent(http.StatusOK)
}

func findTemplate(c echo.Context) error {
	userID := c.Get("userID").(int)
	templateID := c.Param("id")
	meals := db.GetMealByID(templateID)

	nav := view.Nav(userID)
	templateEdit := view.MealEdit(meals)
	mealNav := view.MealEditNav()
	component := view.Full(nav, templateEdit, mealNav)
	return component.Render(context.Background(), c.Response().Writer)
}

func findAllTemplates(c echo.Context) error {
	userID := c.Get("userID").(int)
	macros := db.GetPresetsEntries(userID)

	token := auth.GenToken()
	auth.SetDupToken(c, token)
	macrosByID := db.SumMacrosByID(macros)

	nav := view.Nav(userID)
	overview := view.TemplateOverview(macrosByID, token)
	component := view.Full(nav, overview)
	return component.Render(context.Background(), c.Response().Writer)
}

func createTemplate(c echo.Context) error {
	// NOTE: this will create a empty meal entries, will probably want a way to clean it up in the future.
	userID := c.Get("userID").(int)

	time := time.Now().Format("3:04 PM")
	templateID := db.CreateMeal(time, userID, true)

	c.Response().Header().Set("HX-Location", fmt.Sprint("/template/", templateID, "/food_search"))
	return c.NoContent(http.StatusOK)
}

func deleteTemplate(c echo.Context) error {
	id := c.Param("id")

	db.DeleteMeal(id)

	return c.NoContent(http.StatusOK)
}

// --------  Template Join ----------------------------
func createTemplateJoin(c echo.Context) error {
	templateID := c.Param("tID")
	foodID := c.Param("jID")

	// TODO: Query for default food.grams to show, for now display base 100g
	grams := "100"

	db.CreateMealJoin(templateID, foodID, grams)

	// This will GET the current base URL IE: /template/#/ - Current URL /template/#/foodsearch
	c.Response().Header().Set("HX-Location", ".")
	return c.NoContent(http.StatusOK)
}

func updateTemplateJoin(c echo.Context) error {
	id := c.Param("id")
	grams := c.FormValue("grams")
	updatedFood := db.UpdateMealJoin(id, grams)

	component := view.GramEdit(updatedFood)
	return component.Render(context.Background(), c.Response().Writer)
}

func deleteTemplateJoin(c echo.Context) error {
	id := c.Param("jID")

	db.DeleteMealJoin(id)
	return c.NoContent(http.StatusOK)
}

// --------  Meal Name ----------------------------
func updateMealName(c echo.Context) error {
	id := c.Param("id")
	name := c.FormValue("name")
	db.UpdateMealName(id, name)
	return nil
}

// --------  Template Name ----------------------------
func updateTemplateName(c echo.Context) error {
	id := c.Param("id")
	name := c.FormValue("name")
	db.UpdateMealName(id, name)
	return nil
}

// --------  Meal Join ----------------------------
func createMealJoin(c echo.Context) error {
	mealID := c.Param("mID")
	foodID := c.Param("jID")

	// TODO: Query for default food.grams to show, for now display base 100g
	grams := "100"

	db.CreateMealJoin(mealID, foodID, grams)

	// This will GET the current base URL IE: /meal/#/ - Current URL /meal/#/foodsearch
	c.Response().Header().Set("HX-Location", ".")
	return c.NoContent(http.StatusOK)
}

func updateMealJoin(c echo.Context) error {
	mealID := c.Param("mID")
	id := c.Param("id")
	grams := c.FormValue("grams")
	updatedFood := db.UpdateMealJoin(id, grams)
	meal := db.GetMealByID(mealID)

	view.GramEdit(updatedFood).Render(context.Background(), c.Response().Writer)
	return view.MealTotalsOOB(meal.Foods).Render(context.Background(), c.Response().Writer)
}

func deleteMealJoin(c echo.Context) error {
	mealID := c.Param("mid")
	id := c.Param("id")

	db.DeleteMealJoin(id)
	meal := db.GetMealByID(mealID)
	return view.MealTotalsOOB(meal.Foods).Render(context.Background(), c.Response().Writer)
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
	userID := c.Get("userID").(int)
	targets := db.GetUserTargets(userID)

	if p := c.QueryParam("preset"); p != "" {
		if preset, ok := presets[p]; ok {
			targets = preset
		}
	}

	if c.Request().Header.Get("HX-Request") != "" {
		return view.Settings(targets).Render(context.Background(), c.Response().Writer)
	}

	nav := view.Nav(userID)
	settingsForm := view.Settings(targets)
	component := view.Full(nav, settingsForm)
	return component.Render(context.Background(), c.Response().Writer)
}

func updateSettings(c echo.Context) error {
	userID := c.Get("userID").(int)

	fat, _ := strconv.ParseFloat(c.FormValue("fat"), 32)
	carb, _ := strconv.ParseFloat(c.FormValue("carb"), 32)
	fiber, _ := strconv.ParseFloat(c.FormValue("fiber"), 32)
	protein, _ := strconv.ParseFloat(c.FormValue("protein"), 32)

	targets := db.Macro{
		Calories: float32(fat*9 + carb*4 + protein*4),
		Fat:      float32(fat),
		Carb:     float32(carb),
		Fiber:    float32(fiber),
		Protein:  float32(protein),
	}
	db.UpdateUserTargets(userID, targets)

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
	userID := c.Get("userID").(int)

	fat, _ := strconv.ParseFloat(c.FormValue("fat"), 32)
	carb, _ := strconv.ParseFloat(c.FormValue("carb"), 32)
	fiber, _ := strconv.ParseFloat(c.FormValue("fiber"), 32)
	protein, _ := strconv.ParseFloat(c.FormValue("protein"), 32)

	targets := db.Macro{
		Calories: float32(fat*9 + carb*4 + protein*4),
		Fat:      float32(fat),
		Carb:     float32(carb),
		Fiber:    float32(fiber),
		Protein:  float32(protein),
	}
	db.UpdateUserTargets(userID, targets)

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
func findMeal(c echo.Context) error {
	userID := c.Get("userID").(int)
	id := c.Param("id")

	meals := db.GetMealByID(id)

	nav := view.Nav(userID)
	mealEdit := view.MealEdit(meals)
	mealNav := view.MealEditNav()
	component := view.Full(nav, mealEdit, mealNav)
	return component.Render(context.Background(), c.Response().Writer)
}

func createMeal(c echo.Context) error {
	// NOTE: this will create a empty meal entries, will probably want a way to clean it up in the future.
	userID := c.Get("userID").(int)

	time := time.Now().Format("3:04 PM")
	mealID := db.CreateMeal(time, userID, false)

	c.Response().Header().Set("HX-Location", fmt.Sprint("/meal/", mealID, "/food_search"))
	return c.NoContent(http.StatusOK)
}

func deleteMeal(c echo.Context) error {
	id := c.Param("id")

	db.DeleteMeal(id)

	c.Response().Header().Set("HX-Location", "/")
	return c.NoContent(http.StatusOK)
}

// --------  Util ---------------------------------

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
