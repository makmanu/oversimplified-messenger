package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type Message struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Message string `json:"message"`
}

type StoredMessage struct {
	ID        int    `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

var db *sql.DB

// Initialize the database connection and create the messages table if it doesn't exist
func init() {
	var err error

	// Get database path from environment variable or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		// Create data directory if it doesn't exist
		dataDir := "./data"
		if _, err := os.Stat(dataDir); os.IsNotExist(err) {
			os.MkdirAll(dataDir, 0755)
		}
		dbPath = filepath.Join(dataDir, "messages.db")
	} else {
		// Ensure directory exists for the database file
		dir := filepath.Dir(dbPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.MkdirAll(dir, 0755)
		}
	}

	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	// Create table if it doesn't exist
	createTableSQL := `CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		from_user TEXT NOT NULL,
		to_user TEXT NOT NULL,
		message TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}

// Enable CORS for all responses
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// Handle POST requests to save messages
func handleMessage(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == http.MethodOptions {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg Message
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Validate input
	if msg.From == "" || msg.To == "" || msg.Message == "" {
		http.Error(w, "From, to, and message are required", http.StatusBadRequest)
		return
	}

	// Insert into database
	insertSQL := `INSERT INTO messages (from_user, to_user, message) VALUES (?, ?, ?)`
	result, err := db.Exec(insertSQL, msg.From, msg.To, msg.Message)
	if err != nil {
		http.Error(w, "Failed to save message", http.StatusInternalServerError)
		log.Printf("Database error: %v\n", err)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to retrieve ID", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      id,
		"from":    msg.From,
		"to":      msg.To,
		"message": msg.Message,
		"status":  "saved",
	})
}

// Serve the HTML menu page at the root URL
func handleUIIndex(w http.ResponseWriter, r *http.Request) { 
	html, err := os.ReadFile("html/menu")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, string(html))
}

// Handle GET requests to retrieve messages for a specific recipient
func handleGetMessages(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == http.MethodOptions {
		return
	}

	to := r.URL.Query().Get("to")

	if to == "" {
		http.Error(w, "Parameter 'to' is required", http.StatusBadRequest)
		return
	}

	query := `SELECT id, from_user, to_user, message, created_at FROM messages WHERE to_user = ? ORDER BY created_at DESC`
	rows, err := db.Query(query, to)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Printf("Database error: %v\n", err)
		return
	}
	defer rows.Close()

	var messages []StoredMessage
	for rows.Next() {
		var msg StoredMessage
		err := rows.Scan(&msg.ID, &msg.From, &msg.To, &msg.Message, &msg.CreatedAt)
		if err != nil {
			log.Printf("Scan error: %v\n", err)
			continue
		}
		messages = append(messages, msg)
	}

	if messages == nil {
		messages = []StoredMessage{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// Main function to start the server
func main() {
	defer db.Close()

	// Setup all routes on port 50505
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleUIIndex)
	mux.HandleFunc("/messages", handleMessage)
	mux.HandleFunc("/api/messages", handleGetMessages)

	fmt.Println("Server listening on http://localhost:50505")

	if err := http.ListenAndServe(":50505", mux); err != nil {
		log.Fatal(err)
	}
}
