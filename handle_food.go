package main

import (
	"context"
	db "myapp/DB"
	"myapp/view"
	"net/http"
	"strconv"

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

	id, _ := strconv.Atoi(c.Param("id"))
	nav := view.NavBack(userID, editPath(c, id), "Ingredients")
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
