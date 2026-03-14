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

func registerMealRoutes(e *echo.Echo) {
	e.DELETE("/meal/:id/", deleteMeal, validate)
	e.GET("/meal/:id/", findMealOrTemplate, validate)
	e.GET("/meal/:id/food_search", foodSearch, validate)

	// Shared meal/template handlers
	e.POST("/meal/:id/join/:foodID", addFood, validate)
	e.DELETE("/meal/:id/join/:itemID", removeFood, validate)
	e.PUT("/meal/:id/join/:itemID", updateGrams, validate)
	e.PUT("/meal/:id/name", updateName, validate)
}

func deleteMeal(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	if err := db.DeleteMeal(id, userID, false); err != nil {
		return handleDBErr(c, err)
	}

	c.Response().Header().Set("HX-Location", "/")
	return c.NoContent(http.StatusOK)
}

func findMealOrTemplate(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	return renderMealEdit(c, id)
}

func renderMealEdit(c echo.Context, id int) error {
	return renderMealEditForType(c, id, isSavedMeal(c))
}

func renderMealEditForType(c echo.Context, id int, isTemplate bool) error {
	userID := c.Get(ctxUserID).(int)
	meal, err := db.GetMealByID(id, userID, isTemplate)
	if err != nil {
		return handleDBErr(c, err)
	}

	totals := db.SumMealItemMacros(meal.Items)

	var backURL, title, placeholder string
	if isTemplate {
		backURL = "/template/"
		title = "Edit Saved Meal"
		placeholder = "Meal Name"
	} else {
		backURL = "/"
		title = "Edit Meal"
	}
	nav := view.NavBack(userID, backURL, title)
	mealEdit := view.MealEdit(meal, placeholder, totals)
	mealNav := view.MealEditNav(id, isTemplate)
	component := view.Full(nav, mealEdit, mealNav)
	return component.Render(context.Background(), c.Response().Writer)
}

func addFood(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	foodID, err := strconv.Atoi(c.Param("foodID"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	var mealID int
	if c.Param("id") == newMealParam {
		mealID, err = db.CreateMeal(defaultMealName, userID, false, requestedMealDate(c))
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
	} else {
		mealID, err = strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}
	}

	if err := db.CreateMealItem(mealID, foodID, 100, userID, isSavedMeal(c)); err != nil {
		return handleDBErr(c, err)
	}

	c.Response().Header().Set("HX-Replace-Url", editPath(c, mealID))
	c.Response().Header().Set("HX-Retarget", "body")
	c.Response().Header().Set("HX-Reswap", "innerHTML")
	return renderMealEdit(c, mealID)
}

func removeFood(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	mealID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	itemID, err := strconv.Atoi(c.Param("itemID"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	isTemplate := isSavedMeal(c)
	if err := db.DeleteMealItem(mealID, itemID, userID, isTemplate); err != nil {
		return handleDBErr(c, err)
	}
	meal, err := db.GetMealByID(mealID, userID, isTemplate)
	if err != nil {
		return handleDBErr(c, err)
	}
	totals := db.SumMealItemMacros(meal.Items)
	return view.MealTotalsOOB(totals).Render(context.Background(), c.Response().Writer)
}

func updateGrams(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	mealID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	itemID, err := strconv.Atoi(c.Param("itemID"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	grams, err := strconv.ParseFloat(c.FormValue("grams"), 64)
	if err != nil {
		c.Response().Header().Set("HX-Reswap", "none")
		return view.MealFeedbackOOB("Grams must be a number").Render(context.Background(), c.Response().Writer)
	}
	if grams <= 0 {
		c.Response().Header().Set("HX-Reswap", "none")
		return view.MealFeedbackOOB("Grams must be greater than 0").Render(context.Background(), c.Response().Writer)
	}
	isTemplate := isSavedMeal(c)
	if err := db.UpdateMealItem(mealID, itemID, grams, userID, isTemplate); err != nil {
		return handleDBErr(c, err)
	}
	meal, err := db.GetMealByID(mealID, userID, isTemplate)
	if err != nil {
		return handleDBErr(c, err)
	}
	for _, item := range meal.Items {
		if item.ItemID == itemID {
			if err := view.GramEdit(item).Render(context.Background(), c.Response().Writer); err != nil {
				return err
			}
			break
		}
	}
	if err := view.MealFeedbackOOB("").Render(context.Background(), c.Response().Writer); err != nil {
		return err
	}
	totals := db.SumMealItemMacros(meal.Items)
	return view.MealTotalsOOB(totals).Render(context.Background(), c.Response().Writer)
}

func updateName(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	name := normalizedMealName(c.FormValue("name"), isSavedMeal(c))
	if err := db.UpdateMealName(id, userID, isSavedMeal(c), name); err != nil {
		return handleDBErr(c, err)
	}
	return nil
}

func isSavedMeal(c echo.Context) bool {
	return strings.HasPrefix(c.Path(), "/template")
}

func editPath(c echo.Context, id int) string {
	if isSavedMeal(c) {
		return fmt.Sprintf("/template/%d/", id)
	}
	return fmt.Sprintf("/meal/%d/", id)
}

func normalizedMealName(name string, isTemplate bool) string {
	name = strings.TrimSpace(name)
	if name != "" {
		return name
	}
	if isTemplate {
		return db.DefaultSavedMealName
	}
	return defaultMealName
}

func addFoodByTarget(userID int, targetType string, targetID string, foodID int, mealDate time.Time) (int, error) {
	var mealID int
	var err error
	if targetType == "meal" && targetID == newMealParam {
		mealID, err = db.CreateMeal(defaultMealName, userID, false, mealDate)
		if err != nil {
			return 0, err
		}
	} else {
		mealID, err = strconv.Atoi(targetID)
		if err != nil {
			return 0, errors.New("invalid target")
		}
	}
	if err := db.CreateMealItem(mealID, foodID, 100, userID, targetType == "template"); err != nil {
		return 0, err
	}
	return mealID, nil
}

func createAndAddFoodByTarget(userID int, targetType string, targetID string, name string, fat float64, carb float64, fiber float64, protein float64, grams float64, mealDate time.Time) (int, error) {
	return createAndAddFoodByTargetWithBarcode(userID, targetType, targetID, name, fat, carb, fiber, protein, grams, "", mealDate)
}
