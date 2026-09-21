package handler

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kjwardy/go-url/api/logging"
	"github.com/kjwardy/go-url/api/model"
	"github.com/labstack/echo"
)

func TestLogQueryOutcomesPerNormalizedRequestedKey(t *testing.T) {
	var output bytes.Buffer
	logger := logging.New(logging.Config{JSON: true, Output: &output})
	h := &Handler{Logger: logger}
	e := echo.New()
	e.Use(logger.Middleware())
	e.GET("/test", func(c echo.Context) error {
		h.logQueryOutcomes(c, []string{"Handbook/section", "missing", "handbook/other"}, []*model.URL{{Key: "handbook"}})
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(echo.HeaderXRequestID, "query-request")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	scanner := bufio.NewScanner(&output)
	var queryEntries []map[string]interface{}
	for scanner.Scan() {
		var entry map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			t.Fatalf("invalid JSON log: %v", err)
		}
		if entry["event.action"] == "go_url.query" {
			queryEntries = append(queryEntries, entry)
		}
	}
	if len(queryEntries) != 2 {
		t.Fatalf("got %d query entries, want 2", len(queryEntries))
	}
	assertQueryEntry(t, queryEntries[0], "handbook", true)
	assertQueryEntry(t, queryEntries[1], "missing", false)
}

func assertQueryEntry(t *testing.T, entry map[string]interface{}, key string, successful bool) {
	t.Helper()
	if entry["http.request.id"] != "query-request" {
		t.Errorf("request ID = %#v, want query-request", entry["http.request.id"])
	}
	if entry["go_url.query.key"] != key {
		t.Errorf("key = %#v, want %q", entry["go_url.query.key"], key)
	}
	if entry["go_url.query.successful"] != successful {
		t.Errorf("successful = %#v, want %t", entry["go_url.query.successful"], successful)
	}
}
