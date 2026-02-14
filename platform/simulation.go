package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// PriceFeedMessage represents a single market data row arriving through the simulated feed.
type PriceFeedMessage struct {
	Date            string   `json:"date"`
	Location        string   `json:"location"`
	LowPrice        float64  `json:"low_price"`
	HighPrice       float64  `json:"high_price"`
	MostlyLowPrice  *float64 `json:"mostly_low_price,omitempty"`
	MostlyHighPrice *float64 `json:"mostly_high_price,omitempty"`
	Properties      string   `json:"properties"`
}

// SimulationResults tracks the state and outcomes of a simulation run.
type SimulationResults struct {
	Status          string            `json:"status"` // "running", "completed", "stopped"
	StartedAt       time.Time         `json:"started_at"`
	CurrentDate     string            `json:"current_date"`
	TotalDates      int               `json:"total_dates"`
	ProcessedDates  int               `json:"processed_dates"`
	TotalRows       int               `json:"total_rows"`
	ProcessedRows   int               `json:"processed_rows"`
	AlertsTriggered int               `json:"alerts_triggered"`
	Events          []SimulationEvent `json:"events"`
}

// SimulationEvent records what happened when processing a single date.
type SimulationEvent struct {
	Date                string               `json:"date"`
	RepresentativePrice float64              `json:"representative_price"`
	RowCount            int                  `json:"row_count"`
	TriggeredAlerts     []TriggeredAlertInfo `json:"triggered_alerts,omitempty"`
}

// TriggeredAlertInfo describes an alert that was triggered during simulation.
type TriggeredAlertInfo struct {
	AlertID        int     `json:"alert_id"`
	Condition      string  `json:"condition"`
	ThresholdPrice float64 `json:"threshold_price"`
	TriggerPrice   float64 `json:"trigger_price"`
	Notes          string  `json:"notes"`
}

// SimulationManager controls the lifecycle of a price feed simulation.
type SimulationManager struct {
	service *AlertService
	db      *sql.DB
	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
	results *SimulationResults
}

func NewSimulationManager(service *AlertService, db *sql.DB) *SimulationManager {
	return &SimulationManager{
		service: service,
		db:      db,
	}
}

// dateGroup holds all market data rows for a single date.
type dateGroup struct {
	date string
	rows []PriceFeedMessage
}

// HandleStartSimulation starts the simulation of 2024 market data processing.
func (sm *SimulationManager) HandleStartSimulation(w http.ResponseWriter, r *http.Request) {
	sm.mu.Lock()
	if sm.running {
		sm.mu.Unlock()
		respondError(w, http.StatusConflict, "Simulation is already running")
		return
	}

	// Parse user_id from query param (needed since this may be called without header)
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		userIDStr = r.Header.Get("X-User-ID")
	}
	if userIDStr == "" {
		sm.mu.Unlock()
		respondError(w, http.StatusBadRequest, "user_id query parameter is required")
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		sm.mu.Unlock()
		respondError(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	// Look up user to get client_id
	var clientID int
	err = sm.db.QueryRow("SELECT client_id FROM users WHERE id = $1 AND is_active = TRUE", userID).Scan(&clientID)
	if err != nil {
		sm.mu.Unlock()
		respondError(w, http.StatusBadRequest, "Invalid user")
		return
	}

	// Parse speed
	speedMs := 500
	if s := r.URL.Query().Get("speed"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v >= 10 && v <= 10000 {
			speedMs = v
		}
	}

	// Import 2024 CSV on demand if not already loaded
	if err := import2024MarketData(sm.db); err != nil {
		log.Printf("[Simulation] Warning: could not import 2024 CSV: %v", err)
	}

	// Load 2024 data from market_data table
	groups, totalRows, err := sm.load2024Data()
	if err != nil {
		sm.mu.Unlock()
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to load 2024 data: %v", err))
		return
	}
	if len(groups) == 0 {
		sm.mu.Unlock()
		respondError(w, http.StatusNotFound, "No 2024 market data found. Ensure the 2024 CSV file is in the project root.")
		return
	}

	sm.running = true
	sm.stopCh = make(chan struct{})
	sm.results = &SimulationResults{
		Status:     "running",
		StartedAt:  time.Now(),
		TotalDates: len(groups),
		TotalRows:  totalRows,
		Events:     []SimulationEvent{},
	}
	sm.mu.Unlock()

	// Launch the simulation pipeline in the background
	go sm.runSimulation(groups, userID, clientID, time.Duration(speedMs)*time.Millisecond)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":     "Simulation started",
		"total_dates": len(groups),
		"total_rows":  totalRows,
		"speed_ms":    speedMs,
	})
}

// HandleGetStatus returns the current simulation state.
func (sm *SimulationManager) HandleGetStatus(w http.ResponseWriter, r *http.Request) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.results == nil {
		respondJSON(w, http.StatusOK, map[string]string{"status": "idle"})
		return
	}

	respondJSON(w, http.StatusOK, sm.results)
}

// HandleStopSimulation stops a running simulation.
func (sm *SimulationManager) HandleStopSimulation(w http.ResponseWriter, r *http.Request) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.running {
		respondError(w, http.StatusConflict, "No simulation is running")
		return
	}

	close(sm.stopCh)
	sm.running = false
	sm.results.Status = "stopped"

	respondJSON(w, http.StatusOK, map[string]string{"message": "Simulation stopped"})
}

// HandleResetAlerts resets all triggered CORN alerts back to active for the user's client.
func (sm *SimulationManager) HandleResetAlerts(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		userIDStr = r.Header.Get("X-User-ID")
	}
	if userIDStr == "" {
		respondError(w, http.StatusBadRequest, "user_id query parameter is required")
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user_id")
		return
	}

	var clientID int
	err = sm.db.QueryRow("SELECT client_id FROM users WHERE id = $1 AND is_active = TRUE", userID).Scan(&clientID)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user")
		return
	}

	count, err := sm.service.ResetAlerts(clientID, "CORN")
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to reset alerts: %v", err))
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":       "Alerts reset",
		"alerts_reset":  count,
	})
}

// load2024Data queries market_data for all 2024 corn rows and groups them by date.
func (sm *SimulationManager) load2024Data() ([]dateGroup, int, error) {
	rows, err := sm.db.Query(`
		SELECT report_date::text, location,
		       low_price, high_price, mostly_low_price, mostly_high_price,
		       COALESCE(properties, '')
		FROM market_data
		WHERE EXTRACT(YEAR FROM report_date) = 2024
		  AND commodity ILIKE '%corn%'
		  AND low_price IS NOT NULL
		  AND high_price IS NOT NULL
		ORDER BY report_date ASC, location ASC`)
	if err != nil {
		return nil, 0, fmt.Errorf("querying 2024 market data: %w", err)
	}
	defer rows.Close()

	// Group rows by date
	groupMap := make(map[string][]PriceFeedMessage)
	var dateOrder []string
	totalRows := 0

	for rows.Next() {
		var msg PriceFeedMessage
		if err := rows.Scan(&msg.Date, &msg.Location,
			&msg.LowPrice, &msg.HighPrice, &msg.MostlyLowPrice, &msg.MostlyHighPrice,
			&msg.Properties); err != nil {
			return nil, 0, fmt.Errorf("scanning market data row: %w", err)
		}

		if _, exists := groupMap[msg.Date]; !exists {
			dateOrder = append(dateOrder, msg.Date)
		}
		groupMap[msg.Date] = append(groupMap[msg.Date], msg)
		totalRows++
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterating market data rows: %w", err)
	}

	groups := make([]dateGroup, 0, len(dateOrder))
	for _, date := range dateOrder {
		groups = append(groups, dateGroup{date: date, rows: groupMap[date]})
	}

	return groups, totalRows, nil
}

// runSimulation is the core goroutine that processes the price feed.
func (sm *SimulationManager) runSimulation(groups []dateGroup, userID, clientID int, speed time.Duration) {
	defer func() {
		sm.mu.Lock()
		sm.running = false
		if sm.results.Status == "running" {
			sm.results.Status = "completed"
		}
		sm.mu.Unlock()
	}()

	// Channel-based pipeline: producer sends messages, consumer processes them
	feedCh := make(chan dateGroup, 1)

	// Producer goroutine: feeds date groups through the channel
	go func() {
		defer close(feedCh)
		for _, group := range groups {
			select {
			case <-sm.stopCh:
				return
			case feedCh <- group:
			}
			// Delay between date groups to simulate real-time arrival
			select {
			case <-sm.stopCh:
				return
			case <-time.After(speed):
			}
		}
	}()

	// Consumer: process each date group from the channel
	for group := range feedCh {
		select {
		case <-sm.stopCh:
			return
		default:
		}

		event := sm.processDateGroup(group, userID, clientID)

		sm.mu.Lock()
		sm.results.CurrentDate = group.date
		sm.results.ProcessedDates++
		sm.results.ProcessedRows += len(group.rows)
		sm.results.AlertsTriggered += len(event.TriggeredAlerts)
		sm.results.Events = append(sm.results.Events, event)
		sm.mu.Unlock()

		if len(event.TriggeredAlerts) > 0 {
			log.Printf("[Simulation] Date: %s | Price: $%.2f | Triggered: %d",
				group.date, event.RepresentativePrice, len(event.TriggeredAlerts))
		}
	}
}

// processDateGroup computes the representative price for a date and checks alerts.
func (sm *SimulationManager) processDateGroup(group dateGroup, userID, clientID int) SimulationEvent {
	repPrice := computeRepresentativePrice(group.rows)

	event := SimulationEvent{
		Date:                group.date,
		RepresentativePrice: repPrice,
		RowCount:            len(group.rows),
	}

	// Check active alerts
	triggered, err := sm.checkAndTriggerAlerts(clientID, userID, repPrice)
	if err != nil {
		log.Printf("[Simulation] Error checking alerts for date %s: %v", group.date, err)
		return event
	}
	event.TriggeredAlerts = triggered

	return event
}

// computeRepresentativePrice calculates the daily representative price
// as the average mid-price across all rows for that date.
func computeRepresentativePrice(rows []PriceFeedMessage) float64 {
	if len(rows) == 0 {
		return 0
	}

	var total float64
	for _, row := range rows {
		if row.MostlyLowPrice != nil && row.MostlyHighPrice != nil {
			total += (*row.MostlyLowPrice + *row.MostlyHighPrice) / 2.0
		} else {
			total += (row.LowPrice + row.HighPrice) / 2.0
		}
	}
	return total / float64(len(rows))
}

// checkAndTriggerAlerts queries active CORN alerts and triggers any whose conditions are met.
func (sm *SimulationManager) checkAndTriggerAlerts(clientID, userID int, price float64) ([]TriggeredAlertInfo, error) {
	alerts, err := sm.service.ListAlerts(clientID, "active", "CORN")
	if err != nil {
		return nil, fmt.Errorf("listing active alerts: %w", err)
	}

	var triggered []TriggeredAlertInfo
	for _, alert := range alerts {
		shouldTrigger := false
		if alert.Condition == "above" && price >= alert.ThresholdPrice {
			shouldTrigger = true
		} else if alert.Condition == "below" && price <= alert.ThresholdPrice {
			shouldTrigger = true
		}

		if shouldTrigger {
			_, err := sm.service.TriggerAlert(alert.ID, clientID, userID, price)
			if err != nil {
				log.Printf("[Simulation] Failed to trigger alert %d: %v", alert.ID, err)
				continue
			}

			triggered = append(triggered, TriggeredAlertInfo{
				AlertID:        alert.ID,
				Condition:      alert.Condition,
				ThresholdPrice: alert.ThresholdPrice,
				TriggerPrice:   price,
				Notes:          alert.Notes,
			})

			log.Printf("[Simulation] ALERT TRIGGERED: ID=%d, Condition=%s %s $%.2f, Price=$%.2f",
				alert.ID, alert.CommodityCode, alert.Condition, alert.ThresholdPrice, price)
		}
	}

	return triggered, nil
}

// registerSimulationRoutes adds all simulation-related HTTP endpoints to the mux.
func registerSimulationRoutes(mux *http.ServeMux, sm *SimulationManager) {
	mux.HandleFunc("/simulation/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		sm.HandleStartSimulation(w, r)
	})

	mux.HandleFunc("/simulation/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		sm.HandleGetStatus(w, r)
	})

	mux.HandleFunc("/simulation/stop", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		sm.HandleStopSimulation(w, r)
	})

	mux.HandleFunc("/simulation/reset", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		sm.HandleResetAlerts(w, r)
	})
}

