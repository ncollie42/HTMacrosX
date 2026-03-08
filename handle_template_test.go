package main

import (
	db "myapp/DB"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func setupMainTestDB(t *testing.T) int {
	t.Helper()
	db.Open(filepath.Join(t.TempDir(), "main.db"))
	userID, err := db.CreateUser("main-user", "password1")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	return userID
}

func testSession(t *testing.T, userID int, token string) string {
	t.Helper()
	sessionID := "session-" + token
	if err := db.CreateSession(sessionID, userID, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	return sessionID
}

func newTemplateRequest(method, path string, form url.Values, sessionID string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	body := strings.NewReader(form.Encode())
	req := httptest.NewRequest(method, path, body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.AddCookie(&http.Cookie{Name: "sessionID", Value: sessionID, Path: "/"})
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestTemplateToMealCreatesMeal(t *testing.T) {
	userID := setupMainTestDB(t)
	templateID, err := db.CreateMeal(db.DefaultSavedMealName, userID, true, time.Time{})
	if err != nil {
		t.Fatalf("CreateMeal: %v", err)
	}
	sessionID := testSession(t, userID, "valid-token")

	form := url.Values{}
	ctx, rec := newTemplateRequest(http.MethodPost, "/template/1/", form, sessionID)
	ctx.SetParamNames("id")
	ctx.SetParamValues(strconv.Itoa(templateID))
	ctx.Set(ctxUserID, userID)

	if err := templateToMeal(ctx); err != nil {
		t.Fatalf("templateToMeal returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestTemplateToMealRejectsNormalMealID(t *testing.T) {
	userID := setupMainTestDB(t)
	mealID, err := db.CreateMeal("Lunch", userID, false, time.Now())
	if err != nil {
		t.Fatalf("CreateMeal: %v", err)
	}
	sessionID := testSession(t, userID, "valid-token")

	form := url.Values{"token": {"valid-token"}}
	ctx, rec := newTemplateRequest(http.MethodPost, "/template/1/", form, sessionID)
	ctx.SetParamNames("id")
	ctx.SetParamValues(strconv.Itoa(mealID))
	ctx.Set(ctxUserID, userID)

	if err := templateToMeal(ctx); err != nil {
		t.Fatalf("templateToMeal returned error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
