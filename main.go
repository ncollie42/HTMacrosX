package main

import (
	"context"
	"embed"
	"fmt"
	db "myapp/DB"
	"myapp/view"
	"net/http"
	"strconv"
	"time"

	"myapp/auth"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed js/htmx.js
var htmxJS []byte

// go:embed js/*
var resources embed.FS

// GET       -> SELECT
// POST      -> INSERT -> New
// PUT|PATCH -> UPDATE -> Edit
// DELETE    -> DELETE

func main() {
	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status} latency=${latency_human}\n",
	}))

	e.GET("/", overview, validate)
	e.GET("/:date", overview, validate)
	// TODO: Group Validate + resources Auth
	e.POST("/meal", createMeal, validate)
	e.DELETE("/meal/:id/delete", deleteMeal)
	e.GET("/meal/:id", findMeal, validate)

	e.GET("/meal/:id/food_search", foodSearch, validate)
	e.GET("/template/:id/food_search", foodSearch, validate)

	e.POST("/meal/:id/join", createMealJoin)
	e.DELETE("/meal/:id/join", deleteMealJoin)
	e.PUT("/meal/:id/join", updateMealJoin)

	e.PUT("/meal/:id/name", updateMealName)
	e.PUT("/template/:id/name", updateTemplateName)

	e.GET("/template", findAllTemplates, validate)
	e.POST("/template", createTemplate, validate)
	e.GET("/template/:id", findTemplate, validate)
	e.DELETE("/template/:id/delete", deleteTemplate, validate)
	e.POST("/template/:id", templateToMeal, validate)

	e.POST("/template/:id/join", createTemplateJoin)
	e.DELETE("/template/:id/join", deleteTemplateJoin)
	e.PUT("/template/:id/join", updateTemplateJoin)

	e.GET("/htmx", htmx)
	e.GET("/signin", signinView)
	e.POST("/signin", signin)
	e.GET("/signup", signupView)
	e.POST("/signup", signup)
	e.GET("/signout", signout)

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

func htmx(c echo.Context) error {
	fmt.Fprint(c.Response().Writer, string(htmxJS))
	return nil
}

func overview(c echo.Context) error {
	userID := c.Get("userID").(int)

	// Target macros
	target := db.Macro{
		Calories: 1751.6,
		Fat:      44.8,
		Carb:     247.1,
		Fiber:    32.0,
		Protein:  90.0,
	}

	timeStr := c.Param("date")
	date := strconvTime(timeStr)
	macros := db.GetEntriessByDate(userID, date)

	totalMacros := db.SumMacros(macros)
	macrosByID := db.SumMacrosByID(macros)

	nav := view.Nav(userID)
	overview := view.DayOverview(date, totalMacros, target)
	quickview := view.DayQuickview2(macrosByID)
	component := view.Full(nav, overview, quickview)
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

// --------  Templates ----------------------------
func templateToMeal(c echo.Context) error {
	templateID := c.Param("id")
	userID := c.Get("userID").(int)
	db.TemplateToMeal(templateID, userID)

	c.Response().Header().Set("HX-Location", "/")
	return c.NoContent(http.StatusOK)
}

func findTemplate(c echo.Context) error {
	templateID := c.Param("id")
	meals := db.GetTemplateByID(templateID)

	templateEdit := view.MealEdit(meals)
	component := view.Full(templateEdit)
	return component.Render(context.Background(), c.Response().Writer)
}

func findAllTemplates(c echo.Context) error {
	userID := c.Get("userID").(int)
	macros := db.GetTemplateEntriess(userID)

	macrosByID := db.SumMacrosByID(macros)

	nav := view.Nav(userID)
	overview := view.TemplateOverview(macrosByID)
	component := view.Full(nav, overview)
	return component.Render(context.Background(), c.Response().Writer)
}

func createTemplate(c echo.Context) error {
	// NOTE: this will create a empty meal entries, will probably want a way to clean it up in the future.
	userID := c.Get("userID").(int)

	time := time.Now().Format("3:04 PM")
	templateID := db.CreateTemplate(time, userID)

	c.Response().Header().Set("HX-Location", fmt.Sprint("/template/", templateID, "/food_search"))
	return c.NoContent(http.StatusOK)
}

func deleteTemplate(c echo.Context) error {
	id := c.Param("id")

	db.DeleteTemplate(id)

	c.Response().Header().Set("HX-Location", "/template")
	return c.NoContent(http.StatusOK)
}

// --------  Template Join ----------------------------
func createTemplateJoin(c echo.Context) error {
	templateID := c.Param("id")
	foodID := c.FormValue("foodID")

	// TODO: Query for default food.grams to show, for now display base 100g
	grams := "100"

	db.CreateTemplateJoin(templateID, foodID, grams)

	// TODO: user some HTTP.URL builder, not Sprint
	c.Response().Header().Set("HX-Location", fmt.Sprint("/template/", templateID))
	return c.NoContent(http.StatusOK)
}

func updateTemplateJoin(c echo.Context) error {
	id := c.Param("id")
	grams := c.FormValue("grams")
	updatedFood := db.UpdateTemplateJoin(id, grams)

	component := view.GramEdit(updatedFood)
	return component.Render(context.Background(), c.Response().Writer)
}

func deleteTemplateJoin(c echo.Context) error {
	id := c.Param("id")

	db.DeleteTemplateJoin(id)
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
	db.UpdateTemplateName(id, name)
	return nil
}

// --------  Meal Join ----------------------------
func createMealJoin(c echo.Context) error {
	mealID := c.Param("id")
	foodID := c.FormValue("foodID")

	// TODO: Query for default food.grams to show, for now display base 100g
	grams := "100"

	db.CreateMealJoin(mealID, foodID, grams)

	// TODO: user some HTTP.URL builder, not Sprint
	c.Response().Header().Set("HX-Location", fmt.Sprint("/meal/", mealID))
	return c.NoContent(http.StatusOK)
}

func updateMealJoin(c echo.Context) error {
	id := c.Param("id")
	grams := c.FormValue("grams")
	updatedFood := db.UpdateMealJoin(id, grams)

	component := view.GramEdit(updatedFood)
	return component.Render(context.Background(), c.Response().Writer)
}

func deleteMealJoin(c echo.Context) error {
	id := c.Param("id")

	db.DeleteMealJoin(id)
	return c.NoContent(http.StatusOK)
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

func signup(c echo.Context) error {
	login := c.FormValue("login")
	password := c.FormValue("password")
	err := auth.Signup(c, login, password)

	if err != nil {
		fmt.Println("User name is already taken", login)
		component := view.SignError("Invalid log in or password")
		return component.Render(context.Background(), c.Response().Writer)
	}

	c.Response().Header().Set("HX-Location", "/")
	return c.NoContent(http.StatusOK)
}

func signout(c echo.Context) error {
	auth.ClearCookie(c)
	c.Response().Header().Set("HX-Location", "/signin")
	return c.NoContent(http.StatusOK)
}

func signinView(c echo.Context) error {
	signin := view.Signin()
	component := view.Full(signin)
	return component.Render(context.Background(), c.Response().Writer)
}

func signupView(c echo.Context) error {
	signup := view.Signup()
	component := view.Full(signup)
	return component.Render(context.Background(), c.Response().Writer)
}

// --------  Meals --------------------------------
func findMeal(c echo.Context) error {
	userID := c.Get("userID").(int)
	id := c.Param("id")

	meals := db.GetMealByID(id)

	nav := view.Nav(userID)
	mealEdit := view.MealEdit(meals)
	component := view.Full(nav, mealEdit)
	return component.Render(context.Background(), c.Response().Writer)
}

func createMeal(c echo.Context) error {
	// NOTE: this will create a empty meal entries, will probably want a way to clean it up in the future.
	userID := c.Get("userID").(int)

	time := time.Now().Format("3:04 PM")
	mealID := db.CreateMeal(time, userID)

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
	ii, err := strconv.ParseInt(num, 10, 32)
	if err != nil {
		panic(err)
	}
	return time.Unix(int64(ii), 0)
}
