package weebase

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dracory/api"
)

// handleTableRows handles the request to get table data with pagination
func (wb *App) handleTableRows(w http.ResponseWriter, r *http.Request) {
	if wb.db == nil {
		http.Error(w, "Not connected to any database", http.StatusBadRequest)
		return
	}

	tableName := r.URL.Query().Get("table")
	if tableName == "" {
		http.Error(w, "Table name is required", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	switch {
	case perPage <= 0:
		perPage = 10
	case perPage > 100:
		perPage = 100
	}

	data, err := wb.getTableData(tableName, page, perPage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	api.Respond(w, r, api.SuccessWithData("Table data", map[string]any{"data": data}))
}

// getTableData retrieves paginated data from a table
func (w *App) getTableData(tableName string, page, perPage int) (*TableData, error) {
	var count int64
	if err := w.db.Table(tableName).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("error counting rows: %v", err)
	}

	offset := (page - 1) * perPage

	rows, err := w.db.Table(tableName).
		Offset(offset).
		Limit(perPage).
		Rows()

	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting columns: %v", err)
	}

	// Create a slice of interface{}'s to represent each column,
	// and a second slice to contain pointers to each item in the columns slice
	values := make([]interface{}, len(columns))
	pointers := make([]interface{}, len(columns))
	for i := range values {
		pointers[i] = &values[i]
	}

	var result [][]interface{}
	for rows.Next() {
		if err := rows.Scan(pointers...); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		// Convert each value to a string representation
		row := make([]interface{}, len(columns))
		for i, val := range values {
			switch v := val.(type) {
			case []byte:
				row[i] = string(v)
			case nil:
				row[i] = nil
			default:
				row[i] = v
			}
		}
		result = append(result, row)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return &TableData{
		Columns: columns,
		Rows:    result,
		Total:   count,
		Page:    page,
		PerPage: perPage,
	}, nil
}
