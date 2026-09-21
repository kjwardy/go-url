package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
)

func TestJSONRequestLog(t *testing.T) {
	var output bytes.Buffer
	logger := New(Config{JSON: true, Output: &output, ServiceVersion: "1.2.3", ServiceEnvironment: "test"})
	e := echo.New()
	e.Use(logger.Middleware())
	e.GET("/api/url/:key", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/url/handbook?token=secret", nil)
	req.Header.Set(echo.HeaderXRequestID, "existing-request-id")
	req.Header.Set(echo.HeaderAuthorization, "Bearer secret")
	req.Header.Set(echo.HeaderCookie, "session=secret")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	entry := decodeEntry(t, output.Bytes())
	assertEqual(t, entry["log.level"], "INFO")
	assertEqual(t, entry["message"], "HTTP request completed")
	assertEqual(t, entry["event.action"], "http.server.request")
	assertEqual(t, entry["event.outcome"], "success")
	assertEqual(t, entry["ecs.version"], "9.5.0")
	assertEqual(t, entry["service.name"], "go-url-api")
	assertEqual(t, entry["service.version"], "1.2.3")
	assertEqual(t, entry["service.environment"], "test")
	assertEqual(t, entry["http.request.id"], "existing-request-id")
	assertEqual(t, entry["http.request.method"], http.MethodGet)
	assertEqual(t, entry["http.route"], "/api/url/:key")
	assertEqual(t, entry["http.response.status_code"], float64(http.StatusNoContent))
	if _, ok := entry["event.duration"].(float64); !ok {
		t.Fatalf("event.duration is not numeric: %#v", entry["event.duration"])
	}
	if _, ok := entry["@timestamp"].(string); !ok {
		t.Fatalf("@timestamp is not a string: %#v", entry["@timestamp"])
	}
	logged := output.String()
	for _, sensitive := range []string{"token", "secret", "Authorization", "Cookie", "handbook"} {
		if strings.Contains(logged, sensitive) {
			t.Errorf("log contains excluded value %q: %s", sensitive, logged)
		}
	}
}

func TestRequestIDValidation(t *testing.T) {
	tables := []struct {
		name     string
		incoming string
		retained bool
	}{
		{name: "valid", incoming: "request.ID_123-abc", retained: true},
		{name: "invalid characters", incoming: "request id", retained: false},
		{name: "too long", incoming: strings.Repeat("a", 129), retained: false},
		{name: "missing", retained: false},
	}

	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			var output bytes.Buffer
			logger := New(Config{JSON: true, Output: &output})
			e := echo.New()
			e.Use(logger.Middleware())
			e.GET("/api/test", func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			})
			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			if table.incoming != "" {
				req.Header.Set(echo.HeaderXRequestID, table.incoming)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			requestID := rec.Header().Get(echo.HeaderXRequestID)
			if !validRequestID(requestID) {
				t.Fatalf("response request ID is invalid: %q", requestID)
			}
			if table.retained && requestID != table.incoming {
				t.Errorf("request ID = %q, want %q", requestID, table.incoming)
			}
			if !table.retained && requestID == table.incoming {
				t.Errorf("invalid request ID was retained: %q", requestID)
			}
			entry := decodeEntry(t, output.Bytes())
			assertEqual(t, entry["http.request.id"], requestID)
		})
	}
}

func TestSkippedRequestsAreNotLogged(t *testing.T) {
	for _, path := range []string{"/health", "/go", "/go/config.js", "/go/assets/app.js"} {
		t.Run(path, func(t *testing.T) {
			var output bytes.Buffer
			logger := New(Config{JSON: true, Output: &output})
			e := echo.New()
			e.Use(logger.Middleware())
			e.GET("/*", func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			})
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
			if output.Len() != 0 {
				t.Errorf("unexpected log output: %s", output.String())
			}
			if rec.Header().Get(echo.HeaderXRequestID) == "" {
				t.Error("request ID header was not set")
			}
		})
	}
}

func TestQueryLog(t *testing.T) {
	var output bytes.Buffer
	logger := New(Config{JSON: true, Output: &output})
	logger.Query("request-123", "handbook", false)

	entry := decodeEntry(t, output.Bytes())
	assertEqual(t, entry["log.level"], "INFO")
	assertEqual(t, entry["message"], "URL query unresolved")
	assertEqual(t, entry["event.action"], "go_url.query")
	assertEqual(t, entry["event.outcome"], "failure")
	assertEqual(t, entry["ecs.version"], "9.5.0")
	assertEqual(t, entry["http.request.id"], "request-123")
	assertEqual(t, entry["go_url.query.key"], "handbook")
	assertEqual(t, entry["go_url.query.successful"], false)
}

func TestInternalErrorIsSanitized(t *testing.T) {
	var output bytes.Buffer
	logger := New(Config{JSON: true, Output: &output})
	e := echo.New()
	e.Use(logger.Middleware())
	e.GET("/api/fail", func(c echo.Context) error {
		return errors.New("database password is secret")
	})

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/fail", nil))
	entry := decodeEntry(t, output.Bytes())
	assertEqual(t, entry["log.level"], "ERROR")
	assertEqual(t, entry["event.outcome"], "failure")
	assertEqual(t, entry["error.type"], "internal_server_error")
	assertEqual(t, entry["error.message"], "Internal server error")
	if strings.Contains(output.String(), "database password") || strings.Contains(output.String(), "secret") {
		t.Errorf("internal error leaked into log: %s", output.String())
	}
}

func TestHumanReadableOutput(t *testing.T) {
	var output bytes.Buffer
	logger := New(Config{Output: &output})
	logger.Query("request-123", "handbook", true)
	if got := output.String(); !strings.Contains(got, "INFO URL query resolved") || !strings.Contains(got, "request_id=request-123") || !strings.Contains(got, "key=handbook") {
		t.Errorf("unexpected human-readable output: %s", got)
	}
}

func decodeEntry(t *testing.T, data []byte) map[string]interface{} {
	t.Helper()
	var entry map[string]interface{}
	if err := json.Unmarshal(bytes.TrimSpace(data), &entry); err != nil {
		t.Fatalf("invalid JSON log %q: %v", data, err)
	}
	return entry
}

func assertEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	if got != want {
		t.Errorf("got %#v, want %#v", got, want)
	}
}
