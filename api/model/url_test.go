package model

import (
	"reflect"
	"testing"
	"time"
)

func TestNewURLQueries(t *testing.T) {
	queriedAt := time.Date(2026, time.September, 13, 12, 0, 0, 0, time.UTC)
	queries := newURLQueries([]string{"First", "second/parameter", "first", ""}, queriedAt)
	expected := []URLQuery{
		{URLKey: "first", QueriedAt: queriedAt},
		{URLKey: "second", QueriedAt: queriedAt},
	}

	if !reflect.DeepEqual(queries, expected) {
		t.Errorf("Expected queries %v, got %v", expected, queries)
	}
}
