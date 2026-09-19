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
	logFormat := "${time_rfc3339}\t|${status}|\t${latency_human}| ${remote_ip} | ${method} | ${path} ${error}\n"
	if appConfig.JSONLogs {
		logFormat = ""
	}
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: logFormat,
	}))
	e.Use(middleware.Recover())

	// TODO after updating to echo v4
	// app.Use(sentryecho.New(sentryecho.Options{}))

	e.Debug = appConfig.Debug

	db.Init()
	app.Init(e)
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
