package auth

import (
	"crypto/md5"
	"crypto/rand"
	"io"
	"strconv"

	"encoding/base64"
	"fmt"
	db "myapp/DB"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func Signin(c echo.Context, login string, pass string) error {
	userID, err := db.ValidateUser(login, pass)
	if err != nil {
		fmt.Println("Err:", err)
		return fmt.Errorf("Invalid username or password")
	}

	setCookie(c, userID)
	return nil
}

// ---------------------- Generic --------------------------------

const userID_tag = "userID"

func InitSession() {
	// Nothing needed for local sessions
}

func setCookie(c echo.Context, userID string) {
	sessionID, err := generateSessionID()
	if err != nil {
		panic(err)
	}

	setVarLocal(sessionID, userID_tag, userID)

	expiration := time.Now().Add(time.Hour * 24)
	cookie := http.Cookie{
		Name:    cookieName,
		Value:   sessionID,
		Expires: expiration,
	}
	c.SetCookie(&cookie)
}

func ClearCookie(c echo.Context) {
	cookie, err := c.Cookie(cookieName)
	if err != nil {
		return
	}

	clearSessionIDLocal(cookie.Value)

	newCookie := http.Cookie{
		Name:   cookieName,
		Value:  "",
		MaxAge: -1,
	}
	c.SetCookie(&newCookie)
}

func GetUserFromCookie(c echo.Context) (int, error) {
	userID_str, err := getVarLocal(c, userID_tag)
	if err != nil {
		return 0, err
	}
	userID_int, err := strconv.ParseInt(userID_str, 10, 64)
	return int(userID_int), err
}

func SetDupToken(c echo.Context, token string) string {
	sessionID, _ := getSessionID(c)
	setVarLocal(sessionID, "token", token)
	return token
}

func ValidateDupToken(c echo.Context, token string) bool {
	sessionToken, _ := getVarLocal(c, "token")
	return token == sessionToken
}

func ClearDupToken(c echo.Context) {
	clearVarLocal(c, "token")
}

func getSessionID(c echo.Context) (string, error) {
	cookie, err := c.Cookie(cookieName)
	if err != nil {
		return "", fmt.Errorf("No valid cookie")
	}
	return cookie.Value, nil
}

// ---------------------- Sessions Local --------------------------------
var sessions = make(map[string]map[string]string)

const cookieName = "sessionID"

func setVarLocal(sessionID string, name string, val string) {
	if _, exists := sessions[sessionID]; !exists {
		sessions[sessionID] = make(map[string]string)
	}
	fmt.Println("SessionID:", sessionID)
	sessions[sessionID][name] = val
}

func clearVarLocal(c echo.Context, name string) {
	sessionID, err := getSessionID(c)
	if err != nil {
		return
	}

	if _, exists := sessions[sessionID]; !exists {
		return
	}
	sessions[sessionID][name] = ""
}

func getVarLocal(c echo.Context, name string) (string, error) {
	sessionID, err := getSessionID(c)
	if err != nil {
		return "", err
	}
	if val, ok := sessions[sessionID][name]; ok {
		return val, nil
	}
	return "", fmt.Errorf("No var named %s available in session\n", name)
}

func clearSessionIDLocal(key string) {
	delete(sessions, key)
}

// ----------------------    Util ---------------------------------

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func GenToken() string {
	now := time.Now().Unix()
	h := md5.New()
	io.WriteString(h, strconv.FormatInt(now, 10))
	return fmt.Sprintf("%x", h.Sum(nil))
}
