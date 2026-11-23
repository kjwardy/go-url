package handler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kjwardy/go-url/api/logging"
	"github.com/labstack/echo"
)

func frontendEventServer(output *bytes.Buffer, limiter *FrontendEventLimiter) *echo.Echo {
	e := echo.New()
	logger := logging.New(logging.Config{JSON: true, Output: output})
	h := &Handler{Logger: logger, FrontendEventLimiter: limiter}
	e.Use(logger.Middleware())
	e.POST("/api/frontend-events", h.ReportFrontendEvent)
	return e
}

func frontendEventRequest(e *echo.Echo, remoteAddr, requestID, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/frontend-events", strings.NewReader(body))
	req.RemoteAddr = remoteAddr
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-Forwarded-For", "203.0.113.99")
	if requestID != "" {
		req.Header.Set(echo.HeaderXRequestID, requestID)
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func logEntries(t *testing.T, output string) []map[string]interface{} {
	t.Helper()
	var entries []map[string]interface{}
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		var entry map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			t.Fatalf("invalid JSON log: %v", err)
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return entries
}

func TestReportFrontendEventLogsAllowlistedError(t *testing.T) {
	var output bytes.Buffer
	e := frontendEventServer(&output, NewFrontendEventLimiter())
	response := frontendEventRequest(e, "192.0.2.15:8080", "client-request-1", `{"version":1,"type":"error","component":"search","code":"api_request_failed"}`)

	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusAccepted)
	}
	entries := logEntries(t, output.String())
	var event map[string]interface{}
	for _, entry := range entries {
		if entry["event.action"] == "frontend.error" {
			event = entry
			break
		}
	}
	if event == nil {
		t.Fatalf("frontend event not logged: %s", output.String())
	}
	for key, want := range map[string]interface{}{
		"log.level":          "ERROR",
		"event.outcome":      "failure",
		"event.action":       "frontend.error",
		"event.type":         []interface{}{"error"},
		"frontend.component": "search",
		"error.type":         "api_request_failed",
		"http.request.id":    "client-request-1",
		"source.ip":          "192.0.2.15",
	} {
		if fmt.Sprint(event[key]) != fmt.Sprint(want) {
			t.Errorf("%s = %#v, want %#v", key, event[key], want)
		}
	}
	if _, ok := event["@timestamp"].(string); !ok {
		t.Errorf("@timestamp = %#v, want a server timestamp", event["@timestamp"])
	}
	if event["message"] != "Frontend error reported" {
		t.Errorf("message = %#v", event["message"])
	}
}

func TestReportFrontendOperationLogsServerSelectedFields(t *testing.T) {
	var output bytes.Buffer
	e := frontendEventServer(&output, NewFrontendEventLimiter())
	response := frontendEventRequest(e, "[2001:db8::1]:8080", "operation-request", `{"version":1,"type":"operation","operation":"url.create","outcome":"success"}`)

	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusAccepted)
	}
	entries := logEntries(t, output.String())
	for _, entry := range entries {
		if entry["event.action"] == "frontend.url.create" {
			if entry["event.outcome"] != "success" || entry["log.level"] != "INFO" {
				t.Errorf("unexpected operation fields: %#v", entry)
			}
			if entry["frontend.component"] != "edit_url" || entry["source.ip"] != "2001:db8::1" {
				t.Errorf("unexpected server-enriched fields: %#v", entry)
			}
			return
		}
	}
	t.Fatalf("frontend operation not logged: %s", output.String())
}

func TestReportFrontendEventRejectsInvalidAndSensitivePayloads(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"unknown event", `{"version":1,"type":"arbitrary","component":"search","code":"api_request_failed"}`},
		{"unsupported version", `{"version":2,"type":"error","component":"search","code":"api_request_failed"}`},
		{"unknown component", `{"version":1,"type":"error","component":"unknown","code":"api_request_failed"}`},
		{"unknown error code", `{"version":1,"type":"error","component":"search","code":"raw error"}`},
		{"unknown field", `{"version":1,"type":"error","component":"search","code":"api_request_failed","message":"private-value"}`},
		{"client severity", `{"version":1,"type":"error","component":"search","code":"api_request_failed","severity":"DEBUG"}`},
		{"client timestamp", `{"version":1,"type":"error","component":"search","code":"api_request_failed","timestamp":"2026-01-01T00:00:00Z"}`},
		{"client identity", `{"version":1,"type":"error","component":"search","code":"api_request_failed","identity":"someone"}`},
		{"client ip", `{"version":1,"type":"error","component":"search","code":"api_request_failed","source_ip":"203.0.113.1"}`},
		{"arbitrary metadata", `{"version":1,"type":"error","component":"search","code":"api_request_failed","metadata":{"token":"secret"}}`},
		{"client cookie", `{"version":1,"type":"error","component":"search","code":"api_request_failed","cookie":"session-secret"}`},
		{"client token", `{"version":1,"type":"error","component":"search","code":"api_request_failed","token":"access-secret"}`},
		{"authorization header", `{"version":1,"type":"error","component":"search","code":"api_request_failed","authorization":"Bearer secret"}`},
		{"destination url", `{"version":1,"type":"error","component":"search","code":"api_request_failed","url":"https://private.example/path"}`},
		{"sensitive form values", `{"version":1,"type":"error","component":"edit_url","code":"api_request_failed","form_values":{"url":"https://private.example"}}`},
		{"field not allowed for event type", `{"version":1,"type":"error","component":"search","code":"api_request_failed","outcome":""}`},
		{"invalid outcome", `{"version":1,"type":"operation","operation":"url.create","outcome":"ERROR"}`},
		{"trailing data", `{"version":1,"type":"error","component":"search","code":"api_request_failed"} {}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			e := frontendEventServer(&output, NewFrontendEventLimiter())
			response := frontendEventRequest(e, "192.0.2.15:8080", "", test.body)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
			}
			if strings.Contains(output.String(), "private-value") || strings.Contains(output.String(), "secret") || strings.Contains(output.String(), "private.example") {
				t.Errorf("sensitive payload reached logs: %s", output.String())
			}
		})
	}
}

func TestReportFrontendEventEnforcesPayloadSizeAndContentType(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		contentType string
		wantStatus  int
	}{
		{"oversized", `{"version":1,"type":"error","component":"search","code":"api_request_failed","pad":"` + strings.Repeat("x", maxFrontendEventSize) + `"}`, echo.MIMEApplicationJSON, http.StatusRequestEntityTooLarge},
		{"wrong content type", `{"version":1,"type":"error","component":"search","code":"api_request_failed"}`, "text/plain", http.StatusUnsupportedMediaType},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			e := echo.New()
			logger := logging.New(logging.Config{JSON: true, Output: &output})
			h := &Handler{Logger: logger, FrontendEventLimiter: NewFrontendEventLimiter()}
			e.Use(logger.Middleware())
			e.POST("/api/frontend-events", h.ReportFrontendEvent)
			req := httptest.NewRequest(http.MethodPost, "/api/frontend-events", strings.NewReader(test.body))
			req.RemoteAddr = "192.0.2.15:8080"
			req.Header.Set(echo.HeaderContentType, test.contentType)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			if rec.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, test.wantStatus)
			}
		})
	}
}

func TestFrontendEventRateLimitIsPerIPAndExpires(t *testing.T) {
	limiter := NewFrontendEventLimiter()
	now := time.Now()
	limiter.now = func() time.Time { return now }
	var output bytes.Buffer
	e := frontendEventServer(&output, limiter)
	body := `{"version":1,"type":"operation","operation":"url.delete","outcome":"success"}`

	for i := 0; i < frontendEventLimit; i++ {
		response := frontendEventRequest(e, "192.0.2.15:8080", "", body)
		if response.Code != http.StatusAccepted {
			t.Fatalf("request %d status = %d, want %d", i+1, response.Code, http.StatusAccepted)
		}
	}
	if response := frontendEventRequest(e, "192.0.2.15:8080", "", body); response.Code != http.StatusTooManyRequests {
		t.Fatalf("21st request status = %d, want %d", response.Code, http.StatusTooManyRequests)
	}
	if response := frontendEventRequest(e, "192.0.2.16:8080", "", body); response.Code != http.StatusAccepted {
		t.Fatalf("different IP status = %d, want %d", response.Code, http.StatusAccepted)
	}
	now = now.Add(frontendEventWindow)
	if response := frontendEventRequest(e, "192.0.2.15:8080", "", body); response.Code != http.StatusAccepted {
		t.Fatalf("request after window status = %d, want %d", response.Code, http.StatusAccepted)
	}
}
