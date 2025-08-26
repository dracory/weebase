package weebase

import (
    "net/http"

    "github.com/dracory/weebase/internal/ports"
)

// Exported thin wrappers to allow the api package to call JSON handlers without exposing internals.

// ListSchemas handles listing schemas (JSON)
func (h *Handler) ListSchemas(w http.ResponseWriter, r *http.Request) { h.handleListSchemas(w, r) }

// ListTables handles listing tables (JSON)
func (h *Handler) ListTables(w http.ResponseWriter, r *http.Request) { h.handleListTables(w, r) }

// BrowseRows handles browsing rows (JSON)
func (h *Handler) BrowseRows(w http.ResponseWriter, r *http.Request) { h.handleBrowseRows(w, r) }

// Ensure Handler satisfies the ports.DBAPI interface
var _ ports.DBAPI = (*Handler)(nil)
