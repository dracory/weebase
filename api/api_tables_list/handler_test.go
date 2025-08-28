package api_tables_list_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dracory/weebase/api/api_tables_list"
	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/types"
	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) (*sql.DB, string) {
	tempFile, err := os.CreateTemp("", "testdb-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tempFile.Close()

	db, err := sql.Open("sqlite3", tempFile.Name())
	if err != nil {
		t.Fatalf("failed to open in-memory SQLite database: %v", err)
	}

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT);
		CREATE TABLE products (id INTEGER PRIMARY KEY, name TEXT, price REAL);
	`)
	if err != nil {
		t.Fatalf("failed to create test tables: %v", err)
	}

	return db, tempFile.Name()
}

func TestTablesList_Handle(t *testing.T) {
	t.Run("successful table list", func(t *testing.T) {
		// Setup test database
		db, dbPath := setupTestDB(t)
		defer os.Remove(dbPath)
		defer db.Close()

		// Create a test config with a session secret
		cfg := types.Config{
			SessionSecret: "test-secret",
		}
		handler := api_tables_list.New(cfg)

		req := httptest.NewRequest("GET", "/tables", nil)
		w := httptest.NewRecorder()

		// Create and set up the session with connection details first
		sess := &session.Session{
			ID:        "test-session",
			CreatedAt: time.Now(),
			Conn: &session.ActiveConnection{
				ID:       "test-connection",
				Driver:   "sqlite3",
				DSN:      dbPath,
				LastUsed: time.Now(),
			},
		}

		// Save session to cookie
		session.SaveSession(w, req, sess, cfg.SessionSecret)
		cookies := w.Result().Cookies()

		// Create a new request with the session cookie
		req = httptest.NewRequest("GET", "/tables", nil)
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		w = httptest.NewRecorder()

		handler.Handle(w, req)

		if status := w.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Parse the response
		var response struct {
			Status  string                 `json:"status"`
			Message string                 `json:"message"`
			Data    map[string]interface{} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		// Check if response indicates success
		if response.Status != "success" {
			t.Fatalf("expected status=success, got %s %s", response.Status, response.Message)
		}

		// Check if tables are in the response
		_, ok := response.Data["tables"].([]interface{})
		if !ok {
			t.Fatal("invalid response format: tables not found")
		}

		// Check additional response fields
		if response.Data["driver"].(string) != "sqlite" {
			t.Errorf("expected driver=sqlite, got %v", response.Data["driver"])
		}

		if response.Data["message"].(string) != "Tables listed successfully" {
			t.Errorf("unexpected message: %v", response.Data["message"])
		}
	})

	t.Run("database not connected", func(t *testing.T) {
		// Create a test config with a session secret
		cfg := types.Config{
			SessionSecret: "test-secret",
		}
		handler := api_tables_list.New(cfg)

		req := httptest.NewRequest("GET", "/tables", nil)
		w := httptest.NewRecorder()

		handler.Handle(w, req)

		if status := w.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		var response struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if response.Status != "error" || response.Message != "not connected to database" {
			t.Errorf("handler returned unexpected response: %+v", response)
		}
	})

	t.Run("unsupported HTTP method", func(t *testing.T) {
		handler := api_tables_list.New(types.Config{SessionSecret: "test"})
		req := httptest.NewRequest("POST", "/tables", nil)
		w := httptest.NewRecorder()

		// Create a session with a valid connection
		db, dbPath := setupTestDB(t)
		defer os.Remove(dbPath)
		defer db.Close()

		sess := session.EnsureSession(w, req, "test")
		sess.Conn = &session.ActiveConnection{
			ID:     "test-connection",
			Driver: "sqlite3",
			DSN:    dbPath,
		}
		req.AddCookie(w.Result().Cookies()[0])

		handler.Handle(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if response.Status != "error" || response.Message != "method not allowed" {
			t.Errorf("unexpected response: %+v", response)
		}
	})
}
