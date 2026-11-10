package handler

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

func TestSlackCommandLogsRejectedSignature(t *testing.T) {
	var output bytes.Buffer
	logger := logging.New(logging.Config{JSON: true, Output: &output})
	e := echo.New()
	e.Use(logger.Middleware())
	e.POST("/api/slack", (&Handler{Logger: logger}).SlackCommand)

	request := httptest.NewRequest(http.MethodPost, "/api/slack", bytes.NewBufferString("token=private"))
	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}

	var found bool
	scanner := bufio.NewScanner(&output)
	for scanner.Scan() {
		var entry map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			t.Fatalf("invalid JSON log: %v", err)
		}
		if entry["event.action"] == "slack.command" {
			found = true
			if entry["event.outcome"] != "failure" {
				t.Errorf("event.outcome = %v, want failure", entry["event.outcome"])
			}
			if entry["error.type"] != "invalid_signature" {
				t.Errorf("error.type = %v, want invalid_signature", entry["error.type"])
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("reading log output: %v", err)
	}
	if !found {
		t.Fatal("Slack command failure event was not logged")
	}
	if bytes.Contains(output.Bytes(), []byte("private")) {
		t.Fatal("Slack request body leaked into logs")
	}
}
