package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrInvalidCommodity = errors.New("invalid commodity code")
)

type Commodity struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Unit string `json:"unit"`
}

type PriceAlert struct {
	ID              int        `json:"id"`
	ClientID        int        `json:"client_id"`
	UserID          int        `json:"user_id"`
	CommodityID     int        `json:"commodity_id"`
	CommodityCode   string     `json:"commodity_code"`
	CommodityName   string     `json:"commodity_name"`
	Condition       string     `json:"condition"`
	ThresholdPrice  float64    `json:"threshold_price"`
	Status          string     `json:"status"`
	Notes           string     `json:"notes"`
	TriggeredCount  int        `json:"triggered_count"`
	LastTriggeredAt *time.Time `json:"last_triggered_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateAlertParams struct {
	CommodityCode  string  `json:"commodity_code"`
	Condition      string  `json:"condition"`
	ThresholdPrice float64 `json:"threshold_price"`
	Notes          string  `json:"notes"`
}

type AlertService struct {
	db *sql.DB
}

func NewAlertService(db *sql.DB) *AlertService {
	return &AlertService{db: db}
}

func (s *AlertService) CreateAlert(userID, clientID int, params CreateAlertParams) (*PriceAlert, error) {
	var commodity Commodity
	err := s.db.QueryRow(
		"SELECT id, code, name, unit FROM commodities WHERE code = $1",
		params.CommodityCode,
	).Scan(&commodity.ID, &commodity.Code, &commodity.Name, &commodity.Unit)
	if err == sql.ErrNoRows {
		return nil, ErrInvalidCommodity
	}
	if err != nil {
		return nil, fmt.Errorf("looking up commodity: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	var alert PriceAlert
	err = tx.QueryRow(`
		INSERT INTO price_alerts (client_id, user_id, commodity_id, condition, threshold_price, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, client_id, user_id, commodity_id, condition, threshold_price,
		          status, notes, triggered_count, last_triggered_at, created_at, updated_at`,
		clientID, userID, commodity.ID, params.Condition, params.ThresholdPrice, params.Notes,
	).Scan(
		&alert.ID, &alert.ClientID, &alert.UserID, &alert.CommodityID,
		&alert.Condition, &alert.ThresholdPrice, &alert.Status, &alert.Notes,
		&alert.TriggeredCount, &alert.LastTriggeredAt, &alert.CreatedAt, &alert.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting alert: %w", err)
	}

	alert.CommodityCode = commodity.Code
	alert.CommodityName = commodity.Name

	_, err = tx.Exec(`
		INSERT INTO alert_history (alert_id, changed_by_user_id, change_type, new_status, new_threshold)
		VALUES ($1, $2, 'created', 'active', $3)`,
		alert.ID, userID, params.ThresholdPrice,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting alert history: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	return &alert, nil
}

func (s *AlertService) ListAlerts(clientID int, status, commodityCode string) ([]PriceAlert, error) {
	query := `
		SELECT pa.id, pa.client_id, pa.user_id, pa.commodity_id,
		       c.code, c.name,
		       pa.condition, pa.threshold_price, pa.status, pa.notes,
		       pa.triggered_count, pa.last_triggered_at, pa.created_at, pa.updated_at
		FROM price_alerts pa
		JOIN commodities c ON c.id = pa.commodity_id
		WHERE pa.client_id = $1 AND pa.deleted_at IS NULL`

	args := []interface{}{clientID}
	argIdx := 2

	if status != "" {
		query += fmt.Sprintf(" AND pa.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	if commodityCode != "" {
		query += fmt.Sprintf(" AND c.code = $%d", argIdx)
		args = append(args, commodityCode)
		argIdx++
	}

	query += " ORDER BY pa.created_at DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying alerts: %w", err)
	}
	defer rows.Close()

	alerts := []PriceAlert{}
	for rows.Next() {
		var a PriceAlert
		err := rows.Scan(
			&a.ID, &a.ClientID, &a.UserID, &a.CommodityID,
			&a.CommodityCode, &a.CommodityName,
			&a.Condition, &a.ThresholdPrice, &a.Status, &a.Notes,
			&a.TriggeredCount, &a.LastTriggeredAt, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning alert row: %w", err)
		}
		alerts = append(alerts, a)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating alert rows: %w", err)
	}

	return alerts, nil
}

func (s *AlertService) TriggerAlert(alertID, clientID, userID int, triggerPrice float64) (*PriceAlert, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	var alert PriceAlert
	var previousStatus string
	err = tx.QueryRow(`
		SELECT pa.id, pa.client_id, pa.user_id, pa.commodity_id,
		       c.code, c.name,
		       pa.condition, pa.threshold_price, pa.status, pa.notes,
		       pa.triggered_count, pa.last_triggered_at, pa.created_at, pa.updated_at
		FROM price_alerts pa
		JOIN commodities c ON c.id = pa.commodity_id
		WHERE pa.id = $1 AND pa.client_id = $2 AND pa.deleted_at IS NULL
		FOR UPDATE OF pa`,
		alertID, clientID,
	).Scan(
		&alert.ID, &alert.ClientID, &alert.UserID, &alert.CommodityID,
		&alert.CommodityCode, &alert.CommodityName,
		&alert.Condition, &alert.ThresholdPrice, &previousStatus, &alert.Notes,
		&alert.TriggeredCount, &alert.LastTriggeredAt, &alert.CreatedAt, &alert.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("selecting alert for update: %w", err)
	}

	now := time.Now()
	err = tx.QueryRow(`
		UPDATE price_alerts
		SET status = 'triggered',
		    triggered_count = triggered_count + 1,
		    last_triggered_at = $1,
		    updated_at = $1
		WHERE id = $2
		RETURNING status, triggered_count, last_triggered_at, updated_at`,
		now, alert.ID,
	).Scan(&alert.Status, &alert.TriggeredCount, &alert.LastTriggeredAt, &alert.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("updating alert: %w", err)
	}

	metadataBytes, err := json.Marshal(map[string]float64{"trigger_price": triggerPrice})
	if err != nil {
		return nil, fmt.Errorf("marshaling trigger metadata: %w", err)
	}
	_, err = tx.Exec(`
		INSERT INTO alert_history (alert_id, changed_by_user_id, change_type, previous_status, new_status, metadata)
		VALUES ($1, $2, 'triggered', $3, 'triggered', $4::jsonb)`,
		alert.ID, userID, previousStatus, string(metadataBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("inserting trigger history: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	return &alert, nil
}
