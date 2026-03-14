package main

import (
	"context"
	"errors"
	"fmt"
	db "myapp/DB"
	"myapp/foodsource"
	"myapp/view"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func registerScanRoutes(e *echo.Echo) {
	e.GET("/scan", scanView, validate)
	e.GET("/scan/confirm", scanConfirmView, validate)
	e.POST("/scan/confirm", scanConfirmAdd, validate)
	e.POST("/scan", scanBarcodeManual, validate)
	e.POST("/scan/:barcode", scanBarcode, validate)
}

func scanView(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	mealID, _ := strconv.Atoi(c.QueryParam("meal"))
	mealType := c.QueryParam("type")
	dateValue := queryDayValue(c)
	backURL := "/"
	if mealID != 0 {
		backURL = editPathForType(mealID, mealType)
	} else {
		backURL = overviewPath(c, requestedMealDate(c))
	}
	nav := view.NavBack(userID, backURL, "Scan")
	scan := view.ScanPage(mealID, mealType, dateValue, scanSearchPath(mealID, mealType, dateValue), c.QueryParam("error"))
	component := view.Full(nav, scan)
	return component.Render(context.Background(), c.Response().Writer)
}

func scanConfirmView(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	mealID, _ := strconv.Atoi(c.QueryParam("meal"))
	mealType := c.QueryParam("type")
	dateValue := queryDayValue(c)
	barcode := c.QueryParam("barcode")
	if barcode == "" {
		return c.Redirect(http.StatusSeeOther, "/scan")
	}

	food, err := previewBarcodedFood(barcode, userID)
	if err != nil {
		if errors.Is(err, foodsource.ErrNotFound) {
			return c.Redirect(http.StatusSeeOther, scanPathWithQuery(mealID, mealType, dateValue, "Product not found"))
		}
		return c.Redirect(http.StatusSeeOther, scanPathWithQuery(mealID, mealType, dateValue, "Lookup failed. Please try again"))
	}

	backURL := "/"
	if mealID != 0 {
		backURL = editPathForType(mealID, mealType)
	} else {
		backURL = overviewPath(c, requestedMealDate(c))
	}
	nav := view.NavBack(userID, backURL, "Confirm Scan")
	component := view.Full(nav, view.ScanConfirm(food, barcode, mealID, mealType, dateValue, scanSearchPath(mealID, mealType, dateValue)))
	return component.Render(context.Background(), c.Response().Writer)
}

func scanBarcode(c echo.Context) error {
	if strings.TrimSpace(c.Param("barcode")) == "" {
		return c.NoContent(http.StatusBadRequest)
	}
	mealID, _ := strconv.Atoi(c.QueryParam("meal"))
	mealType := c.QueryParam("type")
	dateValue := queryDayValue(c)
	c.Response().Header().Set("HX-Location", scanConfirmPath(c.Param("barcode"), mealID, mealType, dateValue))
	return c.NoContent(http.StatusOK)
}

func scanBarcodeManual(c echo.Context) error {
	barcode := strings.TrimSpace(c.FormValue("barcode"))
	if barcode == "" {
		mealID, _ := strconv.Atoi(c.FormValue("meal"))
		mealType := c.FormValue("type")
		dateValue := canonicalDayValue(c.FormValue("date"), loadUserLocation(c))
		return c.Redirect(http.StatusSeeOther, scanPathWithQuery(mealID, mealType, dateValue, "Enter a barcode to continue"))
	}
	mealID, _ := strconv.Atoi(c.FormValue("meal"))
	mealType := c.FormValue("type")
	dateValue := canonicalDayValue(c.FormValue("date"), loadUserLocation(c))
	return c.Redirect(http.StatusSeeOther, scanConfirmPath(barcode, mealID, mealType, dateValue))
}

func scanSearchPath(mealID int, mealType string, dateValue string) string {
	if mealID == 0 {
		return addDateQuery("/meal/new/food_search", dateValue)
	}
	if mealType == "template" {
		return fmt.Sprintf("/template/%d/food_search", mealID)
	}
	return fmt.Sprintf("/meal/%d/food_search", mealID)
}

func scanConfirmPath(barcode string, mealID int, mealType string, dateValue string) string {
	path := fmt.Sprintf("/scan/confirm?barcode=%s", url.QueryEscape(barcode))
	if mealID != 0 {
		path += fmt.Sprintf("&meal=%d", mealID)
	}
	if mealType != "" {
		path += "&type=" + url.QueryEscape(mealType)
	}
	if dateValue != "" {
		path += "&date=" + url.QueryEscape(dateValue)
	}
	return path
}

func scanPathWithQuery(mealID int, mealType string, dateValue string, errMsg string) string {
	path := "/scan"
	var params []string
	if mealID != 0 {
		params = append(params, fmt.Sprintf("meal=%d", mealID))
	}
	if mealType != "" {
		params = append(params, "type="+mealType)
	}
	if dateValue != "" {
		params = append(params, "date="+url.QueryEscape(dateValue))
	}
	if errMsg != "" {
		params = append(params, "error="+url.QueryEscape(errMsg))
	}
	if len(params) > 0 {
		path += "?" + strings.Join(params, "&")
	}
	return path
}

func previewBarcodedFood(barcode string, userID int) (db.Food, error) {
	return previewExternalFood(barcode, userID)
}

func scanConfirmAdd(c echo.Context) error {
	userID := c.Get(ctxUserID).(int)
	barcode := strings.TrimSpace(c.FormValue("barcode"))
	if barcode == "" {
		return c.NoContent(http.StatusBadRequest)
	}
	mealID, _ := strconv.Atoi(c.FormValue("meal"))
	mealType := c.FormValue("type")
	dateValue := canonicalDayValue(c.FormValue("date"), loadUserLocation(c))
	mealDate := requestedMealDate(c)
	if day, ok := parseDayValue(dateValue, loadUserLocation(c)); ok {
		mealDate = day
	}
	mealID, err := addExternalFoodByTarget(userID, defaultTargetType(mealType), targetIDForMeal(mealID), barcode, mealDate)
	if err != nil {
		if errors.Is(err, foodsource.ErrNotFound) {
			return c.Redirect(http.StatusSeeOther, scanPathWithQuery(mealID, mealType, dateValue, "Product not found"))
		}
		return c.Redirect(http.StatusSeeOther, scanPathWithQuery(mealID, mealType, dateValue, "Lookup failed. Please try again"))
	}
	return c.Redirect(http.StatusSeeOther, editPathForType(mealID, mealType))
}
