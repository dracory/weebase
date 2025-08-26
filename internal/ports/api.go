package ports

import "net/http"

// DBAPI defines the JSON API surface that the api package can call without depending on weebase internals.
type DBAPI interface {
	ListSchemas(http.ResponseWriter, *http.Request)
	ListTables(http.ResponseWriter, *http.Request)
	BrowseRows(http.ResponseWriter, *http.Request)
	CreateTable(http.ResponseWriter, *http.Request)
	// Profiles
	Profiles(http.ResponseWriter, *http.Request)
	ProfilesSave(http.ResponseWriter, *http.Request)
	// Row operations
	InsertRow(http.ResponseWriter, *http.Request)
	UpdateRow(http.ResponseWriter, *http.Request)
	DeleteRow(http.ResponseWriter, *http.Request)
}
