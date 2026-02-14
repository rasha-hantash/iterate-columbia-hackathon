package main

import (
	"strings"
	"testing"
)

func TestFormatAnalysisAsAssistantMessage_WithSuggestions(t *testing.T) {
	analysis := &AnalysisResponse{
		Reasoning: "Portfolio analysis shows downside risk on CORN position.",
		Suggestions: []AlertSuggestion{
			{CommodityCode: "CORN", Condition: "below", ThresholdPrice: 4.20, Notes: "Stop-loss protection"},
			{CommodityCode: "CORN", Condition: "above", ThresholdPrice: 5.10, Notes: "Take-profit target"},
		},
	}
	msg := formatAnalysisAsAssistantMessage(analysis)

	if !strings.Contains(msg, "Portfolio analysis") {
		t.Error("expected reasoning in message")
	}
	if !strings.Contains(msg, "CORN below $4.20") {
		t.Errorf("expected first suggestion, got: %s", msg)
	}
	if !strings.Contains(msg, "CORN above $5.10") {
		t.Errorf("expected second suggestion, got: %s", msg)
	}
	if !strings.Contains(msg, "1.") || !strings.Contains(msg, "2.") {
		t.Error("expected numbered suggestions")
	}
}

func TestFormatAnalysisAsAssistantMessage_ReasoningOnly(t *testing.T) {
	analysis := &AnalysisResponse{
		Reasoning:   "No alerts needed at this time.",
		Suggestions: []AlertSuggestion{},
	}
	msg := formatAnalysisAsAssistantMessage(analysis)

	if !strings.Contains(msg, "No alerts needed") {
		t.Errorf("expected reasoning, got: %s", msg)
	}
	if strings.Contains(msg, "Alert Suggestions") {
		t.Error("should not contain suggestions header when empty")
	}
}

func TestFormatAnalysisAsAssistantMessage_Empty(t *testing.T) {
	analysis := &AnalysisResponse{
		Suggestions: []AlertSuggestion{},
	}
	msg := formatAnalysisAsAssistantMessage(analysis)

	if msg != "" {
		t.Errorf("expected empty message for empty analysis, got %q", msg)
	}
}
