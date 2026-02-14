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
	ErrNoPositions      = errors.New("no positions found")
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

type Position struct {
	ID            int     `json:"id"`
	ClientID      int     `json:"client_id"`
	UserID        int     `json:"user_id"`
	CommodityID   int     `json:"commodity_id"`
	CommodityCode string  `json:"commodity_code"`
	CommodityName string  `json:"commodity_name"`
	Volume        float64 `json:"volume"`
	Direction     string  `json:"direction"`
	EntryPrice    float64 `json:"entry_price"`
}

type PricePoint struct {
	CommodityID   int     `json:"commodity_id"`
	CommodityCode string  `json:"commodity_code"`
	CommodityName string  `json:"commodity_name"`
	Price         float64 `json:"price"`
	RecordedAt    string  `json:"recorded_at"`
}

type MonthlyPriceSummary struct {
	Year        int     `json:"year"`
	Month       int     `json:"month"`
	MonthName   string  `json:"month_name"`
	SampleCount int     `json:"sample_count"`
	AvgPrice    float64 `json:"avg_price"`
	MinPrice    float64 `json:"min_price"`
	MaxPrice    float64 `json:"max_price"`
	AvgLow      float64 `json:"avg_low"`
	AvgHigh     float64 `json:"avg_high"`
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

func (s *AlertService) ListCommodities() ([]Commodity, error) {
	rows, err := s.db.Query("SELECT id, code, name, unit FROM commodities ORDER BY code")
	if err != nil {
		return nil, fmt.Errorf("querying commodities: %w", err)
	}
	defer rows.Close()

	commodities := []Commodity{}
	for rows.Next() {
		var c Commodity
		if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.Unit); err != nil {
			return nil, fmt.Errorf("scanning commodity row: %w", err)
		}
		commodities = append(commodities, c)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating commodity rows: %w", err)
	}
	return commodities, nil
}

func (s *AlertService) ListPositions(userID, clientID int) ([]Position, error) {
	rows, err := s.db.Query(`
		SELECT p.id, p.client_id, p.user_id, p.commodity_id, c.code, c.name,
		       p.volume, p.direction, p.entry_price
		FROM positions p
		JOIN commodities c ON c.id = p.commodity_id
		WHERE p.user_id = $1 AND p.client_id = $2
		ORDER BY p.id`, userID, clientID)
	if err != nil {
		return nil, fmt.Errorf("querying positions: %w", err)
	}
	defer rows.Close()

	positions := []Position{}
	for rows.Next() {
		var p Position
		if err := rows.Scan(&p.ID, &p.ClientID, &p.UserID, &p.CommodityID,
			&p.CommodityCode, &p.CommodityName,
			&p.Volume, &p.Direction, &p.EntryPrice); err != nil {
			return nil, fmt.Errorf("scanning position row: %w", err)
		}
		positions = append(positions, p)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating position rows: %w", err)
	}
	return positions, nil
}

func (s *AlertService) GetCurrentPrices() ([]PricePoint, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT ON (pd.commodity_id)
		       pd.commodity_id, c.code, c.name, pd.price, pd.recorded_at::text
		FROM price_data pd
		JOIN commodities c ON c.id = pd.commodity_id
		ORDER BY pd.commodity_id, pd.recorded_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("querying current prices: %w", err)
	}
	defer rows.Close()

	prices := []PricePoint{}
	for rows.Next() {
		var p PricePoint
		if err := rows.Scan(&p.CommodityID, &p.CommodityCode, &p.CommodityName, &p.Price, &p.RecordedAt); err != nil {
			return nil, fmt.Errorf("scanning price row: %w", err)
		}
		prices = append(prices, p)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating price rows: %w", err)
	}
	return prices, nil
}

func (s *AlertService) GetMonthlyPriceAnalysis(year int, commodity string) ([]MonthlyPriceSummary, error) {
	monthNames := []string{"", "January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December"}

	rows, err := s.db.Query(`
		SELECT
			EXTRACT(MONTH FROM report_date)::int AS month,
			COUNT(*) AS sample_count,
			ROUND(AVG(COALESCE((mostly_low_price + mostly_high_price) / 2.0,
			                    (low_price + high_price) / 2.0))::numeric, 2) AS avg_price,
			ROUND(MIN(low_price)::numeric, 2) AS min_price,
			ROUND(MAX(high_price)::numeric, 2) AS max_price,
			ROUND(AVG(low_price)::numeric, 2) AS avg_low,
			ROUND(AVG(high_price)::numeric, 2) AS avg_high
		FROM market_data
		WHERE commodity ILIKE '%' || $1 || '%'
		  AND EXTRACT(YEAR FROM report_date) = $2
		  AND low_price IS NOT NULL
		GROUP BY EXTRACT(MONTH FROM report_date)
		ORDER BY month`, commodity, year)
	if err != nil {
		return nil, fmt.Errorf("querying monthly price analysis: %w", err)
	}
	defer rows.Close()

	summaries := []MonthlyPriceSummary{}
	for rows.Next() {
		var s MonthlyPriceSummary
		if err := rows.Scan(&s.Month, &s.SampleCount, &s.AvgPrice,
			&s.MinPrice, &s.MaxPrice, &s.AvgLow, &s.AvgHigh); err != nil {
			return nil, fmt.Errorf("scanning monthly summary row: %w", err)
		}
		s.Year = year
		if s.Month >= 1 && s.Month <= 12 {
			s.MonthName = monthNames[s.Month]
		}
		summaries = append(summaries, s)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating monthly summary rows: %w", err)
	}
	return summaries, nil
}

func (s *AlertService) ResetAlerts(clientID int, commodityCode string) (int, error) {
	result, err := s.db.Exec(`
		UPDATE price_alerts pa
		SET status = 'active', updated_at = NOW()
		FROM commodities c
		WHERE c.id = pa.commodity_id
		  AND pa.client_id = $1
		  AND c.code = $2
		  AND pa.status = 'triggered'
		  AND pa.deleted_at IS NULL`, clientID, commodityCode)
	if err != nil {
		return 0, fmt.Errorf("resetting alerts: %w", err)
	}
	n, _ := result.RowsAffected()
	return int(n), nil
}
