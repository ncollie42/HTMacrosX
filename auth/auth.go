package auth

import (
	"crypto/md5"
	"crypto/rand"
	"io"
	"os"
	"strconv"

	"encoding/base64"
	"fmt"
	db "myapp/DB"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"context"

	"github.com/redis/go-redis/v9"
)

func Signin(c echo.Context, login string, pass string) error {
	// TODO: sanitize input
	userID, err := db.ValidateUser(login, pass)
	if err != nil {
		fmt.Println("Err:", err)
		return fmt.Errorf("Invalid username or password")
	}

	setCookie(c, userID)
	return nil
}

func Signup(c echo.Context, login string, pass string) error {
	// TODO: sanitize input
	userID, err := db.CreateUser(login, pass)
	if err != nil {
		fmt.Println("Err:", err)
		return fmt.Errorf("Invalid username or password")
	}

	setCookie(c, userID)
	return nil
}

// ---------------------- Generic --------------------------------
var prod = false

const userID_tag = "userID"
const token_tag = "token"

// NOTE: using if/else vs func pointers for now.
func InitSession(isProd bool) {
	prod = isProd
	if prod {
		initRedis()
	}
	// NOTE: nothing needed to init Local
}

func setCookie(c echo.Context, userID string) {
	sessionID, err := generateSessionID()
	if err != nil {
		panic(err)
	}

	if prod {
		setVarRedis(sessionID, userID_tag, userID)
	} else {
		setVarLocal(sessionID, userID_tag, userID)
	}

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
	// No cookie
	if err != nil {
		return
	}

	if prod {
		clearSessionIDRedis(cookie.Value)
	} else {
		clearSessionIDLocal(cookie.Value)
	}

	newCookie := http.Cookie{
		Name:   cookieName,
		Value:  "",
		MaxAge: -1,
	}
	c.SetCookie(&newCookie)
}

func GetUserFromCookie(c echo.Context) (int, error) {
	// TODO: Return an object with more userData; ID | UserName | LastLogin ext.

	// TODO: update from int to string or full object
	var userID_str string
	var err error

	if prod {
		userID_str, err = getVarRedis(c, userID_tag)
	} else {
		userID_str, err = getVarLocal(c, userID_tag)
	}
	userID_int, err := strconv.ParseInt(userID_str, 10, 64)
	return int(userID_int), err
}

func SetDupToken(c echo.Context, token string) string {
	// To prevent duplicate submitions from double tapping template; set in cache and in button param
	// TODO: make this a seperate API?
	// TODO: return error
	sessionID, _ := getSessionID(c)

	if prod {
		setVarRedis(sessionID, "token", token)
	} else {
		setVarLocal(sessionID, "token", token)
	}
	return token
}

func ValidateDupToken(c echo.Context, token string) bool {
	if prod {
		sessionToken, _ := getVarRedis(c, "token")
		return token == sessionToken
	} else {
		sessionToken, _ := getVarLocal(c, "token")
		return token == sessionToken
	}
}

func ClearDupToken(c echo.Context) {
	if prod {
		clearVarRedis(c, "token")
	} else {
		clearVarLocal(c, "token")
	}
}

func getSessionID(c echo.Context) (string, error) {
	cookie, err := c.Cookie(cookieName)

	if err != nil {
		return "", fmt.Errorf("No valid cookie")
	}
	return cookie.Value, nil
}

// ----------------------   Redis   --------------------------------
var ctx = context.Background()
var rClient *redis.Client

// Redis -> SetSession, ClearSession, GetSession
func initRedis() {
	redisURL := os.Getenv("REDIS_URL")
	fmt.Println("Connecting to Redis: ", redisURL)

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(err)
	}
	rClient = redis.NewClient(opt)
}

func setVarRedis(sessionID string, name string, val string) {
	err := rClient.HSet(ctx, sessionID, name, val).Err()
	if err != nil {
		panic(err)
	}
}

func clearVarRedis(c echo.Context, name string) {
	sessionID, err := getSessionID(c)
	// No cookie
	if err != nil {
		panic(err)
	}

	_, err = rClient.HDel(ctx, sessionID, name).Result()
	if err != nil {
		panic(err)
	}
}

func getVarRedis(c echo.Context, name string) (string, error) {
	sessionID, err := getSessionID(c)
	// No cookie
	if err != nil {
		return "", err
	}
	str, err := rClient.HGet(ctx, sessionID, name).Result()
	if err != nil {
		return "", fmt.Errorf("No var named %s available in session\n", name)
	}
	return str, nil
}

func clearSessionIDRedis(key string) {
	_, err := rClient.Del(ctx, key).Result()
	if err != nil {
		panic(err)
	}
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
	// No cookie
	if err != nil {
		return
	}

	if _, exists := sessions[sessionID]; !exists {
		return
	}
	// delete(sessions[sessionID], name)
	sessions[sessionID][name] = ""
}

func getVarLocal(c echo.Context, name string) (string, error) {
	sessionID, err := getSessionID(c)
	// No cookie
	if err != nil {
		return "", err
	}
	if val, ok := sessions[sessionID][name]; ok {
		return val, nil
	}
	// Invalid cookie
	return "", fmt.Errorf("No var named %s available in session\n", name)
}

func clearSessionIDLocal(key string) {
	delete(sessions, key)
}

// ----------------------    JWT   --------------------------------
func getUserFromCookieJWT() (int, error) {
	// Not yet implemented
	return 0, nil
}

func setJWTCookie(w http.ResponseWriter, userID int) {
	// Not yet implemented
}

// ----------------------    Util ---------------------------------

// generateSessionID creates a secure random session ID
func generateSessionID() (string, error) {
	b := make([]byte, 32) // Using 32 bytes, but you can adjust as per your needs
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
