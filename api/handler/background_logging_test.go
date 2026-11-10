package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/kjwardy/go-url/api/logging"
)

func TestBackgroundTaskFailureIsStructuredAndSanitized(t *testing.T) {
	var output bytes.Buffer
	h := &Handler{Logger: logging.New(logging.Config{JSON: true, Output: &output})}
	h.logBackgroundResult("record_successful_queries", errors.New("database password=secret"))

	var entry map[string]interface{}
	if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &entry); err != nil {
		t.Fatalf("invalid JSON log: %v", err)
	}
	if entry["event.action"] != "background.task" || entry["event.outcome"] != "failure" {
		t.Errorf("unexpected background task event: %#v", entry)
	}
	if entry["background.task"] != "record_successful_queries" {
		t.Errorf("background.task = %v, want record_successful_queries", entry["background.task"])
	}
	if entry["error.type"] != "database_error" {
		t.Errorf("error.type = %v, want database_error", entry["error.type"])
	}
	if bytes.Contains(output.Bytes(), []byte("secret")) {
		t.Errorf("database error leaked into log: %s", output.String())
	}
}
