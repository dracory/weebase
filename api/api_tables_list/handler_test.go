package api_tables_list_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dracory/weebase/api/api_tables_list"
	"github.com/dracory/weebase/shared/session"
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

		handler := api_tables_list.New(&session.ActiveConnection{
			DB: db,
		})

		req := httptest.NewRequest("GET", "/tables", nil)
		w := httptest.NewRecorder()

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
		handler := api_tables_list.New(nil)

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

	t.Run("invalid database connection type", func(t *testing.T) {
		// Create a mock DB that's not a *gorm.DB
		handler := api_tables_list.New(&session.ActiveConnection{
			DB: "not a gorm.DB",
		})

		req := httptest.NewRequest("GET", "/tables", nil)
		w := httptest.NewRecorder()

		handler.Handle(w, req)

		if status := w.Code; status != http.StatusInternalServerError {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
		}

		var response struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if response.Status != "error" || response.Message != "invalid database connection type" {
			t.Errorf("handler returned unexpected response: %+v", response)
		}
	})

	t.Run("empty database", func(t *testing.T) {
		// Create a new in-memory database without any tables
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("failed to open in-memory SQLite database: %v", err)
		}

		handler := api_tables_list.New(&session.ActiveConnection{
			DB: db,
		})

		req := httptest.NewRequest("GET", "/tables", nil)
		w := httptest.NewRecorder()

		handler.Handle(w, req)

		if status := w.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Parse the response
		var response struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Data    struct {
				Tables []string `json:"tables"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		if response.Status != "success" {
			t.Errorf("expected status 'success', got '%s'", response.Status)
		}

		if len(response.Data.Tables) != 0 {
			t.Errorf("expected 0 tables, got %d", len(response.Data.Tables))
		}
	})
}
