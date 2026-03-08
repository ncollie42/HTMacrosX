package main

import (
	"context"
	"errors"
	"fmt"
	db "myapp/DB"
	"myapp/view"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func registerFoodRoutes(e *echo.Echo) {
	e.POST("/food", createFood, validate)
	e.DELETE("/food/:id", deleteFood, validate)
}

func foodSearch(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	search := c.FormValue("search")
	foods := db.FoodSearch(search, userID)
	searchPath, joinPrefix, querySuffix, targetType, targetID, dateUnix := foodSearchPaths(c)

	var backURL string
	if c.Param("id") == newMealParam {
		backURL = overviewPath(requestedMealDate(c))
	} else {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
		backURL = editPath(c, id)
	}
	nav := view.NavBack(userID, backURL, "Ingredients")
	searchResult := view.FoodSearch(foods, searchPath, joinPrefix, querySuffix, targetType, targetID, dateUnix)
	component := view.Full(nav, searchResult)
	return component.Render(context.Background(), c.Response().Writer)
}

func createFood(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	name := strings.TrimSpace(c.FormValue("name"))
	fat, err := parseNonNegativeFloat(c.FormValue("fat"), "Fat")
	if err != nil {
		c.Response().Header().Set("HX-Reswap", "none")
		return view.FoodFeedbackOOB(err.Error()).Render(context.Background(), c.Response().Writer)
	}
	carb, err := parseNonNegativeFloat(c.FormValue("carb"), "Carbs")
	if err != nil {
		c.Response().Header().Set("HX-Reswap", "none")
		return view.FoodFeedbackOOB(err.Error()).Render(context.Background(), c.Response().Writer)
	}
	fiber, err := parseNonNegativeFloat(c.FormValue("fiber"), "Fiber")
	if err != nil {
		c.Response().Header().Set("HX-Reswap", "none")
		return view.FoodFeedbackOOB(err.Error()).Render(context.Background(), c.Response().Writer)
	}
	protein, err := parseNonNegativeFloat(c.FormValue("protein"), "Protein")
	if err != nil {
		c.Response().Header().Set("HX-Reswap", "none")
		return view.FoodFeedbackOOB(err.Error()).Render(context.Background(), c.Response().Writer)
	}
	grams, err := parsePositiveFloat(c.FormValue("grams"), "Serving size")
	if err != nil {
		c.Response().Header().Set("HX-Reswap", "none")
		return view.FoodFeedbackOOB(err.Error()).Render(context.Background(), c.Response().Writer)
	}
	if name == "" {
		c.Response().Header().Set("HX-Reswap", "none")
		return view.FoodFeedbackOOB("Name is required").Render(context.Background(), c.Response().Writer)
	}

	targetType := c.FormValue("targetType")
	targetID := c.FormValue("targetID")
	if c.FormValue("autoAdd") == "1" && targetType != "" && targetID != "" {
		mealDate := requestedMealDate(c)
		if unix := parseDateUnixValue(c.FormValue("dateUnix")); unix != 0 {
			mealDate = time.Unix(unix, 0)
		}
		mealID, err := createAndAddFoodByTarget(userID, targetType, targetID, name, fat, carb, fiber, protein, grams, mealDate)
		if err != nil {
			return handleDBErr(c, err)
		}
		c.Response().Header().Set("HX-Replace-Url", editPathForType(mealID, targetType))
		c.Response().Header().Set("HX-Retarget", "body")
		c.Response().Header().Set("HX-Reswap", "innerHTML")
		return renderMealEditForType(c, mealID, targetType == "template")
	}

	if _, err := db.CreateFood(name, fat, carb, fiber, protein, grams, userID); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	foods := db.FoodSearch("", userID)
	joinPrefix := c.FormValue("joinPrefix")
	querySuffix := c.FormValue("querySuffix")
	component := view.FoodSearchResults(foods, joinPrefix, querySuffix)
	if err := component.Render(context.Background(), c.Response().Writer); err != nil {
		return err
	}
	return view.FoodFeedbackOOB("").Render(context.Background(), c.Response().Writer)
}

func deleteFood(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	foodID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if err := db.DeleteFood(foodID, userID); err != nil {
		if errors.Is(err, db.ErrFoodInUse) {
			c.Response().Header().Set("HX-Reswap", "none")
			return view.FoodFeedbackOOB("Can't delete this ingredient while it's used in meals or saved meals").Render(context.Background(), c.Response().Writer)
		}
		return handleDBErr(c, err)
	}
	if err := view.FoodFeedbackOOB("").Render(context.Background(), c.Response().Writer); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

func foodSearchPaths(c echo.Context) (string, string, string, string, string, int64) {
	basePath := c.Request().URL.Path
	querySuffix := querySuffixForDate(c)
	targetType := "meal"
	if strings.HasPrefix(basePath, "/template/") {
		targetType = "template"
	}
	return basePath + querySuffix, strings.TrimSuffix(basePath, "food_search") + "join/", querySuffix, targetType, c.Param("id"), queryDateUnix(c)
}

func parsePositiveFloat(raw string, label string) (float64, error) {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number", label)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be greater than 0", label)
	}
	return value, nil
}

func parseNonNegativeFloat(raw string, label string) (float64, error) {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number", label)
	}
	if value < 0 {
		return 0, fmt.Errorf("%s cannot be negative", label)
	}
	return value, nil
}

func parseDateUnixValue(raw string) int64 {
	parsed, _ := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	return parsed
}
