package ports

import "net/http"

// DBAPI defines the JSON API surface that the api package can call without depending on weebase internals.
type DBAPI interface {
	ListSchemas(http.ResponseWriter, *http.Request)
	ListTables(http.ResponseWriter, *http.Request)
	BrowseRows(http.ResponseWriter, *http.Request)
}
