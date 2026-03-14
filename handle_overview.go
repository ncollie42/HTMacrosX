package main

import (
	"context"
	db "myapp/DB"
	"myapp/view"

	"github.com/labstack/echo/v4"
)

func registerOverviewRoutes(e *echo.Echo) {
	e.GET("/", overview, validate)
	e.GET("/:date", overview, validate)
}

func overview(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	loc := loadUserLocation(c)
	date := parseRequestedDay(c)

	target, err := db.GetUserTargets(userID)
	if err != nil {
		return handleDBErr(c, err)
	}
	macros := db.GetMealItemsByDate(userID, date)

	totalMacros := db.SumMacros(macros)
	macrosByID := db.SumMacrosByID(macros)
	label := dayLabel(date, loc)
	prevPath := overviewPathForDay(date.AddDate(0, 0, -1), loc)
	nextPath := overviewPathForDay(date.AddDate(0, 0, 1), loc)
	scanPath := addDateQuery("/scan", dayQueryValue(date, loc))
	addFoodPath := addDateQuery("/meal/new/food_search", dayQueryValue(date, loc))

	nav := view.Nav(userID)
	overview := view.DayOverview(totalMacros, target, prevPath, "/", nextPath, label)
	quickview := view.DayQuickview(macrosByID, label)
	bottomNav := view.BottomNav(scanPath, addFoodPath)
	component := view.Full(nav, overview, quickview, bottomNav)
	return component.Render(context.Background(), c.Response().Writer)
}
