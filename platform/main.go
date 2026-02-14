// Run: go run .

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

func main() {
	var err error
	db, err = sql.Open("postgres", "host=localhost port=5432 user=edge password=edge_local dbname=edge_interview sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	svc := NewAlertService(db)
	handler := NewAlertHandler(svc)

	http.HandleFunc("/health", healthHandler)

	http.HandleFunc("/alerts", func(w http.ResponseWriter, r *http.Request) {
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

	http.HandleFunc("/alerts/", func(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println("Server running on http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
