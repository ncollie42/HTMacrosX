package main

import (
	"context"
	db "myapp/DB"
	"myapp/auth"
	"myapp/view"
	"net/http"
	"strconv"

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
		return view.MacroTargets(targets, view.SettingsCfg).Render(context.Background(), c.Response().Writer)
	}

	nav := view.NavBack(userID, "/", "Settings")
	settingsForm := view.MacroTargets(targets, view.SettingsCfg)
	component := view.Full(nav, settingsForm)
	return component.Render(context.Background(), c.Response().Writer)
}

func parseMacroForm(c echo.Context) db.Macro {
	fat, _ := strconv.ParseFloat(c.FormValue("fat"), 32)
	carb, _ := strconv.ParseFloat(c.FormValue("carb"), 32)
	fiber, _ := strconv.ParseFloat(c.FormValue("fiber"), 32)
	protein, _ := strconv.ParseFloat(c.FormValue("protein"), 32)
	return db.Macro{
		Calories: db.CaloriesFromGrams(fat, carb, protein),
		Fat:      float32(fat),
		Carb:     float32(carb),
		Fiber:    float32(fiber),
		Protein:  float32(protein),
	}
}

func updateSettings(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	targets := parseMacroForm(c)
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
		return view.MacroTargets(targets, view.OnboardingCfg).Render(context.Background(), c.Response().Writer)
	}

	component := view.Full(view.Onboarding(targets))
	return component.Render(context.Background(), c.Response().Writer)
}

func saveOnboarding(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	targets := parseMacroForm(c)
	if err := db.UpdateUserTargets(userID, targets); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	c.Response().Header().Set("HX-Location", "/")
	return c.NoContent(http.StatusOK)
}
