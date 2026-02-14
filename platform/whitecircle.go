package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const whiteCircleBaseURL = "https://us.whitecircle.ai/api/session/check"
const whiteCircleAPIVersion = "2025-12-01"

type WhiteCircleClient struct {
	apiKey       string
	deploymentID string
	httpClient   *http.Client
}

func NewWhiteCircleClient(apiKey, deploymentID string) *WhiteCircleClient {
	return &WhiteCircleClient{
		apiKey:       apiKey,
		deploymentID: deploymentID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type wcMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type wcRequest struct {
	DeploymentID string      `json:"deployment_id"`
	Messages     []wcMessage `json:"messages"`
}

type wcPolicyResult struct {
	Flagged       bool     `json:"flagged"`
	FlaggedSource []string `json:"flagged_source"`
	Name          string   `json:"name"`
}

type wcResponse struct {
	Flagged           bool                      `json:"flagged"`
	InternalSessionID string                    `json:"internal_session_id"`
	Policies          map[string]wcPolicyResult `json:"policies"`
}

// CheckSession sends the conversation to White Circle for policy evaluation.
// Designed to be called as a goroutine (fire-and-forget). Logs the result.
func (wc *WhiteCircleClient) CheckSession(ctx context.Context, userMessage string, analysis *AnalysisResponse) {
	assistantMessage := formatAnalysisAsAssistantMessage(analysis)

	reqBody := wcRequest{
		DeploymentID: wc.deploymentID,
		Messages: []wcMessage{
			{Role: "user", Content: userMessage},
			{Role: "assistant", Content: assistantMessage},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[WhiteCircle] ERROR marshaling request: %v", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", whiteCircleBaseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("[WhiteCircle] ERROR creating request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+wc.apiKey)
	req.Header.Set("whitecircle-version", whiteCircleAPIVersion)

	resp, err := wc.httpClient.Do(req)
	if err != nil {
		log.Printf("[WhiteCircle] ERROR calling API: %v", err)
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[WhiteCircle] ERROR reading response: %v", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("[WhiteCircle] API returned status %d: %s", resp.StatusCode, string(respBody))
		return
	}

	var wcResp wcResponse
	if err := json.Unmarshal(respBody, &wcResp); err != nil {
		log.Printf("[WhiteCircle] ERROR parsing response: %v", err)
		return
	}

	log.Printf("[WhiteCircle] session=%s flagged=%t", wcResp.InternalSessionID, wcResp.Flagged)
	for policyID, policy := range wcResp.Policies {
		if policy.Flagged {
			log.Printf("[WhiteCircle]   FLAGGED policy=%s name=%q sources=%v",
				policyID, policy.Name, policy.FlaggedSource)
		}
	}
}

// formatAnalysisAsAssistantMessage converts the structured AnalysisResponse
// into a plain-text message suitable for White Circle's messages array.
func formatAnalysisAsAssistantMessage(analysis *AnalysisResponse) string {
	var msg string
	if analysis.Reasoning != "" {
		msg = analysis.Reasoning + "\n\n"
	}
	if len(analysis.Suggestions) > 0 {
		msg += "Alert Suggestions:\n"
		for i, s := range analysis.Suggestions {
			msg += fmt.Sprintf("%d. %s %s $%.2f - %s\n",
				i+1, s.CommodityCode, s.Condition, s.ThresholdPrice, s.Notes)
		}
	}
	return msg
}
