package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRespondJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"hello": "world"}
	respondJSON(rec, http.StatusCreated, data)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["hello"] != "world" {
		t.Errorf("expected body hello=world, got %v", body)
	}
}

func TestRespondError(t *testing.T) {
	rec := httptest.NewRecorder()
	respondError(rec, http.StatusBadRequest, "bad input")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["error"] != "bad input" {
		t.Errorf("expected error 'bad input', got %q", body["error"])
	}
}

func TestHandleCreateAlert_InvalidJSON(t *testing.T) {
	handler := NewAlertHandler(nil) // service not needed for validation
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader("not json"))

	handler.HandleCreateAlert(rec, req, 1, 1)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["error"] != "Invalid JSON body" {
		t.Errorf("expected 'Invalid JSON body', got %q", body["error"])
	}
}

func TestHandleCreateAlert_BadCondition(t *testing.T) {
	handler := NewAlertHandler(nil)
	rec := httptest.NewRecorder()
	reqBody := `{"commodity_code":"CORN","condition":"invalid","threshold_price":4.50}`
	req := httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(reqBody))

	handler.HandleCreateAlert(rec, req, 1, 1)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if !strings.Contains(body["error"], "Condition must be") {
		t.Errorf("expected condition error, got %q", body["error"])
	}
}

func TestHandleCreateAlert_ZeroPrice(t *testing.T) {
	handler := NewAlertHandler(nil)
	rec := httptest.NewRecorder()
	reqBody := `{"commodity_code":"CORN","condition":"above","threshold_price":0}`
	req := httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(reqBody))

	handler.HandleCreateAlert(rec, req, 1, 1)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if !strings.Contains(body["error"], "Threshold price") {
		t.Errorf("expected threshold price error, got %q", body["error"])
	}
}

func TestHandleCreateAlert_NegativePrice(t *testing.T) {
	handler := NewAlertHandler(nil)
	rec := httptest.NewRecorder()
	reqBody := `{"commodity_code":"CORN","condition":"above","threshold_price":-1.5}`
	req := httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(reqBody))

	handler.HandleCreateAlert(rec, req, 1, 1)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandleCreateAlert_MissingCommodityCode(t *testing.T) {
	handler := NewAlertHandler(nil)
	rec := httptest.NewRecorder()
	reqBody := `{"commodity_code":"","condition":"above","threshold_price":4.50}`
	req := httptest.NewRequest(http.MethodPost, "/alerts", strings.NewReader(reqBody))

	handler.HandleCreateAlert(rec, req, 1, 1)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if !strings.Contains(body["error"], "Commodity code is required") {
		t.Errorf("expected commodity code error, got %q", body["error"])
	}
}

func TestHandleTriggerAlert_InvalidJSON(t *testing.T) {
	handler := NewAlertHandler(nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/alerts/1/trigger", strings.NewReader("bad"))

	handler.HandleTriggerAlert(rec, req, 1, 1, 1)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandleTriggerAlert_ZeroTriggerPrice(t *testing.T) {
	handler := NewAlertHandler(nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/alerts/1/trigger", strings.NewReader(`{"trigger_price":0}`))

	handler.HandleTriggerAlert(rec, req, 1, 1, 1)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}
