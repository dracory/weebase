package api

import (
	"net/http"
)

// HandlerFunc is a plain HTTP handler signature passed in by the router.
type HandlerFunc func(http.ResponseWriter, *http.Request)

// TablesList forwards to the provided handler function.
func TablesList(fn HandlerFunc) http.HandlerFunc { return http.HandlerFunc(fn) }

// SchemasList forwards to the provided handler function.
func SchemasList(fn HandlerFunc) http.HandlerFunc { return http.HandlerFunc(fn) }

// RowsBrowse forwards to the provided handler function.
func RowsBrowse(fn HandlerFunc) http.HandlerFunc { return http.HandlerFunc(fn) }


// Profiles forwards to the provided handler function.
func Profiles(fn HandlerFunc) http.HandlerFunc { return http.HandlerFunc(fn) }

// ProfilesSave forwards to the provided handler function.
func ProfilesSave(fn HandlerFunc) http.HandlerFunc { return http.HandlerFunc(fn) }

// InsertRow forwards to the provided handler function.
func InsertRow(fn HandlerFunc) http.HandlerFunc { return http.HandlerFunc(fn) }

// UpdateRow forwards to the provided handler function.
func UpdateRow(fn HandlerFunc) http.HandlerFunc { return http.HandlerFunc(fn) }

// DeleteRow forwards to the provided handler function.
func DeleteRow(fn HandlerFunc) http.HandlerFunc { return http.HandlerFunc(fn) }
