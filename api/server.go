package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	_ "github.com/joho/godotenv/autoload"
	"github.com/kjwardy/go-url/api/app"
	"github.com/kjwardy/go-url/api/config"
	"github.com/kjwardy/go-url/api/db"
	"github.com/kjwardy/go-url/api/logging"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	e := echo.New()
	config.Init()
	appConfig := config.GetConfig()

	if appConfig.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn: appConfig.SentryDSN,
		}); err != nil {
			fmt.Printf("Sentry initialization failed: %v\n", err)
		}
	}

	e.Pre(middleware.RemoveTrailingSlash())
	logger := logging.New(logging.Config{
		JSON:               appConfig.JSONLogs,
		ServiceVersion:     appConfig.ServiceVersion,
		ServiceEnvironment: appConfig.ServiceEnvironment,
	})
	e.Use(logger.Middleware())
	e.Use(middleware.Recover())

	// TODO after updating to echo v4
	// app.Use(sentryecho.New(sentryecho.Options{}))

	e.Debug = appConfig.Debug

	db.Init()
	if err := app.Init(e, logger); err != nil {
		e.Logger.Fatal(err)
	}
	// Start server
	go func() {
		if err := e.Start(fmt.Sprintf(":%d", appConfig.Port)); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal(err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Error(err)
	}
}
