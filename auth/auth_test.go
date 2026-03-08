package auth

import (
	db "myapp/DB"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func setupAuthTestDB(t *testing.T) int {
	t.Helper()
	db.Open(filepath.Join(t.TempDir(), "auth.db"))
	userID, err := db.CreateUser("auth-user", "password1")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	return userID
}

func newAuthContext(t *testing.T, sessionID string) echo.Context {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: sessionID, Path: "/"})
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestConsumeDupToken(t *testing.T) {
	userID := setupAuthTestDB(t)
	sessionID := "session-1"
	if err := db.CreateSession(sessionID, userID, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	ctx := newAuthContext(t, sessionID)
	if err := db.SetSessionToken(sessionID, "abc123"); err != nil {
		t.Fatalf("SetSessionToken: %v", err)
	}
	if err := ConsumeDupToken(ctx, "abc123"); err != nil {
		t.Fatalf("ConsumeDupToken valid: %v", err)
	}
	token, err := db.GetSessionToken(sessionID)
	if err != nil {
		t.Fatalf("GetSessionToken: %v", err)
	}
	if token != "" {
		t.Fatalf("token after consume = %q, want empty", token)
	}

	if err := ConsumeDupToken(ctx, ""); err != ErrDupTokenMissing {
		t.Fatalf("ConsumeDupToken empty err = %v, want ErrDupTokenMissing", err)
	}
	if err := db.SetSessionToken(sessionID, "new-token"); err != nil {
		t.Fatalf("SetSessionToken mismatch: %v", err)
	}
	if err := ConsumeDupToken(ctx, "wrong-token"); err != ErrDupTokenMismatch {
		t.Fatalf("ConsumeDupToken mismatch err = %v, want ErrDupTokenMismatch", err)
	}
}

func TestSetCookieUsesForwardedHTTPS(t *testing.T) {
	userID := setupAuthTestDB(t)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/signin", strings.NewReader(""))
	req.Header.Set(echo.HeaderXForwardedProto, "https")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := setCookie(ctx, userID); err != nil {
		t.Fatalf("setCookie: %v", err)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	if !cookies[0].Secure {
		t.Fatalf("cookie Secure = false, want true")
	}
}

func TestClearCookieMatchesOriginalAttributes(t *testing.T) {
	userID := setupAuthTestDB(t)
	sessionID := "session-clear"
	if err := db.CreateSession(sessionID, userID, time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/signout", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: sessionID, Path: "/"})
	req.Header.Set(echo.HeaderXForwardedProto, "https")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	ClearCookie(ctx)
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %d, want 1", len(cookies))
	}
	got := cookies[0]
	if got.Path != "/" || !got.HttpOnly || !got.Secure {
		t.Fatalf("clear cookie attrs = %+v", got)
	}
	if got.MaxAge != -1 {
		t.Fatalf("clear cookie MaxAge = %d, want -1", got.MaxAge)
	}
}
