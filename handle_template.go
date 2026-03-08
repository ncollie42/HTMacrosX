package main

import (
	"context"
	"fmt"
	db "myapp/DB"
	"myapp/view"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

func registerTemplateRoutes(e *echo.Echo) {
	e.GET("/template/", findAllTemplates, validate)
	e.POST("/template/", createTemplate, validate)
	e.GET("/template/:id/", findMealOrTemplate, validate)
	e.GET("/template/:id/food_search", foodSearch, validate)
	e.DELETE("/template/:id/", deleteTemplate, validate)
	e.POST("/template/:id/", templateToMeal, validate)

	e.POST("/template/:id/join/:foodID", addFood, validate)
	e.DELETE("/template/:id/join/:itemID", removeFood, validate)
	e.PUT("/template/:id/join/:itemID", updateGrams, validate)
	e.PUT("/template/:id/name", updateName, validate)
}

func templateToMeal(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	templateID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if _, err := db.TemplateToMeal(templateID, userID, requestedMealDate(c)); err != nil {
		return handleDBErr(c, err)
	}

	c.Response().Header().Set("HX-Location", "/")
	return c.NoContent(http.StatusOK)
}

func findAllTemplates(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	macros := db.GetTemplates(userID)
	macrosByID := db.SumMacrosByID(macros)

	nav := view.NavBack(userID, "/", "Saved Meals")
	overview := view.TemplateOverview(macrosByID)
	component := view.Full(nav, overview)
	return component.Render(context.Background(), c.Response().Writer)
}

func createTemplate(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)

	templateID, err := db.CreateMeal(db.DefaultSavedMealName, userID, true, time.Time{})
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	c.Response().Header().Set("HX-Location", fmt.Sprint("/template/", templateID, "/"))
	return c.NoContent(http.StatusOK)
}

func deleteTemplate(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	if err := db.DeleteMeal(id, userID, true); err != nil {
		return handleDBErr(c, err)
	}
	return c.NoContent(http.StatusOK)
}
