package api_table_create_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dracory/weebase/api/api_table_create"
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

	return db, tempFile.Name()
}

func TestTableCreate_Handle(t *testing.T) {
	setupTestHandler := func(t *testing.T, safeMode bool) (*api_table_create.TableCreate, *sql.DB, string, *http.Cookie) {
		db, dbPath := setupTestDB(t)

		// Create a test config with a session secret
		cfg := types.Config{
			SessionSecret: "test-secret",
		}

		// Create and save session
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

		// Save session to a cookie
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		session.SaveSession(w, req, sess, "test-secret")
		cookies := w.Result().Cookies()
		if len(cookies) == 0 {
			t.Fatal("failed to create session cookie")
		}

		return api_table_create.New(cfg, safeMode), db, dbPath, cookies[0]
	}

	t.Run("successful table creation", func(t *testing.T) {
		handler, db, dbPath, cookie := setupTestHandler(t, false)
		defer os.Remove(dbPath)
		defer db.Close()

		// Create a form with table creation data
		form := url.Values{
			"table":         {"test_table"},
			"col_name[]":    {"id", "name"},
			"col_type[]":    {"INTEGER", "TEXT"},
			"col_length[]":  {"", "255"},
			"col_nullable[]": {"1"}, // Index 1 is the second column (name)
			"col_pk[]":      {"0"}, // Index 0 is the first column (id)
			"col_ai[]":      {"0"}, // Index 0 is the first column (id)
		}

		// Create request with form data and session cookie
		req := httptest.NewRequest("POST", "/tables", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(cookie)

		w := httptest.NewRecorder()
		handler.Handle(w, req)

		if status := w.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Verify the table was created
		rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='test_table'")
		if err != nil {
			t.Fatalf("failed to query for created table: %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			t.Error("table was not created in the database")
		}
	})

	t.Run("safe mode requires confirmation", func(t *testing.T) {
		handler, db, dbPath, cookie := setupTestHandler(t, true) // Enable safe mode
		defer os.Remove(dbPath)
		defer db.Close()

		form := url.Values{
			"table": {"test_table"},
			"col_name[]": {"id"},
			"col_type[]": {"INTEGER"},
			"col_length[]": {""},
			"col_nullable[]": {},
			"col_pk[]": {},
			"col_ai[]": {},
		}

		req := httptest.NewRequest("POST", "/tables", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(cookie)
		w := httptest.NewRecorder()

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

		if response.Status != "error" || !strings.Contains(response.Message, "confirmation required") {
			t.Errorf("unexpected response: %+v", response)
		}
	})

	t.Run("missing table name", func(t *testing.T) {
		handler, db, dbPath, cookie := setupTestHandler(t, false)
		defer os.Remove(dbPath)
		defer db.Close()

		form := url.Values{
			"col_name[]": {"id"},
			"col_type[]":  {"INTEGER"},
			"col_length[]": {""},
		}

		req := httptest.NewRequest("POST", "/tables", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(cookie)
		w := httptest.NewRecorder()

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

		if response.Status != "error" || response.Message != "table name is required" {
			t.Errorf("unexpected response: %+v", response)
		}
	})

	t.Run("missing columns", func(t *testing.T) {
		handler, db, dbPath, cookie := setupTestHandler(t, false)
		defer os.Remove(dbPath)
		defer db.Close()

		form := url.Values{
			"table": {"test_table"},
			"col_name[]": {},
			"col_type[]": {},
		}

		req := httptest.NewRequest("POST", "/tables", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(cookie)
		w := httptest.NewRecorder()

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

		if response.Status != "error" || response.Message != "at least one column is required" {
			t.Errorf("unexpected response: %+v", response)
		}
	})
}
