package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type MarketDataRow struct {
	ID              int      `json:"id"`
	ReportDate      string   `json:"report_date"`
	Location        string   `json:"location"`
	Commodity       string   `json:"commodity"`
	Variety         *string  `json:"variety"`
	Package         *string  `json:"package"`
	Origin          *string  `json:"origin"`
	ItemSize        *string  `json:"item_size"`
	LowPrice        *float64 `json:"low_price"`
	HighPrice       *float64 `json:"high_price"`
	MostlyLowPrice  *float64 `json:"mostly_low_price"`
	MostlyHighPrice *float64 `json:"mostly_high_price"`
	Properties      *string  `json:"properties"`
	Comment         *string  `json:"comment"`
}

func autoImportMarketData(database *sql.DB) error {
	// Create table if not exists
	_, err := database.Exec(`
		CREATE TABLE IF NOT EXISTS market_data (
			id                  SERIAL PRIMARY KEY,
			report_date         DATE NOT NULL,
			location            VARCHAR(255) NOT NULL,
			commodity           VARCHAR(255) NOT NULL,
			variety             VARCHAR(255),
			package             VARCHAR(255),
			origin              VARCHAR(255),
			item_size           VARCHAR(255),
			low_price           DECIMAL(10,2),
			high_price          DECIMAL(10,2),
			mostly_low_price    DECIMAL(10,2),
			mostly_high_price   DECIMAL(10,2),
			properties          VARCHAR(255),
			comment             TEXT
		)`)
	if err != nil {
		return fmt.Errorf("creating market_data table: %w", err)
	}

	// Create indexes
	database.Exec("CREATE INDEX IF NOT EXISTS idx_market_data_location ON market_data (location)")
	database.Exec("CREATE INDEX IF NOT EXISTS idx_market_data_date ON market_data (report_date)")
	database.Exec("CREATE INDEX IF NOT EXISTS idx_market_data_commodity ON market_data (commodity)")

	// Find and import all CSV files
	csvPaths := findCSVPaths()
	if len(csvPaths) == 0 {
		log.Println("No CSV files found, skipping market data import")
		return nil
	}

	for _, csvPath := range csvPaths {
		// Check if data for this file's year is already loaded
		year := extractYearFromPath(csvPath)
		if year > 0 {
			var count int
			if err := database.QueryRow(
				"SELECT COUNT(*) FROM market_data WHERE EXTRACT(YEAR FROM report_date) = $1", year,
			).Scan(&count); err != nil {
				return fmt.Errorf("checking market_data count for year %d: %w", year, err)
			}
			if count > 0 {
				log.Printf("Market data for %d already loaded (%d rows), skipping %s", year, count, csvPath)
				continue
			}
		}

		if err := importCSV(database, csvPath); err != nil {
			log.Printf("Warning: failed to import %s: %v", csvPath, err)
		}
	}

	return nil
}

func findCSVPaths() []string {
	csvFiles := []string{
		"AMS_sc_terminal_daily_2023.csv",
		"AMS_sc_terminal_daily_2024.csv",
	}

	var found []string
	for _, name := range csvFiles {
		candidates := []string{
			"../" + name,
			name,
		}

		// Also try relative to source file
		_, filename, _, ok := runtime.Caller(0)
		if ok {
			dir := filepath.Dir(filename)
			candidates = append(candidates, filepath.Join(dir, "..", name))
		}

		for _, path := range candidates {
			if _, err := os.Stat(path); err == nil {
				found = append(found, path)
				break
			}
		}
	}
	return found
}

func extractYearFromPath(path string) int {
	base := filepath.Base(path)
	// Look for 4-digit year pattern in filename
	for i := 0; i <= len(base)-4; i++ {
		if y, err := strconv.Atoi(base[i : i+4]); err == nil && y >= 2000 && y <= 2099 {
			return y
		}
	}
	return 0
}

func importCSV(database *sql.DB, csvPath string) error {
	f, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("opening CSV file: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.LazyQuotes = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("reading CSV header: %w", err)
	}

	// Build column index map
	colIdx := make(map[string]int)
	for i, name := range header {
		colIdx[strings.TrimSpace(name)] = i
	}

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("reading CSV records: %w", err)
	}

	tx, err := database.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO market_data (report_date, location, commodity, variety, package, origin, item_size,
		                         low_price, high_price, mostly_low_price, mostly_high_price, properties, comment)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`)
	if err != nil {
		return fmt.Errorf("preparing insert statement: %w", err)
	}
	defer stmt.Close()

	imported := 0
	for _, record := range records {
		reportDate, err := parseDate(getCol(record, colIdx, "report_date"))
		if err != nil {
			continue // Skip rows with bad dates
		}

		location := getCol(record, colIdx, "location")
		commodity := getCol(record, colIdx, "commodity")
		if location == "" || commodity == "" {
			continue
		}

		_, err = stmt.Exec(
			reportDate,
			location,
			commodity,
			nullableString(getCol(record, colIdx, "variety")),
			nullableString(getCol(record, colIdx, "package")),
			nullableString(getCol(record, colIdx, "origin")),
			nullableString(getCol(record, colIdx, "item_size")),
			nullableFloat(getCol(record, colIdx, "low_price")),
			nullableFloat(getCol(record, colIdx, "high_price")),
			nullableFloat(getCol(record, colIdx, "mostly_low_price")),
			nullableFloat(getCol(record, colIdx, "mostly_high_price")),
			nullableString(getCol(record, colIdx, "properties")),
			nullableString(getCol(record, colIdx, "comment")),
		)
		if err != nil {
			return fmt.Errorf("inserting row: %w", err)
		}
		imported++
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	log.Printf("Imported %d market data rows from %s", imported, csvPath)
	return nil
}

func getCol(record []string, colIdx map[string]int, name string) string {
	idx, ok := colIdx[name]
	if !ok || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}

func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	// CSV format: MM/DD/YYYY
	t, err := time.Parse("01/02/2006", s)
	if err != nil {
		// Try ISO format
		t, err = time.Parse("2006-01-02", s)
	}
	return t, err
}

func nullableString(s string) *string {
	if s == "" || s == "N/A" {
		return nil
	}
	return &s
}

func nullableFloat(s string) *float64 {
	if s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &f
}

func handleListMarketData(w http.ResponseWriter, r *http.Request, database *sql.DB) {
	query := `SELECT id, report_date::text, location, commodity, variety, package, origin, item_size,
	                 low_price, high_price, mostly_low_price, mostly_high_price, properties, comment
	          FROM market_data WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if loc := r.URL.Query().Get("location"); loc != "" {
		query += fmt.Sprintf(" AND location ILIKE $%d", argIdx)
		args = append(args, "%"+loc+"%")
		argIdx++
	}

	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		query += fmt.Sprintf(" AND report_date >= $%d", argIdx)
		args = append(args, startDate)
		argIdx++
	}

	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		query += fmt.Sprintf(" AND report_date <= $%d", argIdx)
		args = append(args, endDate)
		argIdx++
	}

	query += " ORDER BY report_date DESC, location LIMIT 500"

	rows, err := database.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to query market data")
		return
	}
	defer rows.Close()

	data := []MarketDataRow{}
	for rows.Next() {
		var row MarketDataRow
		if err := rows.Scan(&row.ID, &row.ReportDate, &row.Location, &row.Commodity,
			&row.Variety, &row.Package, &row.Origin, &row.ItemSize,
			&row.LowPrice, &row.HighPrice, &row.MostlyLowPrice, &row.MostlyHighPrice,
			&row.Properties, &row.Comment); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to scan market data")
			return
		}
		data = append(data, row)
	}
	if err = rows.Err(); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to iterate market data")
		return
	}

	respondJSON(w, http.StatusOK, data)
}
