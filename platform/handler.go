package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
)

type AlertHandler struct {
	service *AlertService
}

func NewAlertHandler(service *AlertService) *AlertHandler {
	return &AlertHandler{service: service}
}

type createAlertRequest struct {
	CommodityCode  string  `json:"commodity_code"`
	Condition      string  `json:"condition"`
	ThresholdPrice float64 `json:"threshold_price"`
	Notes          string  `json:"notes"`
}

type triggerAlertRequest struct {
	TriggerPrice float64 `json:"trigger_price"`
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *AlertHandler) HandleCreateAlert(w http.ResponseWriter, r *http.Request, userID, clientID int) {
	var req createAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if req.Condition != "above" && req.Condition != "below" {
		respondError(w, http.StatusBadRequest, "Condition must be 'above' or 'below'")
		return
	}

	if req.ThresholdPrice <= 0 || math.IsInf(req.ThresholdPrice, 0) || math.IsNaN(req.ThresholdPrice) {
		respondError(w, http.StatusBadRequest, "Threshold price must be a finite positive number")
		return
	}

	if req.CommodityCode == "" {
		respondError(w, http.StatusBadRequest, "Commodity code is required")
		return
	}

	alert, err := h.service.CreateAlert(userID, clientID, CreateAlertParams{
		CommodityCode:  req.CommodityCode,
		Condition:      req.Condition,
		ThresholdPrice: req.ThresholdPrice,
		Notes:          req.Notes,
	})
	if err != nil {
		if errors.Is(err, ErrInvalidCommodity) {
			respondError(w, http.StatusBadRequest, "Invalid commodity code: "+req.CommodityCode)
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create alert")
		return
	}

	respondJSON(w, http.StatusCreated, alert)
}

func (h *AlertHandler) HandleListAlerts(w http.ResponseWriter, r *http.Request, clientID int) {
	status := r.URL.Query().Get("status")
	commodityCode := r.URL.Query().Get("commodity_code")

	alerts, err := h.service.ListAlerts(clientID, status, commodityCode)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list alerts")
		return
	}

	respondJSON(w, http.StatusOK, alerts)
}

func (h *AlertHandler) HandleTriggerAlert(w http.ResponseWriter, r *http.Request, alertID, userID, clientID int) {
	var req triggerAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if req.TriggerPrice <= 0 || math.IsInf(req.TriggerPrice, 0) || math.IsNaN(req.TriggerPrice) {
		respondError(w, http.StatusBadRequest, "Trigger price must be a finite positive number")
		return
	}

	alert, err := h.service.TriggerAlert(alertID, clientID, userID, req.TriggerPrice)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			respondError(w, http.StatusNotFound, "Alert not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to trigger alert")
		return
	}

	respondJSON(w, http.StatusOK, alert)
}

func (h *AlertHandler) HandleListCommodities(w http.ResponseWriter, r *http.Request) {
	commodities, err := h.service.ListCommodities()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list commodities")
		return
	}
	respondJSON(w, http.StatusOK, commodities)
}

func (h *AlertHandler) HandleListPositions(w http.ResponseWriter, r *http.Request, userID, clientID int) {
	positions, err := h.service.ListPositions(userID, clientID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to list positions")
		return
	}
	respondJSON(w, http.StatusOK, positions)
}

func (h *AlertHandler) HandleGetPrices(w http.ResponseWriter, r *http.Request) {
	prices, err := h.service.GetCurrentPrices()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get prices")
		return
	}
	respondJSON(w, http.StatusOK, prices)
}

func (h *AlertHandler) HandleGetMonthlyAnalysis(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	if yearStr == "" {
		yearStr = "2023"
	}
	year := 2023
	if _, err := fmt.Sscanf(yearStr, "%d", &year); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid year parameter")
		return
	}

	commodity := r.URL.Query().Get("commodity")
	if commodity == "" {
		commodity = "corn"
	}

	summaries, err := h.service.GetMonthlyPriceAnalysis(year, commodity)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get monthly analysis")
		return
	}
	respondJSON(w, http.StatusOK, summaries)
}
