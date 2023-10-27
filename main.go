package main

import (
	"context"
	"log"
	"os"

	"github.com/TiZir/service_with_enrichment/background"
	"github.com/TiZir/service_with_enrichment/handler"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	repo, err := background.NewUserRepository(ctx)
	if err != nil {
		log.Fatalf("Error creating repository: %v", err)
	}

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Routes
	e.GET("/", handler.HomePageHandler)
	e.GET("/users", func(c echo.Context) error {

		return handler.GetUsersHandler(c, repo)
	})
	e.POST("/users", func(c echo.Context) error {
		return handler.AddUsersHandler(c, repo)
	})
	e.GET("/users/:id", func(c echo.Context) error {
		return handler.GetUserByIdHandler(c, repo)
	})
	e.PUT("/users/:id", func(c echo.Context) error {
		return handler.UpdateUsersHandler(c, repo)
	})
	e.DELETE("/users/:id", func(c echo.Context) error {
		return handler.DeleteUsersHandler(c, repo)
	})

	// Start server
	e.Logger.Fatal(e.Start(":" + os.Getenv("HTTP_PORT")))
	return nil
}
