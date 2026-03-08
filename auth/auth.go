package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	db "myapp/DB"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func Signup(c echo.Context, login string, pass string) error {
	userID, err := db.CreateUser(login, pass)
	if err != nil {
		return err
	}
	return setCookie(c, userID)
}

func Signin(c echo.Context, login string, pass string) error {
	userID, err := db.ValidateUser(login, pass)
	if err != nil {
		return fmt.Errorf("Invalid username or password")
	}
	return setCookie(c, userID)
}

const cookieName = "sessionID"

var ErrDupTokenMissing = errors.New("duplicate-submit token missing")
var ErrDupTokenMismatch = errors.New("duplicate-submit token mismatch")

func InitSession() {
	db.CleanExpiredSessions()
}

func setCookie(c echo.Context, userID int) error {
	sessionID, err := generateSessionID()
	if err != nil {
		return err
	}

	expiration := time.Now().Add(time.Hour * 24)
	if err := db.CreateSession(sessionID, userID, expiration); err != nil {
		return err
	}

	cookie := http.Cookie{
		Name:     cookieName,
		Value:    sessionID,
		Expires:  expiration,
		HttpOnly: true,
		Secure:   isSecureRequest(c),
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
	c.SetCookie(&cookie)
	return nil
}

func ClearCookie(c echo.Context) {
	cookie, err := c.Cookie(cookieName)
	if err != nil {
		return
	}
	db.DeleteSession(cookie.Value)

	newCookie := http.Cookie{
		Name:     cookieName,
		Value:    "",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   isSecureRequest(c),
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
	c.SetCookie(&newCookie)
}

func GetUserFromCookie(c echo.Context) (int, error) {
	cookie, err := c.Cookie(cookieName)
	if err != nil {
		return 0, fmt.Errorf("No valid cookie")
	}
	return db.GetSessionUserID(cookie.Value)
}

func SetDupToken(c echo.Context, token string) string {
	cookie, err := c.Cookie(cookieName)
	if err != nil {
		return token
	}
	db.SetSessionToken(cookie.Value, token)
	return token
}

func ValidateDupToken(c echo.Context, token string) error {
	cookie, err := c.Cookie(cookieName)
	if err != nil {
		return ErrDupTokenMissing
	}
	if token == "" {
		return ErrDupTokenMissing
	}
	sessionToken, err := db.GetSessionToken(cookie.Value)
	if err != nil {
		return ErrDupTokenMissing
	}
	if sessionToken == "" {
		return ErrDupTokenMissing
	}
	if token != sessionToken {
		return ErrDupTokenMismatch
	}
	return nil
}

func ConsumeDupToken(c echo.Context, token string) error {
	if err := ValidateDupToken(c, token); err != nil {
		return err
	}
	ClearDupToken(c)
	return nil
}

func ClearDupToken(c echo.Context) {
	cookie, err := c.Cookie(cookieName)
	if err != nil {
		return
	}
	db.ClearSessionToken(cookie.Value)
}

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func GenToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func isSecureRequest(c echo.Context) bool {
	host := c.Request().Host
	if forwarded := strings.TrimSpace(c.Request().Header.Get("X-Forwarded-Host")); forwarded != "" {
		host = forwarded
	}
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return false
	}
	if c.Request().TLS != nil {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(c.Request().Header.Get(echo.HeaderXForwardedProto)), "https")
}
