package main

import (
	"context"
	"fmt"
	db "myapp/DB"
	"myapp/auth"
	"myapp/view"
	"net/http"
	"net/url"
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
	if err := validateAuthPost(c); err != nil {
		component := view.SignError(err.Error())
		return component.Render(context.Background(), c.Response().Writer)
	}
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
	if err := validateAuthPost(c); err != nil {
		component := view.SignError(err.Error())
		return component.Render(context.Background(), c.Response().Writer)
	}
	login := c.FormValue("login")
	password := c.FormValue("password")
	confirm := c.FormValue("confirm")

	if strings.TrimSpace(login) == "" {
		component := view.SignError("Username is required")
		return component.Render(context.Background(), c.Response().Writer)
	}
	if len(password) < 8 {
		component := view.SignError("Password must be at least 8 characters")
		return component.Render(context.Background(), c.Response().Writer)
	}
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
	timezone, err := db.GetUserTimezone(userID)
	if err != nil {
		return handleDBErr(c, err)
	}
	if timezone == "" {
		if browser := browserTimezone(c); browser != "" {
			timezone = browser
		} else {
			timezone = defaultAppTimezone
		}
	}

	targets, ok := presets[c.QueryParam("preset")]
	if !ok {
		targets, err = db.GetUserTargets(userID)
		if err != nil {
			return handleDBErr(c, err)
		}
	}
	form := view.NewMacroTargetsForm(targets)

	if c.Request().Header.Get("HX-Request") != "" {
		return view.MacroTargets(targets, form, view.SettingsCfg, timezone, "").Render(context.Background(), c.Response().Writer)
	}

	nav := view.NavBack(userID, "/", "Settings")
	settingsForm := view.MacroTargets(targets, form, view.SettingsCfg, timezone, "")
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
	timezone, timezoneErr := requireTimezone(c.FormValue("timezone"))
	form := view.NewMacroTargetsFormFromRaw(
		c.FormValue("fat"),
		c.FormValue("carb"),
		c.FormValue("fiber"),
		c.FormValue("protein"),
	)
	if err != nil {
		return view.MacroTargets(form.ToMacro(), form, view.SettingsCfg, c.FormValue("timezone"), err.Error()).Render(context.Background(), c.Response().Writer)
	}
	if timezoneErr != nil {
		return view.MacroTargets(targets, form, view.SettingsCfg, c.FormValue("timezone"), timezoneErr.Error()).Render(context.Background(), c.Response().Writer)
	}
	if err := db.UpdateUserTargets(userID, targets); err != nil {
		return handleDBErr(c, err)
	}
	if err := db.UpdateUserTimezone(userID, timezone); err != nil {
		return handleDBErr(c, err)
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
	form := view.NewMacroTargetsForm(targets)

	if c.Request().Header.Get("HX-Request") != "" {
		return view.MacroTargets(targets, form, view.OnboardingCfg, "", "").Render(context.Background(), c.Response().Writer)
	}

	component := view.Full(view.Onboarding(targets, form))
	return component.Render(context.Background(), c.Response().Writer)
}

func saveOnboarding(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	targets, err := parseMacroForm(c)
	form := view.NewMacroTargetsFormFromRaw(
		c.FormValue("fat"),
		c.FormValue("carb"),
		c.FormValue("fiber"),
		c.FormValue("protein"),
	)
	if err != nil {
		return view.MacroTargets(form.ToMacro(), form, view.OnboardingCfg, "", err.Error()).Render(context.Background(), c.Response().Writer)
	}
	if err := db.UpdateUserTargets(userID, targets); err != nil {
		return handleDBErr(c, err)
	}

	c.Response().Header().Set("HX-Location", "/")
	return c.NoContent(http.StatusOK)
}

func requestedMealDate(c echo.Context) time.Time {
	return parseRequestedDay(c)
}

func querySuffixForDate(c echo.Context) string {
	return querySuffixForDay(c)
}

func overviewPath(c echo.Context, date time.Time) string {
	return overviewPathForDay(date, loadUserLocation(c))
}

func editPathForType(id int, mealType string) string {
	if mealType == "template" {
		return "/template/" + strconv.Itoa(id) + "/"
	}
	return "/meal/" + strconv.Itoa(id) + "/"
}

func validateAuthPost(c echo.Context) error {
	if secFetchSite := strings.TrimSpace(c.Request().Header.Get("Sec-Fetch-Site")); strings.EqualFold(secFetchSite, "cross-site") {
		return fmt.Errorf("Cross-site auth requests are not allowed")
	}
	origin := strings.TrimSpace(c.Request().Header.Get("Origin"))
	if origin == "" {
		return nil
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" {
		return fmt.Errorf("Invalid auth request origin")
	}
	if !strings.EqualFold(parsed.Host, c.Request().Host) {
		return fmt.Errorf("Cross-site auth requests are not allowed")
	}
	return nil
}
