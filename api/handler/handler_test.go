package handler

import "testing"

func TestValidateKey(t *testing.T) {
	tables := []struct {
		key   string
		valid bool
	}{
		{"key", true},
		{"key with spaces", true},
		{"key  with  spaces", true},
		{"key-with_spaces 123", true},
		{" leading space", false},
		{"trailing space ", false},
		{"   ", false},
		{"key/with/path", false},
		{"key$invalid", false},
	}

	for _, table := range tables {
		if valid := ValidateKey(table.key); valid != table.valid {
			t.Errorf("Expected ValidateKey(%q) to be %t", table.key, table.valid)
		}
	}
}

func TestHistoryLimit(t *testing.T) {
	tables := []struct {
		value string
		limit int
		valid bool
	}{
		{"", 25, true},
		{"25", 25, true},
		{"50", 50, true},
		{"100", 100, true},
		{"10", 0, false},
		{"invalid", 0, false},
	}

	for _, table := range tables {
		limit, valid := historyLimit(table.value)
		if limit != table.limit || valid != table.valid {
			t.Errorf("Expected historyLimit(%q) to return (%d, %t), got (%d, %t)", table.value, table.limit, table.valid, limit, valid)
		}
	}
}

func TestValidateKeyPath(t *testing.T) {
	tables := []struct {
		key   string
		valid bool
	}{
		{"key with spaces", true},
		{"key with spaces/path value", true},
		{"key with spaces/another/path", true},
		{"key with spaces/", false},
		{"key with spaces/$invalid", false},
	}

	for _, table := range tables {
		if valid := ValidateKeyPath(table.key); valid != table.valid {
			t.Errorf("Expected ValidateKeyPath(%q) to be %t", table.key, table.valid)
		}
	}
}
