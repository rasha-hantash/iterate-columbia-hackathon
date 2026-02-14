package main

import (
	"encoding/json"
	"testing"
)

func TestParseAnthropicResponse_TextAndToolUse(t *testing.T) {
	resp := &anthropicResponse{
		Content: []anthropicContentBlock{
			{
				Type: "text",
				Text: "Based on your positions, here are my recommendations.",
			},
			{
				Type:  "tool_use",
				Name:  "create_alert",
				Input: json.RawMessage(`{"commodity_code":"CORN","condition":"below","threshold_price":4.20,"notes":"Stop-loss for corn"}`),
			},
			{
				Type:  "tool_use",
				Name:  "create_alert",
				Input: json.RawMessage(`{"commodity_code":"CORN","condition":"above","threshold_price":42.00,"notes":"Take-profit for corn short"}`),
			},
		},
	}

	result, err := parseAnthropicResponse(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Reasoning != "Based on your positions, here are my recommendations." {
		t.Errorf("unexpected reasoning: %q", result.Reasoning)
	}

	if len(result.Suggestions) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(result.Suggestions))
	}

	if result.Suggestions[0].CommodityCode != "CORN" {
		t.Errorf("expected CORN, got %q", result.Suggestions[0].CommodityCode)
	}
	if result.Suggestions[0].Condition != "below" {
		t.Errorf("expected below, got %q", result.Suggestions[0].Condition)
	}
	if result.Suggestions[0].ThresholdPrice != 4.20 {
		t.Errorf("expected 4.20, got %f", result.Suggestions[0].ThresholdPrice)
	}
	if result.Suggestions[0].Notes != "Stop-loss for corn" {
		t.Errorf("expected 'Stop-loss for corn', got %q", result.Suggestions[0].Notes)
	}

	if result.Suggestions[1].CommodityCode != "CORN" {
		t.Errorf("expected CORN, got %q", result.Suggestions[1].CommodityCode)
	}
	if result.Suggestions[1].Condition != "above" {
		t.Errorf("expected above, got %q", result.Suggestions[1].Condition)
	}
}

func TestParseAnthropicResponse_TextOnly(t *testing.T) {
	resp := &anthropicResponse{
		Content: []anthropicContentBlock{
			{
				Type: "text",
				Text: "I don't have enough data to make suggestions.",
			},
		},
	}

	result, err := parseAnthropicResponse(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Reasoning != "I don't have enough data to make suggestions." {
		t.Errorf("unexpected reasoning: %q", result.Reasoning)
	}
	if len(result.Suggestions) != 0 {
		t.Errorf("expected 0 suggestions, got %d", len(result.Suggestions))
	}
}

func TestParseAnthropicResponse_EmptyContent(t *testing.T) {
	resp := &anthropicResponse{
		Content: []anthropicContentBlock{},
	}

	result, err := parseAnthropicResponse(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Reasoning != "" {
		t.Errorf("expected empty reasoning, got %q", result.Reasoning)
	}
	if len(result.Suggestions) != 0 {
		t.Errorf("expected 0 suggestions, got %d", len(result.Suggestions))
	}
}

func TestParseAnthropicResponse_MultipleTextBlocks(t *testing.T) {
	resp := &anthropicResponse{
		Content: []anthropicContentBlock{
			{Type: "text", Text: "First part."},
			{Type: "text", Text: "Second part."},
		},
	}

	result, err := parseAnthropicResponse(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Reasoning != "First part.\nSecond part." {
		t.Errorf("unexpected reasoning: %q", result.Reasoning)
	}
}

func TestParseAnthropicResponse_InvalidToolInput(t *testing.T) {
	resp := &anthropicResponse{
		Content: []anthropicContentBlock{
			{
				Type:  "tool_use",
				Name:  "create_alert",
				Input: json.RawMessage(`not valid json`),
			},
		},
	}

	_, err := parseAnthropicResponse(resp)
	if err == nil {
		t.Error("expected error for invalid tool input JSON")
	}
}

func TestParseAnthropicResponse_IgnoresUnknownToolNames(t *testing.T) {
	resp := &anthropicResponse{
		Content: []anthropicContentBlock{
			{
				Type:  "tool_use",
				Name:  "unknown_tool",
				Input: json.RawMessage(`{"key":"value"}`),
			},
		},
	}

	result, err := parseAnthropicResponse(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Suggestions) != 0 {
		t.Errorf("expected 0 suggestions for unknown tool, got %d", len(result.Suggestions))
	}
}
