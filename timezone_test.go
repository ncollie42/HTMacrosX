package main

import (
	db "myapp/DB"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func newSettingsContext(method, path string, form url.Values) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	body := strings.NewReader(form.Encode())
	req := httptest.NewRequest(method, path, body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestValidatePersistsInitialBrowserTimezoneOnce(t *testing.T) {
	userID := setupMainTestDB(t)
	sessionID := testSession(t, userID, "timezone-bootstrap")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "sessionID", Value: sessionID, Path: "/"})
	req.AddCookie(&http.Cookie{Name: "tz", Value: url.QueryEscape("America/New_York"), Path: "/"})
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	handler := validate(func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	if err := handler(ctx); err != nil {
		t.Fatalf("validate first call: %v", err)
	}

	timezone, err := db.GetUserTimezone(userID)
	if err != nil {
		t.Fatalf("GetUserTimezone first call: %v", err)
	}
	if timezone != "America/New_York" {
		t.Fatalf("timezone = %q, want %q", timezone, "America/New_York")
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.AddCookie(&http.Cookie{Name: "sessionID", Value: sessionID, Path: "/"})
	req2.AddCookie(&http.Cookie{Name: "tz", Value: url.QueryEscape("Europe/Berlin"), Path: "/"})
	rec2 := httptest.NewRecorder()
	ctx2 := e.NewContext(req2, rec2)
	if err := handler(ctx2); err != nil {
		t.Fatalf("validate second call: %v", err)
	}

	timezone, err = db.GetUserTimezone(userID)
	if err != nil {
		t.Fatalf("GetUserTimezone second call: %v", err)
	}
	if timezone != "America/New_York" {
		t.Fatalf("timezone after second call = %q, want unchanged", timezone)
	}
}

func TestUpdateSettingsValidatesAndPersistsTimezone(t *testing.T) {
	userID := setupMainTestDB(t)

	invalidForm := url.Values{
		"fat":      {"50"},
		"carb":     {"220"},
		"fiber":    {"30"},
		"protein":  {"145"},
		"timezone": {"PST"},
	}
	ctx, rec := newSettingsContext(http.MethodPut, "/settings", invalidForm)
	ctx.Set(ctxUserID, userID)

	if err := updateSettings(ctx); err != nil {
		t.Fatalf("updateSettings invalid timezone: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("invalid status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Timezone must be a valid IANA timezone") {
		t.Fatalf("invalid body = %q", rec.Body.String())
	}

	validForm := url.Values{
		"fat":      {"50"},
		"carb":     {"220"},
		"fiber":    {"30"},
		"protein":  {"145"},
		"timezone": {"America/Los_Angeles"},
	}
	ctx, rec = newSettingsContext(http.MethodPut, "/settings", validForm)
	ctx.Set(ctxUserID, userID)

	if err := updateSettings(ctx); err != nil {
		t.Fatalf("updateSettings valid: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("valid status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Header().Get("HX-Location") != "/" {
		t.Fatalf("HX-Location = %q, want /", rec.Header().Get("HX-Location"))
	}
	timezone, err := db.GetUserTimezone(userID)
	if err != nil {
		t.Fatalf("GetUserTimezone: %v", err)
	}
	if timezone != "America/Los_Angeles" {
		t.Fatalf("timezone = %q, want %q", timezone, "America/Los_Angeles")
	}
}

func TestParseRequestedDayAcceptsLegacyUnixInUserTimezone(t *testing.T) {
	userID := setupMainTestDB(t)
	if err := db.UpdateUserTimezone(userID, "America/New_York"); err != nil {
		t.Fatalf("UpdateUserTimezone: %v", err)
	}

	e := echo.New()
	instant := time.Date(2026, 3, 8, 1, 30, 0, 0, time.UTC)
	req := httptest.NewRequest(http.MethodGet, "/?date="+strconv.FormatInt(instant.Unix(), 10), nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.Set(ctxUserID, userID)

	got := parseRequestedDay(ctx)
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}
	want := time.Date(2026, 3, 7, 12, 0, 0, 0, loc)
	if !got.Equal(want) {
		t.Fatalf("parseRequestedDay = %v, want %v", got, want)
	}
	if queryDayValue(ctx) != "2026-03-07" {
		t.Fatalf("queryDayValue = %q, want %q", queryDayValue(ctx), "2026-03-07")
	}
}

func TestCurrentLocalDayUsesSavedTimezone(t *testing.T) {
	userID := setupMainTestDB(t)
	if err := db.UpdateUserTimezone(userID, "America/Los_Angeles"); err != nil {
		t.Fatalf("UpdateUserTimezone: %v", err)
	}

	originalNow := timeNow
	t.Cleanup(func() { timeNow = originalNow })
	timeNow = func() time.Time {
		return time.Date(2026, 3, 8, 1, 30, 0, 0, time.UTC)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.Set(ctxUserID, userID)

	got := parseRequestedDay(ctx)
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf("LoadLocation: %v", err)
	}
	want := time.Date(2026, 3, 7, 12, 0, 0, 0, loc)
	if !got.Equal(want) {
		t.Fatalf("parseRequestedDay current day = %v, want %v", got, want)
	}
}
