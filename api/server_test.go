package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kjwardy/go-url/api/logging"
	"github.com/labstack/echo"
)

func TestRecoveredPanicEmitsSanitizedEvents(t *testing.T) {
	var output bytes.Buffer
	logger := logging.New(logging.Config{JSON: true, Output: &output})
	e := echo.New()
	e.Use(logger.Middleware())
	e.Use(recoverWithLogger(logger))
	e.GET("/panic", func(echo.Context) error {
		panic("private panic detail")
	})

	response := httptest.NewRecorder()
	e.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/panic", nil))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}

	var actions []string
	scanner := bufio.NewScanner(&output)
	for scanner.Scan() {
		var entry map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			t.Fatalf("invalid JSON log: %v", err)
		}
		actions = append(actions, entry["event.action"].(string))
		if entry["event.action"] == "application.panic" && entry["error.type"] != "panic_recovered" {
			t.Errorf("error.type = %v, want panic_recovered", entry["error.type"])
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("reading log output: %v", err)
	}
	if len(actions) != 2 || actions[0] != "application.panic" || actions[1] != "http.server.request" {
		t.Errorf("event actions = %v, want application.panic and http.server.request", actions)
	}
	if bytes.Contains(output.Bytes(), []byte("private panic detail")) {
		t.Fatal("panic details leaked into logs")
	}
}
