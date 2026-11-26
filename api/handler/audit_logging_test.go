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

func TestURLChangeAuditIsStructuredAndExcludesDestination(t *testing.T) {
	var output bytes.Buffer
	logger := logging.New(logging.Config{Output: &output})
	h := &Handler{Logger: logger}
	e := echo.New()
	e.Use(logger.Middleware())
	e.POST("/:key", func(c echo.Context) error {
		h.auditURLChange(c, logging.AuditURLCreate, logging.AuditSuccess, "handbook")
		return c.NoContent(http.StatusOK)
	})

	request := httptest.NewRequest(http.MethodPost, "/handbook", nil)
	request.RemoteAddr = "198.51.100.10:4321"
	request.Header.Set(echo.HeaderXRequestID, "create-request")
	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)

	var auditEvent map[string]interface{}
	scanner := bufio.NewScanner(&output)
	for scanner.Scan() {
		var entry map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil && entry["event.dataset"] == "go_url.audit" {
			auditEvent = entry
			break
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if auditEvent == nil {
		t.Fatalf("no structured audit event found: %s", output.String())
	}
	if auditEvent["event.action"] != "audit.url.create" || auditEvent["event.outcome"] != "success" {
		t.Errorf("audit event = %#v", auditEvent)
	}
	if auditEvent["http.request.id"] != "create-request" || auditEvent["source.ip"] != "198.51.100.10" {
		t.Errorf("request attribution = %#v", auditEvent)
	}
	if auditEvent["go_url.url.key"] != "handbook" {
		t.Errorf("logged URL key = %#v, want handbook", auditEvent["go_url.url.key"])
	}
	if _, ok := auditEvent["go_url.url.url"]; ok {
		t.Errorf("destination field unexpectedly logged: %#v", auditEvent)
	}
}

func TestInvalidURLMutationAttemptsAreAuditedWithoutLoggingInvalidKey(t *testing.T) {
	var output bytes.Buffer
	logger := logging.New(logging.Config{Output: &output})
	h := &Handler{Logger: logger}
	e := echo.New()
	e.Use(logger.Middleware())
	e.POST("/:key", h.CreateUrl)
	e.PUT("/:key", h.UpdateUrl)
	e.DELETE("/:key", h.DeleteUrl)

	tests := []struct {
		method    string
		action    string
		requestID string
	}{
		{method: http.MethodPost, action: "audit.url.create", requestID: "invalid-create"},
		{method: http.MethodPut, action: "audit.url.update", requestID: "invalid-update"},
		{method: http.MethodDelete, action: "audit.url.delete", requestID: "invalid-delete"},
	}
	for _, test := range tests {
		request := httptest.NewRequest(test.method, "/invalid$key", nil)
		request.Header.Set(echo.HeaderXRequestID, test.requestID)
		response := httptest.NewRecorder()
		e.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Errorf("%s response status = %d, want %d", test.method, response.Code, http.StatusBadRequest)
		}
	}

	var auditEntries []map[string]interface{}
	scanner := bufio.NewScanner(&output)
	for scanner.Scan() {
		var entry map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil && entry["event.dataset"] == "go_url.audit" {
			auditEntries = append(auditEntries, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(auditEntries) != len(tests) {
		t.Fatalf("got %d audit events, want %d: %s", len(auditEntries), len(tests), output.String())
	}
	for i, entry := range auditEntries {
		if entry["event.action"] != tests[i].action || entry["event.outcome"] != "failure" {
			t.Errorf("audit event %d = %#v", i, entry)
		}
		if _, ok := entry["go_url.url.key"]; ok {
			t.Errorf("invalid key was logged: %#v", entry)
		}
	}
}
