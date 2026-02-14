package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type AIHandler struct {
	service    *AlertService
	apiKey     string
	httpClient *http.Client
}

func NewAIHandler(service *AlertService, apiKey string) *AIHandler {
	return &AIHandler{
		service:    service,
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

type AlertSuggestion struct {
	CommodityCode  string  `json:"commodity_code"`
	Condition      string  `json:"condition"`
	ThresholdPrice float64 `json:"threshold_price"`
	Notes          string  `json:"notes"`
}

type AnalysisResponse struct {
	Reasoning   string            `json:"reasoning"`
	Suggestions []AlertSuggestion `json:"suggestions"`
}

// Anthropic API request/response types

type anthropicRequest struct {
	Model     string            `json:"model"`
	MaxTokens int               `json:"max_tokens"`
	System    string            `json:"system"`
	Messages  []anthropicMsg    `json:"messages"`
	Tools     []anthropicTool   `json:"tools"`
}

type anthropicMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"input_schema"`
}

type anthropicResponse struct {
	Content []anthropicContentBlock `json:"content"`
	Error   *anthropicError         `json:"error,omitempty"`
}

type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type anthropicContentBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

func (h *AIHandler) HandleAnalyzePositions(w http.ResponseWriter, r *http.Request, userID, clientID int) {
	positions, err := h.service.ListPositions(userID, clientID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch positions")
		return
	}
	if len(positions) == 0 {
		respondError(w, http.StatusNotFound, "No positions found for this user")
		return
	}

	prices, err := h.service.GetCurrentPrices()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch current prices")
		return
	}

	analysis, err := h.callAnthropicAPI(positions, prices)
	if err != nil {
		log.Printf("AI analysis failed: %v", err)
		respondError(w, http.StatusBadGateway, "AI analysis is currently unavailable")
		return
	}

	respondJSON(w, http.StatusOK, analysis)
}

func (h *AIHandler) callAnthropicAPI(positions []Position, prices []PricePoint) (*AnalysisResponse, error) {
	positionsJSON, err := json.Marshal(positions)
	if err != nil {
		return nil, fmt.Errorf("marshaling positions: %w", err)
	}
	pricesJSON, err := json.Marshal(prices)
	if err != nil {
		return nil, fmt.Errorf("marshaling prices: %w", err)
	}

	userMessage := fmt.Sprintf(
		"Here are the user's current commodity positions:\n%s\n\nHere are the current market prices:\n%s\n\nPlease analyze these positions and suggest appropriate price alerts. For each suggestion, use the create_alert tool.",
		string(positionsJSON),
		string(pricesJSON),
	)

	reqBody := anthropicRequest{
		Model:     "claude-sonnet-4-5-20250929",
		MaxTokens: 1024,
		System:    "You are a commodity risk management assistant. Analyze the user's commodity positions and current market prices, then suggest appropriate price alerts to help manage risk. Consider stop-loss alerts for downside protection and take-profit alerts for upside targets. Use the create_alert tool for each suggestion. Provide brief reasoning for your analysis.",
		Messages: []anthropicMsg{
			{Role: "user", Content: userMessage},
		},
		Tools: []anthropicTool{
			{
				Name:        "create_alert",
				Description: "Create a price alert for a commodity position",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"commodity_code": map[string]interface{}{
							"type":        "string",
							"description": "The commodity code (e.g., CORN, WHEAT, SOYBEAN_OIL)",
						},
						"condition": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"above", "below"},
							"description": "Whether the alert triggers when price goes above or below the threshold",
						},
						"threshold_price": map[string]interface{}{
							"type":        "number",
							"description": "The price threshold that triggers the alert",
						},
						"notes": map[string]interface{}{
							"type":        "string",
							"description": "Explanation of why this alert is recommended",
						},
					},
					"required": []string{"commodity_code", "condition", "threshold_price", "notes"},
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling API request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating API request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", h.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling Anthropic API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading API response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Anthropic API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp anthropicResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("parsing API response: %w", err)
	}

	return parseAnthropicResponse(&apiResp)
}

func parseAnthropicResponse(resp *anthropicResponse) (*AnalysisResponse, error) {
	result := &AnalysisResponse{
		Suggestions: []AlertSuggestion{},
	}

	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			if result.Reasoning != "" {
				result.Reasoning += "\n"
			}
			result.Reasoning += block.Text
		case "tool_use":
			if block.Name == "create_alert" {
				var suggestion AlertSuggestion
				if err := json.Unmarshal(block.Input, &suggestion); err != nil {
					return nil, fmt.Errorf("parsing tool_use input: %w", err)
				}
				result.Suggestions = append(result.Suggestions, suggestion)
			}
		}
	}

	return result, nil
}
