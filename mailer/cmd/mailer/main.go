package main

import (
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// for startup probe
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, struct{ Status string }{Status: "OK"})
	})

	// for readiness probe
	e.GET("/ready", func(c echo.Context) error {
		return c.JSON(http.StatusOK, struct{ Status string }{Status: "OK"})
	})

	e.GET("/hello", func(c echo.Context) error {
		return c.HTML(http.StatusOK, "Hello, Docker!")
	})

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8082"
	}

	time.Sleep(5 * time.Second)

	e.Logger.Fatal(e.Start(":" + httpPort))
}
