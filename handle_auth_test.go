package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func newAuthFormContext(method, path string, form url.Values) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	body := strings.NewReader(form.Encode())
	req := httptest.NewRequest(method, path, body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Host = "example.com"
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestSignupRejectsCrossSiteOrigin(t *testing.T) {
	form := url.Values{
		"login":    {"alice"},
		"password": {"password1"},
		"confirm":  {"password1"},
	}
	ctx, rec := newAuthFormContext(http.MethodPost, "/signup", form)
	ctx.Request().Header.Set("Origin", "https://evil.example")

	if err := signup(ctx); err != nil {
		t.Fatalf("signup returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Cross-site auth requests are not allowed") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestSignupRejectsShortPassword(t *testing.T) {
	form := url.Values{
		"login":    {"alice"},
		"password": {"short"},
		"confirm":  {"short"},
	}
	ctx, rec := newAuthFormContext(http.MethodPost, "/signup", form)
	ctx.Request().Header.Set("Origin", "http://example.com")

	if err := signup(ctx); err != nil {
		t.Fatalf("signup returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Password must be at least 8 characters") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}
