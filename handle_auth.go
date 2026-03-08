package main

import (
	"context"
	"fmt"
	db "myapp/DB"
	"myapp/auth"
	"myapp/view"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

var presets = map[string]db.Macro{
	"1750": {Calories: 1750, Fat: 50, Carb: 195, Fiber: 28, Protein: 130},
	"2000": {Calories: 2000, Fat: 60, Carb: 220, Fiber: 30, Protein: 145},
	"2250": {Calories: 2250, Fat: 65, Carb: 250, Fiber: 32, Protein: 165},
}

func registerAuthRoutes(e *echo.Echo) {
	e.GET("/signin", signinView)
	e.POST("/signin", signin)
	e.GET("/signup", signupView)
	e.POST("/signup", signup)
	e.POST("/signout", signout, validate)

	e.GET("/onboarding", onboardingView, validate)
	e.POST("/onboarding", saveOnboarding, validate)

	e.GET("/settings", settings, validate)
	e.PUT("/settings", updateSettings, validate)
}

func signinView(c echo.Context) error {
	signin := view.Signin()
	component := view.Full(signin)
	return component.Render(context.Background(), c.Response().Writer)
}

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

func signout(c echo.Context) error {
	auth.ClearCookie(c)
	c.Response().Header().Set("HX-Location", "/signin")
	return c.NoContent(http.StatusOK)
}

func settings(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)

	targets, ok := presets[c.QueryParam("preset")]
	if !ok {
		targets = db.GetUserTargets(userID)
	}

	if c.Request().Header.Get("HX-Request") != "" {
		return view.MacroTargets(targets, view.SettingsCfg, "").Render(context.Background(), c.Response().Writer)
	}

	nav := view.NavBack(userID, "/", "Settings")
	settingsForm := view.MacroTargets(targets, view.SettingsCfg, "")
	component := view.Full(nav, settingsForm)
	return component.Render(context.Background(), c.Response().Writer)
}

func parseMacroField(raw string, label string) (float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("%s is required", label)
	}
	value, err := strconv.ParseFloat(raw, 32)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number", label)
	}
	if value < 0 {
		return 0, fmt.Errorf("%s cannot be negative", label)
	}
	return value, nil
}

func parseMacroForm(c echo.Context) (db.Macro, error) {
	fat, err := parseMacroField(c.FormValue("fat"), "Fat")
	if err != nil {
		return db.Macro{}, err
	}
	carb, err := parseMacroField(c.FormValue("carb"), "Carbs")
	if err != nil {
		return db.Macro{}, err
	}
	fiber, err := parseMacroField(c.FormValue("fiber"), "Fiber")
	if err != nil {
		return db.Macro{}, err
	}
	protein, err := parseMacroField(c.FormValue("protein"), "Protein")
	if err != nil {
		return db.Macro{}, err
	}
	calories := db.CaloriesFromGrams(fat, carb, protein)
	if calories <= 0 {
		return db.Macro{}, fmt.Errorf("Calories must be greater than 0")
	}
	return db.Macro{
		Calories: calories,
		Fat:      float32(fat),
		Carb:     float32(carb),
		Fiber:    float32(fiber),
		Protein:  float32(protein),
	}, nil
}

func updateSettings(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	targets, err := parseMacroForm(c)
	if err != nil {
		return view.MacroTargets(db.Macro{
			Calories: db.CaloriesFromGrams(
				mustFloat(c.FormValue("fat")),
				mustFloat(c.FormValue("carb")),
				mustFloat(c.FormValue("protein")),
			),
			Fat:     float32(mustFloat(c.FormValue("fat"))),
			Carb:    float32(mustFloat(c.FormValue("carb"))),
			Fiber:   float32(mustFloat(c.FormValue("fiber"))),
			Protein: float32(mustFloat(c.FormValue("protein"))),
		}, view.SettingsCfg, err.Error()).Render(context.Background(), c.Response().Writer)
	}
	if err := db.UpdateUserTargets(userID, targets); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	c.Response().Header().Set("HX-Location", "/")
	return c.NoContent(http.StatusOK)
}

func onboardingView(c echo.Context) error {
	targets := presets["2000"]
	if p := c.QueryParam("preset"); p != "" {
		if preset, ok := presets[p]; ok {
			targets = preset
		}
	}

	if c.Request().Header.Get("HX-Request") != "" {
		return view.MacroTargets(targets, view.OnboardingCfg, "").Render(context.Background(), c.Response().Writer)
	}

	component := view.Full(view.Onboarding(targets))
	return component.Render(context.Background(), c.Response().Writer)
}

func saveOnboarding(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	targets, err := parseMacroForm(c)
	if err != nil {
		return view.MacroTargets(db.Macro{
			Calories: db.CaloriesFromGrams(
				mustFloat(c.FormValue("fat")),
				mustFloat(c.FormValue("carb")),
				mustFloat(c.FormValue("protein")),
			),
			Fat:     float32(mustFloat(c.FormValue("fat"))),
			Carb:    float32(mustFloat(c.FormValue("carb"))),
			Fiber:   float32(mustFloat(c.FormValue("fiber"))),
			Protein: float32(mustFloat(c.FormValue("protein"))),
		}, view.OnboardingCfg, err.Error()).Render(context.Background(), c.Response().Writer)
	}
	if err := db.UpdateUserTargets(userID, targets); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	c.Response().Header().Set("HX-Location", "/")
	return c.NoContent(http.StatusOK)
}

func queryDateUnix(c echo.Context) int64 {
	raw := c.QueryParam("date")
	if raw == "" {
		return 0
	}
	ii, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0
	}
	return ii
}

func requestedMealDate(c echo.Context) time.Time {
	if unix := queryDateUnix(c); unix != 0 {
		return time.Unix(unix, 0)
	}
	return time.Now()
}

func querySuffixForDate(c echo.Context) string {
	if unix := queryDateUnix(c); unix != 0 {
		return "?date=" + strconv.FormatInt(unix, 10)
	}
	return ""
}

func overviewPath(date time.Time) string {
	now := time.Now()
	if date.Year() == now.Year() && date.YearDay() == now.YearDay() {
		return "/"
	}
	return "/" + strconv.FormatInt(date.Unix(), 10)
}

func editPathForType(id int, mealType string) string {
	if mealType == "template" {
		return "/template/" + strconv.Itoa(id) + "/"
	}
	return "/meal/" + strconv.Itoa(id) + "/"
}

func mustFloat(raw string) float64 {
	value, _ := strconv.ParseFloat(strings.TrimSpace(raw), 32)
	if value < 0 {
		return 0
	}
	return value
}
