package main

import (
	db "myapp/DB"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCreateFoodAutoAddDoesNotPersistForMissingTarget(t *testing.T) {
	userID := setupMainTestDB(t)

	form := url.Values{
		"name":       {"Test Food"},
		"fat":        {"1"},
		"carb":       {"2"},
		"fiber":      {"0"},
		"protein":    {"3"},
		"grams":      {"100"},
		"autoAdd":    {"1"},
		"targetType": {"meal"},
		"targetID":   {"9999"},
	}
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/food", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.Set(ctxUserID, userID)

	if err := createFood(ctx); err != nil {
		t.Fatalf("createFood returned error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if foods := db.FoodSearch("Test Food", userID); len(foods) != 0 {
		t.Fatalf("food persisted on failed auto-add: %+v", foods)
	}
}
