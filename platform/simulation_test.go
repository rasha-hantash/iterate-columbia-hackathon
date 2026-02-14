package main

import (
	"testing"
)

func TestComputeRepresentativePrice_Basic(t *testing.T) {
	rows := []PriceFeedMessage{
		{LowPrice: 30, HighPrice: 34},
		{LowPrice: 32, HighPrice: 36},
	}
	// (30+34)/2 = 32, (32+36)/2 = 34, avg = 33
	got := computeRepresentativePrice(rows)
	if got != 33.0 {
		t.Errorf("expected 33.0, got %f", got)
	}
}

func TestComputeRepresentativePrice_PrefersMostlyPrices(t *testing.T) {
	ml1, mh1 := 31.0, 33.0
	rows := []PriceFeedMessage{
		{LowPrice: 30, HighPrice: 34, MostlyLowPrice: &ml1, MostlyHighPrice: &mh1},
		{LowPrice: 28, HighPrice: 40}, // no mostly prices, uses low/high
	}
	// Row 1: (31+33)/2 = 32, Row 2: (28+40)/2 = 34, avg = 33
	got := computeRepresentativePrice(rows)
	if got != 33.0 {
		t.Errorf("expected 33.0, got %f", got)
	}
}

func TestComputeRepresentativePrice_Empty(t *testing.T) {
	got := computeRepresentativePrice(nil)
	if got != 0 {
		t.Errorf("expected 0, got %f", got)
	}
}

func TestComputeRepresentativePrice_SingleRow(t *testing.T) {
	rows := []PriceFeedMessage{
		{LowPrice: 26, HighPrice: 30},
	}
	got := computeRepresentativePrice(rows)
	if got != 28.0 {
		t.Errorf("expected 28.0, got %f", got)
	}
}

func TestAlertShouldTrigger_Above(t *testing.T) {
	tests := []struct {
		name      string
		condition string
		threshold float64
		price     float64
		want      bool
	}{
		{"above - price exceeds threshold", "above", 35.0, 36.0, true},
		{"above - price equals threshold", "above", 35.0, 35.0, true},
		{"above - price below threshold", "above", 35.0, 34.0, false},
		{"below - price below threshold", "below", 28.0, 27.0, true},
		{"below - price equals threshold", "below", 28.0, 28.0, true},
		{"below - price above threshold", "below", 28.0, 29.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldTrigger := false
			if tt.condition == "above" && tt.price >= tt.threshold {
				shouldTrigger = true
			} else if tt.condition == "below" && tt.price <= tt.threshold {
				shouldTrigger = true
			}
			if shouldTrigger != tt.want {
				t.Errorf("condition=%s threshold=%.2f price=%.2f: got %v, want %v",
					tt.condition, tt.threshold, tt.price, shouldTrigger, tt.want)
			}
		})
	}
}

func TestGroupByDate(t *testing.T) {
	// Simulate the grouping logic used in load2024Data
	messages := []PriceFeedMessage{
		{Date: "2024-01-02", Location: "Chicago", LowPrice: 30, HighPrice: 34},
		{Date: "2024-01-02", Location: "Philadelphia", LowPrice: 32, HighPrice: 36},
		{Date: "2024-01-03", Location: "Chicago", LowPrice: 31, HighPrice: 35},
	}

	groupMap := make(map[string][]PriceFeedMessage)
	var dateOrder []string
	for _, msg := range messages {
		if _, exists := groupMap[msg.Date]; !exists {
			dateOrder = append(dateOrder, msg.Date)
		}
		groupMap[msg.Date] = append(groupMap[msg.Date], msg)
	}

	if len(dateOrder) != 2 {
		t.Fatalf("expected 2 dates, got %d", len(dateOrder))
	}
	if dateOrder[0] != "2024-01-02" {
		t.Errorf("expected first date 2024-01-02, got %s", dateOrder[0])
	}
	if len(groupMap["2024-01-02"]) != 2 {
		t.Errorf("expected 2 rows for 2024-01-02, got %d", len(groupMap["2024-01-02"]))
	}
	if len(groupMap["2024-01-03"]) != 1 {
		t.Errorf("expected 1 row for 2024-01-03, got %d", len(groupMap["2024-01-03"]))
	}
}

func TestExtractYearFromPath(t *testing.T) {
	tests := []struct {
		path string
		want int
	}{
		{"AMS_sc_terminal_daily_2023.csv", 2023},
		{"AMS_sc_terminal_daily_2024.csv", 2024},
		{"../AMS_sc_terminal_daily_2023.csv", 2023},
		{"nodate.csv", 0},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := extractYearFromPath(tt.path)
			if got != tt.want {
				t.Errorf("extractYearFromPath(%q) = %d, want %d", tt.path, got, tt.want)
			}
		})
	}
}
