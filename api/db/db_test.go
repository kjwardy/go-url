package db

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/kjwardy/go-url/api/logging"
)

func TestSchemaErrorIsStructuredAndSanitized(t *testing.T) {
	var output bytes.Buffer
	logger := logging.New(logging.Config{JSON: true, Output: &output})
	err := schemaError(logger, "create_urls_table", errors.New("password=secret"))
	if err == nil {
		t.Fatal("schemaError returned nil for a failed operation")
	}

	var entry map[string]interface{}
	if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &entry); err != nil {
		t.Fatalf("invalid JSON log: %v", err)
	}
	if entry["event.action"] != "database.schema" || entry["event.outcome"] != "failure" {
		t.Errorf("unexpected database event: %#v", entry)
	}
	if entry["database.operation"] != "create_urls_table" {
		t.Errorf("database.operation = %v, want create_urls_table", entry["database.operation"])
	}
	if entry["error.type"] != "database_error" {
		t.Errorf("error.type = %v, want database_error", entry["error.type"])
	}
	if bytes.Contains(output.Bytes(), []byte("secret")) {
		t.Errorf("database error leaked into log: %s", output.String())
	}
}
