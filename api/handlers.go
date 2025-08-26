package api

import (
	"net/http"

	"github.com/dracory/weebase/internal/ports"
)

// TablesList handles GET tables list via the ports.DBAPI interface.
func TablesList(api ports.DBAPI, w http.ResponseWriter, r *http.Request) { api.ListTables(w, r) }

// SchemasList handles GET schemas list via the ports.DBAPI interface.
func SchemasList(api ports.DBAPI, w http.ResponseWriter, r *http.Request) { api.ListSchemas(w, r) }

// RowsBrowse handles browsing rows via the ports.DBAPI interface.
func RowsBrowse(api ports.DBAPI, w http.ResponseWriter, r *http.Request) { api.BrowseRows(w, r) }
