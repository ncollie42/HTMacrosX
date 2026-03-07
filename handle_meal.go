package main

import (
	"context"
	"fmt"
	db "myapp/DB"
	"myapp/view"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func registerMealRoutes(e *echo.Echo) {
	e.POST("/meal", createMeal, validate)
	e.DELETE("/meal/:id/", deleteMeal, validate)
	e.GET("/meal/:id/", findMealOrTemplate, validate)
	e.GET("/meal/:id/food_search", foodSearch, validate)

	// Shared meal/template handlers
	e.POST("/meal/:id/join/:foodID", addFood, validate)
	e.DELETE("/meal/:id/join/:itemID", removeFood, validate)
	e.PUT("/meal/:id/join/:itemID", updateGrams, validate)
	e.PUT("/meal/:id/name", updateName, validate)
}

func createMeal(c echo.Context) error {
	// NOTE: this will create a empty meal entries, will probably want a way to clean it up in the future.
	userID := c.Get(ctxUserID).(int)

	mealID, err := db.CreateMeal(defaultMealName, userID, false)
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

func findMealOrTemplate(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	return renderMealEdit(c, id)
}

func renderMealEdit(c echo.Context, id int) error {
	userID := c.Get(ctxUserID).(int)
	meal, err := db.GetMealByID(id, userID)
	if err != nil {
		return handleDBErr(c, err)
	}

	totals := db.SumMealItemMacros(meal.Items)

	var backURL, title, placeholder string
	if isSavedMeal(c) {
		backURL = "/template/"
		title = "Edit Saved Meal"
		placeholder = "Meal Name"
	} else {
		backURL = "/"
		title = "Edit Meal"
	}
	nav := view.NavBack(userID, backURL, title)
	mealEdit := view.MealEdit(meal, placeholder, totals)
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
	if err := db.DeleteMealItem(itemID, userID); err != nil {
		return handleDBErr(c, err)
	}
	meal, err := db.GetMealByID(mealID, userID)
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
		return c.NoContent(http.StatusBadRequest)
	}
	if err := db.UpdateMealItem(itemID, grams, userID); err != nil {
		return handleDBErr(c, err)
	}
	meal, err := db.GetMealByID(mealID, userID)
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
	totals := db.SumMealItemMacros(meal.Items)
	return view.MealTotalsOOB(totals).Render(context.Background(), c.Response().Writer)
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

func isSavedMeal(c echo.Context) bool {
	return strings.HasPrefix(c.Path(), "/template")
}

func editPath(c echo.Context, id int) string {
	if isSavedMeal(c) {
		return fmt.Sprintf("/template/%d/", id)
	}
	return fmt.Sprintf("/meal/%d/", id)
}
