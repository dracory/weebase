package api_tables_list_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dracory/weebase/api/api_tables_list"
	"github.com/dracory/weebase/shared/session"
	"github.com/dracory/weebase/shared/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory SQLite database: %v", err)
	}

	// Create test tables
	db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	db.Exec("CREATE TABLE products (id INTEGER PRIMARY KEY, name TEXT, price REAL)")

	return db
}

func TestTablesList_Handle(t *testing.T) {
	t.Run("successful table list", func(t *testing.T) {
		// Setup test database
		db := setupTestDB(t)

		// Create a test config with a session secret
		cfg := types.Config{
			SessionSecret: "test-secret",
		}
		handler := api_tables_list.New(cfg)

		req := httptest.NewRequest("GET", "/tables", nil)
		w := httptest.NewRecorder()

		// Create a session using EnsureSession which will handle the cookie
		sess := session.EnsureSession(w, req, cfg.SessionSecret)
		sess.Conn = &session.ActiveConnection{
			ID:     "test-connection",
			DB:     db,
			Driver: "sqlite",
		}

		// Apply the cookie to the request
		req.AddCookie(w.Result().Cookies()[0])

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

		// Check if both tables are in the response
		tables, ok := response.Data["tables"].([]interface{})
		if !ok {
			t.Fatal("invalid response format: tables not found")
		}

		if len(tables) != 2 {
			t.Errorf("expected 2 tables, got %d", len(tables))
		}

		// Convert to map for easier checking
		tableMap := make(map[string]bool)
		for _, t := range tables {
			tableMap[t.(string)] = true
		}

		if !tableMap["users"] || !tableMap["products"] {
			t.Errorf("expected tables 'users' and 'products' in response, got %v", tables)
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

		if status := w.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
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
}
