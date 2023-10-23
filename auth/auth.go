package auth

import (
	"crypto/rand"
	"os"

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

// NOTE: using if/else vs func pointers for now.
func InitSession(isProd bool) {
	prod = isProd
	if prod {
		initRedis()
	}
	// NOTE: nothing needed to init Local
}

func setCookie(c echo.Context, userID int) {
	if prod {
		setSessionCookieRedis(c, userID)
	} else {
		setSessionCookieLocal(c, userID)
	}
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
	if prod {
		return getUserFromCookieSessionRedis(c)
	} else {
		return getUserFromCookieSessionLocal(c)
	}
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

	// err = rClient.Set(ctx, "foo", "userID -> ??", 0).Err()
	// if err != nil {
	// 	panic(err)
	// }
	// val := rClient.Get(ctx, "foo").Val()
	// fmt.Println("Redis:", val)
}

func getUserFromCookieSessionRedis(c echo.Context) (int, error) {
	cookie, err := c.Cookie(cookieName)
	// No cookie
	if err != nil {
		return 0, err
	}
	// val, err := rClient.Get(ctx, cookie.Value).Result()
	val, err := rClient.Get(ctx, cookie.Value).Int()
	// Invalid cookie
	if err != nil {
		return 0, fmt.Errorf("Invalid Session ID\n")
	}
	return val, nil
}

func setSessionCookieRedis(c echo.Context, userID int) {
	sessionID, err := generateSessionID()
	if err != nil {
		panic(err)
	}
	// TODO: set expire time on Redis
	err = rClient.Set(ctx, sessionID, userID, 0).Err()
	if err != nil {
		panic(err)
	}

	expiration := time.Now().Add(time.Hour * 24)
	cookie := http.Cookie{
		Name:    cookieName,
		Value:   sessionID,
		Expires: expiration,
	}
	c.SetCookie(&cookie)
}

func clearSessionIDRedis(key string) {
	_, err := rClient.Del(ctx, key).Result()
	if err != nil {
		panic(err)
	}
}

// ---------------------- Sessions --------------------------------
var sessions = make(map[string]int)

const cookieName = "sessionID"

func setSessionCookieLocal(c echo.Context, userID int) {
	sessionID, err := generateSessionID()
	if err != nil {
		panic(err)
	}
	sessions[sessionID] = userID

	expiration := time.Now().Add(time.Hour * 24)
	cookie := http.Cookie{
		Name:    cookieName,
		Value:   sessionID,
		Expires: expiration,
	}
	c.SetCookie(&cookie)
}

func getUserFromCookieSessionLocal(c echo.Context) (int, error) {
	cookie, err := c.Cookie(cookieName)
	// No cookie
	if err != nil {
		return 0, err
	}
	if val, ok := sessions[cookie.Value]; ok {
		return val, nil
	}
	// Invalid cookie
	return 0, fmt.Errorf("Invalid Session ID\n")
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

// ----------------------    Util --------------------------------

// generateSessionID creates a secure random session ID
func generateSessionID() (string, error) {
	b := make([]byte, 32) // Using 32 bytes, but you can adjust as per your needs
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
