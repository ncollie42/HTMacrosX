package main

import (
	_ "embed"
	"errors"
	db "myapp/DB"
	"myapp/auth"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed js/htmx.min.js
var htmxJS []byte

//go:embed css/daisy.min.css
var daisyCSS []byte

//go:embed js/html5-qrcode.min.js
var html5QrcodeJS []byte

//go:embed pwa/manifest.json
var manifestJSON []byte

//go:embed pwa/sw.js
var swJS []byte

//go:embed icons/icon.svg
var iconSVG []byte

//go:embed css/app.css
var appCSS []byte

const ctxUserID = "userID"
const defaultMealName = "Quick Add Meal"

func main() {
	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status} latency=${latency_human} Error=${error}\n",
	}))

	// Static assets
	e.GET("/favicon.ico", fav)
	e.GET("/htmx", htmx)
	e.GET("/daisy", daisy)
	e.GET("/app.css", appCSSHandler)
	e.GET("/html5qrcode", html5qrcode)
	e.GET("/manifest.json", manifest)
	e.GET("/sw.js", sw)
	e.GET("/icon", icon)

	// Feature routes
	registerOverviewRoutes(e)
	registerMealRoutes(e)
	registerTemplateRoutes(e)
	registerFoodRoutes(e)
	registerScanRoutes(e)
	registerAuthRoutes(e)

	e.Logger.Fatal(e.Start(":8080"))
}

// ---------- middleware -------------------

func validate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userID, err := auth.GetUserFromCookie(c)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/signin")
		}
		c.Set(ctxUserID, userID)
		return next(c)
	}
}

func handleDBErr(c echo.Context, err error) error {
	if errors.Is(err, db.ErrNotOwned) {
		return c.NoContent(http.StatusNotFound)
	}
	return c.NoContent(http.StatusInternalServerError)
}

// ---------- Static asset handlers -----------

func fav(c echo.Context) error {
	return c.NoContent(http.StatusNotFound)
}

func htmx(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/javascript")
	_, err := c.Response().Writer.Write(htmxJS)
	return err
}

func daisy(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/css")
	_, err := c.Response().Writer.Write(daisyCSS)
	return err
}

func appCSSHandler(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/css")
	_, err := c.Response().Writer.Write(appCSS)
	return err
}

func html5qrcode(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/javascript")
	_, err := c.Response().Writer.Write(html5QrcodeJS)
	return err
}

func manifest(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/manifest+json")
	c.Response().Header().Set("Cache-Control", "public, max-age=86400")
	_, err := c.Response().Writer.Write(manifestJSON)
	return err
}

func sw(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/javascript")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Service-Worker-Allowed", "/")
	_, err := c.Response().Writer.Write(swJS)
	return err
}

func icon(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "image/svg+xml")
	c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	_, err := c.Response().Writer.Write(iconSVG)
	return err
}
