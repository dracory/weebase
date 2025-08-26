//go:build ignore
// +build ignore

// Deprecated: this file is excluded from builds. HTTP handlers live in the api/ package.
package weebase

import (
    "net/http"
)

// Exported thin wrappers to allow the api package to call JSON handlers without exposing internals.

// ListSchemas handles listing schemas (JSON)
func (h *Handler) ListSchemas(w http.ResponseWriter, r *http.Request) { h.handleListSchemas(w, r) }

// ListTables handles listing tables (JSON)
func (h *Handler) ListTables(w http.ResponseWriter, r *http.Request) { h.handleListTables(w, r) }

// BrowseRows handles browsing rows (JSON)
func (h *Handler) BrowseRows(w http.ResponseWriter, r *http.Request) { h.handleBrowseRows(w, r) }

// CreateTable handles POST create-table (JSON)
func (h *Handler) CreateTable(w http.ResponseWriter, r *http.Request) {
    s := EnsureSession(w, r, h.opts.SessionSecret)
    if s.Conn == nil || s.Conn.DB == nil {
        WriteError(w, r, "not connected")
        return
    }
    if r.Method != http.MethodPost {
        WriteError(w, r, "method not allowed")
        return
    }
    h.executeCreateTable(w, r, s)
}

// Profiles returns saved profiles (JSON)
func (h *Handler) Profiles(w http.ResponseWriter, r *http.Request) { h.handleProfiles(w, r) }

// ProfilesSave saves a profile (JSON)
func (h *Handler) ProfilesSave(w http.ResponseWriter, r *http.Request) { h.handleProfilesSave(w, r) }

// InsertRow inserts a row (JSON)
func (h *Handler) InsertRow(w http.ResponseWriter, r *http.Request) { h.handleInsertRow(w, r) }

// UpdateRow updates a row (JSON)
func (h *Handler) UpdateRow(w http.ResponseWriter, r *http.Request) { h.handleUpdateRow(w, r) }

// DeleteRow deletes a row (JSON)
func (h *Handler) DeleteRow(w http.ResponseWriter, r *http.Request) { h.handleDeleteRow(w, r) }
