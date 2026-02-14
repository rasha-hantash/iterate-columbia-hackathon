// Run: go run .

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

var db *sql.DB

// User represents the authenticated user from the X-User-ID header.
type User struct {
	ID       int `json:"id"`
	ClientID int `json:"client_id"`
}

// getCurrentUser reads X-User-ID header, returns the user or writes an error.
func getCurrentUser(w http.ResponseWriter, r *http.Request) *User {
	raw := r.Header.Get("X-User-ID")
	if raw == "" {
		http.Error(w, `{"error":"Missing X-User-ID header"}`, 401)
		return nil
	}
	uid, err := strconv.Atoi(raw)
	if err != nil {
		http.Error(w, `{"error":"Invalid X-User-ID"}`, 401)
		return nil
	}
	var u User
	err = db.QueryRow("SELECT id, client_id FROM users WHERE id = $1 AND is_active = TRUE", uid).Scan(&u.ID, &u.ClientID)
	if err != nil {
		http.Error(w, `{"error":"Invalid user"}`, 401)
		return nil
	}
	return &u
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if err := db.Ping(); err != nil {
		http.Error(w, `{"error":"db down"}`, 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "database": "connected"})
}

// corsMiddleware wraps an http.Handler to add CORS headers.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "X-User-ID, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey == "" {
		log.Println("Warning: ANTHROPIC_API_KEY not set, /analyze-positions endpoint will be disabled")
	}

	var err error
	db, err = sql.Open("postgres", "host=localhost port=5432 user=edge password=edge_local dbname=edge_interview sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	svc := NewAlertService(db)
	handler := NewAlertHandler(svc)
	var aiHandler *AIHandler
	if anthropicKey != "" {
		aiHandler = NewAIHandler(svc, anthropicKey)
	}

	// Auto-import market data on startup
	if err := autoImportMarketData(db); err != nil {
		log.Printf("Warning: market data import failed: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler)

	mux.HandleFunc("/commodities", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		handler.HandleListCommodities(w, r)
	})

	mux.HandleFunc("/positions", func(w http.ResponseWriter, r *http.Request) {
		user := getCurrentUser(w, r)
		if user == nil {
			return
		}
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		handler.HandleListPositions(w, r, user.ID, user.ClientID)
	})

	mux.HandleFunc("/prices", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		handler.HandleGetPrices(w, r)
	})

	mux.HandleFunc("/alerts", func(w http.ResponseWriter, r *http.Request) {
		user := getCurrentUser(w, r)
		if user == nil {
			return
		}

		switch r.Method {
		case http.MethodPost:
			handler.HandleCreateAlert(w, r, user.ID, user.ClientID)
		case http.MethodGet:
			handler.HandleListAlerts(w, r, user.ClientID)
		default:
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/alerts/", func(w http.ResponseWriter, r *http.Request) {
		user := getCurrentUser(w, r)
		if user == nil {
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/alerts/")
		parts := strings.Split(path, "/")

		if len(parts) == 2 && parts[1] == "trigger" && r.Method == http.MethodPost {
			alertID, err := strconv.Atoi(parts[0])
			if err != nil {
				respondError(w, http.StatusBadRequest, "Invalid alert ID")
				return
			}
			handler.HandleTriggerAlert(w, r, alertID, user.ID, user.ClientID)
			return
		}

		respondError(w, http.StatusNotFound, "Not found")
	})

	mux.HandleFunc("/analyze-positions", func(w http.ResponseWriter, r *http.Request) {
		user := getCurrentUser(w, r)
		if user == nil {
			return
		}
		if r.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		if aiHandler == nil {
			respondError(w, http.StatusServiceUnavailable, "AI analysis is not configured (ANTHROPIC_API_KEY not set)")
			return
		}
		aiHandler.HandleAnalyzePositions(w, r, user.ID, user.ClientID)
	})

	mux.HandleFunc("/market-data", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		handleListMarketData(w, r, db)
	})

	fmt.Println("Server running on http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", corsMiddleware(mux)))
}
