package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	db "myapp/DB"
	"net/http"
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
		Name:    cookieName,
		Value:   sessionID,
		Expires: expiration,
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
		Name:   cookieName,
		Value:  "",
		MaxAge: -1,
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

func ValidateDupToken(c echo.Context, token string) bool {
	cookie, err := c.Cookie(cookieName)
	if err != nil {
		return false
	}
	sessionToken, err := db.GetSessionToken(cookie.Value)
	if err != nil {
		return false
	}
	return token == sessionToken
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
