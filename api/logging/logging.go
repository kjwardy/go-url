package logging

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/labstack/echo"
)

const (
	ecsVersion          = "9.5.0" // ECS schema version emitted in all events
	requestIDContextKey = "request_id"
)

var (
	requestIDPattern  = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`) // Valid request ID characters and length
	requestIDFallback uint64                                          // Counter for request ID generation when crypto rand fails
)

type Config struct {
	JSON               bool
	Output             io.Writer
	ServiceVersion     string
	ServiceEnvironment string
}

type Logger struct {
	json               bool
	output             io.Writer
	serviceVersion     string
	serviceEnvironment string
	mu                 sync.Mutex
}

func New(config Config) *Logger {
	if config.Output == nil {
		config.Output = os.Stdout
	}
	return &Logger{
		json:               config.JSON,
		output:             config.Output,
		serviceVersion:     config.ServiceVersion,
		serviceEnvironment: config.ServiceEnvironment,
	}
}

func (l *Logger) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Validate or replace incoming request ID
			requestID := c.Request().Header.Get(echo.HeaderXRequestID)
			if !validRequestID(requestID) {
				requestID = newRequestID()
			}
			c.Set(requestIDContextKey, requestID)
			c.Response().Header().Set(echo.HeaderXRequestID, requestID)

			// Measure request duration and log on completion
			start := time.Now()
			err := next(c)
			if err != nil {
				c.Error(err)
			}
			if !skipRequest(c.Request().URL.Path) {
				l.request(c, requestID, time.Since(start), err)
			}
			return nil
		}
	}
}

func RequestID(c echo.Context) string {
	requestID, _ := c.Get(requestIDContextKey).(string)
	return requestID
}

func (l *Logger) AuthenticationFailure(requestID, provider, errorType string) {
	fields := l.base("WARN", "Authentication failed", "authentication", time.Now().UTC())
	fields["event.outcome"] = "failure"
	fields["http.request.id"] = requestID
	fields["auth.provider"] = provider
	fields["error.type"] = errorType
	l.write(fields)
}

func (l *Logger) Query(requestID, key string, successful bool) {
	// Set message and outcome based on resolution result
	message := "URL query unresolved"
	if successful {
		message = "URL query resolved"
	}
	fields := l.base("INFO", message, "go_url.query", time.Now().UTC())
	fields["event.outcome"] = outcome(successful)
	fields["http.request.id"] = requestID
	fields["go_url.query.key"] = key
	fields["go_url.query.successful"] = successful
	l.write(fields)
}

func (l *Logger) request(c echo.Context, requestID string, duration time.Duration, err error) {
	// Determine log level and outcome from HTTP status
	status := c.Response().Status
	level := "INFO"
	if status >= http.StatusInternalServerError {
		level = "ERROR"
	} else if status >= http.StatusBadRequest {
		level = "WARN"
	}
	fields := l.base(level, "HTTP request completed", "http.server.request", time.Now().UTC())
	fields["event.outcome"] = outcome(status < http.StatusBadRequest)
	fields["event.duration"] = duration.Nanoseconds()
	fields["http.request.id"] = requestID
	fields["http.request.method"] = c.Request().Method
	fields["http.route"] = route(c)
	fields["http.response.status_code"] = status
	// Add sanitized error details for failures
	if err != nil || status >= http.StatusInternalServerError {
		errorType, message := safeError(err, status)
		fields["error.type"] = errorType
		fields["error.message"] = message
	}
	l.write(fields)
}

func (l *Logger) base(level, message, action string, timestamp time.Time) map[string]interface{} {
	// Build common ECS fields present in all events
	fields := map[string]interface{}{
		"@timestamp":   timestamp.Format(time.RFC3339Nano),
		"ecs.version":  ecsVersion,
		"log.level":    level,
		"message":      message,
		"event.action": action,
		"service.name": "go-url-api",
	}
	// Add optional service metadata
	if l.serviceVersion != "" {
		fields["service.version"] = l.serviceVersion
	}
	if l.serviceEnvironment != "" {
		fields["service.environment"] = l.serviceEnvironment
	}
	return fields
}

func (l *Logger) write(fields map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	// Emit JSON when configured, otherwise use human-readable format
	if l.json {
		_ = json.NewEncoder(l.output).Encode(fields)
		return
	}
	_, _ = fmt.Fprintf(l.output, "%s %s %s", fields["@timestamp"], fields["log.level"], fields["message"])
	// Append relevant fields in key=value format
	for _, key := range []string{"event.action", "event.outcome", "auth.provider", "http.request.id", "http.request.method", "http.route", "http.response.status_code", "event.duration", "go_url.query.key", "go_url.query.successful", "error.type", "error.message"} {
		if value, ok := fields[key]; ok {
			_, _ = fmt.Fprintf(l.output, " %s=%v", humanKey(key), value)
		}
	}
	_, _ = fmt.Fprintln(l.output)
}

func humanKey(key string) string {
	switch key {
	case "http.request.id":
		return "request_id"
	case "http.request.method":
		return "method"
	case "http.response.status_code":
		return "status"
	case "event.duration":
		return "duration_ns"
	case "go_url.query.key":
		return "key"
	case "go_url.query.successful":
		return "successful"
	case "event.action":
		return "event"
	case "event.outcome":
		return "outcome"
	case "http.route":
		return "route"
	case "error.type":
		return "error_type"
	case "error.message":
		return "error_message"
	default:
		return key
	}
}

func outcome(successful bool) string {
	if successful {
		return "success"
	}
	return "failure"
}

func validRequestID(requestID string) bool {
	return requestIDPattern.MatchString(requestID)
}

func newRequestID() string {
	// Prefer cryptographically random bytes; fall back to a deterministic hash if unavailable
	value := make([]byte, 16)
	if _, err := rand.Read(value); err == nil {
		return hex.EncodeToString(value)
	}
	fallback := fmt.Sprintf("%d-%d", time.Now().UnixNano(), atomic.AddUint64(&requestIDFallback, 1))
	hash := sha256.Sum256([]byte(fallback))
	return hex.EncodeToString(hash[:16])
}

func skipRequest(path string) bool {
	return path == "/health" || path == "/go" || strings.HasPrefix(path, "/go/")
}

func route(c echo.Context) string {
	if path := c.Path(); path != "" {
		return path
	}
	return "unmatched"
}

func safeError(err error, status int) (string, string) {
	// Return a generic message for 5xx errors; use public error message for 4xx responses
	if status >= http.StatusInternalServerError {
		return "internal_server_error", "Internal server error"
	}
	if httpError, ok := err.(*echo.HTTPError); ok {
		if message, ok := httpError.Message.(string); ok {
			return "http_error", message
		}
	}
	return "http_error", http.StatusText(status)
}
