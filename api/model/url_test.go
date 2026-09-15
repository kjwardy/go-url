package model

import (
	"reflect"
	"testing"
	"time"
)

func TestNewURLQueries(t *testing.T) {
	queriedAt := time.Date(2026, time.September, 13, 12, 0, 0, 0, time.UTC)
	queries := newURLQueries([]string{"First", "second/parameter", "first", ""}, queriedAt, true)
	expected := []URLQuery{
		{URLKey: "first", QueriedAt: queriedAt, Successful: true},
		{URLKey: "second", QueriedAt: queriedAt, Successful: true},
	}

	if !reflect.DeepEqual(queries, expected) {
		t.Errorf("Expected queries %v, got %v", expected, queries)
	}

	failed := newURLQueries([]string{"missing"}, queriedAt, false)
	if len(failed) != 1 || failed[0].Successful {
		t.Errorf("Expected an unsuccessful query, got %v", failed)
	}
}
