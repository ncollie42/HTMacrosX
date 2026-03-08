package main

import (
	"context"
	db "myapp/DB"
	"myapp/view"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

func registerOverviewRoutes(e *echo.Echo) {
	e.GET("/", overview, validate)
	e.GET("/:date", overview, validate)
}

func overview(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)

	target := db.GetUserTargets(userID)

	timeStr := c.Param("date")
	date := strconvTime(timeStr)
	macros := db.GetMealItemsByDate(userID, date)

	totalMacros := db.SumMacros(macros)
	macrosByID := db.SumMacrosByID(macros)

	nav := view.Nav(userID)
	overview := view.DayOverview(date, totalMacros, target)
	quickview := view.DayQuickview(macrosByID, date)
	bottomNav := view.BottomNav(date)
	component := view.Full(nav, overview, quickview, bottomNav)
	return component.Render(context.Background(), c.Response().Writer)
}

func strconvTime(num string) time.Time {
	if num == "" {
		return time.Now()
	}
	ii, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return time.Now()
	}
	return time.Unix(ii, 0)
}
