package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
	e.HideBanner = true
	e.HidePort = true
	e.Logger.SetOutput(io.Discard)
	jsonLogs, _ := strconv.ParseBool(strings.TrimSpace(os.Getenv("JSON_LOGS")))
	logger := logging.New(logging.Config{
		JSON:               jsonLogs,
		ServiceVersion:     os.Getenv("SERVICE_VERSION"),
		ServiceEnvironment: os.Getenv("SERVICE_ENVIRONMENT"),
	})
	if err := config.Init(); err != nil {
		logger.Event("ERROR", "Application startup failed", "application.startup", "failure", map[string]interface{}{
			"application.component": "configuration",
			"error.type":            "configuration_error",
		})
		os.Exit(1)
	}
	appConfig := config.GetConfig()

	if appConfig.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn: appConfig.SentryDSN,
		}); err != nil {
			logger.Event("ERROR", "Sentry initialization failed", "application.startup", "failure", map[string]interface{}{
				"application.component": "sentry",
				"error.type":            "sentry_initialization_error",
			})
		}
	}

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(logger.Middleware())
	e.Use(recoverWithLogger(logger))

	// TODO after updating to echo v4
	// app.Use(sentryecho.New(sentryecho.Options{}))

	e.Debug = appConfig.Debug

	if err := db.Init(logger); err != nil {
		logger.Event("ERROR", "Application startup failed", "application.startup", "failure", map[string]interface{}{
			"application.component": "database",
			"error.type":            "database_initialization_error",
		})
		os.Exit(1)
	}
	if err := app.Init(e, logger); err != nil {
		logger.Event("ERROR", "Application startup failed", "application.startup", "failure", map[string]interface{}{
			"application.component": "routes",
			"error.type":            "application_initialization_error",
		})
		os.Exit(1)
	}

	listener, err := net.Listen("tcp", net.JoinHostPort("", strconv.Itoa(appConfig.Port)))
	if err != nil {
		logger.Event("ERROR", "Application startup failed", "application.startup", "failure", map[string]interface{}{
			"application.component": "http_server",
			"server.port":           appConfig.Port,
			"error.type":            "server_listen_error",
		})
		os.Exit(1)
	}
	logger.Event("INFO", "Application started", "application.startup", "success", map[string]interface{}{
		"server.port": appConfig.Port,
	})

	serverErrors := make(chan error, 1)
	go func() {
		if err := e.Server.Serve(listener); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	select {
	case <-shutdown:
	case <-serverErrors:
		logger.Event("ERROR", "HTTP server stopped unexpectedly", "application.shutdown", "failure", map[string]interface{}{
			"error.type": "server_error",
		})
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logger.Event("INFO", "Application shutdown started", "application.shutdown", "success", nil)
	if err := e.Shutdown(ctx); err != nil {
		logger.Event("ERROR", "Application shutdown failed", "application.shutdown", "failure", map[string]interface{}{
			"error.type": "graceful_shutdown_error",
		})
		os.Exit(1)
	}
	logger.Event("INFO", "Application stopped", "application.shutdown", "success", nil)
}

func recoverWithLogger(logger *logging.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Event("ERROR", "Request panic recovered", "application.panic", "failure", map[string]interface{}{
						"http.request.id": logging.RequestID(c),
						"error.type":      "panic_recovered",
					})
					err = echo.NewHTTPError(http.StatusInternalServerError)
				}
			}()
			return next(c)
		}
	}
}
