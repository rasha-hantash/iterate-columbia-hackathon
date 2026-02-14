package main

import (
	"testing"
)

func TestParseDate_MMDDYYYY(t *testing.T) {
	d, err := parseDate("12/31/2024")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Year() != 2024 || d.Month() != 12 || d.Day() != 31 {
		t.Errorf("expected 2024-12-31, got %v", d)
	}
}

func TestParseDate_ISO(t *testing.T) {
	d, err := parseDate("2024-12-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Year() != 2024 || d.Month() != 12 || d.Day() != 31 {
		t.Errorf("expected 2024-12-31, got %v", d)
	}
}

func TestParseDate_Empty(t *testing.T) {
	_, err := parseDate("")
	if err == nil {
		t.Error("expected error for empty date")
	}
}

func TestParseDate_Invalid(t *testing.T) {
	_, err := parseDate("not-a-date")
	if err == nil {
		t.Error("expected error for invalid date")
	}
}

func TestNullableString(t *testing.T) {
	tests := []struct {
		input    string
		wantNil  bool
	}{
		{"", true},
		{"N/A", true},
		{"hello", false},
		{"  ", false}, // spaces are not trimmed by this function
	}
	for _, tt := range tests {
		result := nullableString(tt.input)
		if tt.wantNil && result != nil {
			t.Errorf("nullableString(%q) = %v, want nil", tt.input, *result)
		}
		if !tt.wantNil && result == nil {
			t.Errorf("nullableString(%q) = nil, want non-nil", tt.input)
		}
		if !tt.wantNil && result != nil && *result != tt.input {
			t.Errorf("nullableString(%q) = %q, want %q", tt.input, *result, tt.input)
		}
	}
}

func TestNullableFloat(t *testing.T) {
	tests := []struct {
		input   string
		wantNil bool
		wantVal float64
	}{
		{"", true, 0},
		{"abc", true, 0},
		{"35", false, 35.0},
		{"4.50", false, 4.50},
		{"0", false, 0},
	}
	for _, tt := range tests {
		result := nullableFloat(tt.input)
		if tt.wantNil && result != nil {
			t.Errorf("nullableFloat(%q) = %v, want nil", tt.input, *result)
		}
		if !tt.wantNil && result == nil {
			t.Errorf("nullableFloat(%q) = nil, want %f", tt.input, tt.wantVal)
		}
		if !tt.wantNil && result != nil && *result != tt.wantVal {
			t.Errorf("nullableFloat(%q) = %f, want %f", tt.input, *result, tt.wantVal)
		}
	}
}

func TestGetCol(t *testing.T) {
	record := []string{"2024-01-01", "New York", "Corn"}
	colIdx := map[string]int{"date": 0, "location": 1, "commodity": 2}

	if got := getCol(record, colIdx, "location"); got != "New York" {
		t.Errorf("getCol(location) = %q, want 'New York'", got)
	}
	if got := getCol(record, colIdx, "missing"); got != "" {
		t.Errorf("getCol(missing) = %q, want empty", got)
	}
	if got := getCol(record, colIdx, "date"); got != "2024-01-01" {
		t.Errorf("getCol(date) = %q, want '2024-01-01'", got)
	}
}

func TestGetCol_OutOfBounds(t *testing.T) {
	record := []string{"only_one"}
	colIdx := map[string]int{"a": 0, "b": 5}

	if got := getCol(record, colIdx, "b"); got != "" {
		t.Errorf("getCol out-of-bounds = %q, want empty", got)
	}
}
