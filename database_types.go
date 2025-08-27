package weebase

// DatabaseConnection represents a database connection request
type DatabaseConnection struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"sslmode"`
}

// DatabaseInfo represents information about a database
type DatabaseInfo struct {
	Name string `json:"name"`
}

// TableInfo represents information about a database table
type TableInfo struct {
	Name    string `json:"name"`
	Schema  string `json:"schema,omitempty"`
	Comment string `json:"comment,omitempty"`
}

// ColumnInfo represents information about a table column
type ColumnInfo struct {
	Name            string      `json:"name"`
	Type            string      `json:"type"`
	Nullable        bool        `json:"nullable"`
	PrimaryKey      bool        `json:"primary_key"`
	DefaultValue    interface{} `json:"default_value,omitempty"`
	MaxLength       *int        `json:"max_length,omitempty"`
	NumericPrecision *int       `json:"numeric_precision,omitempty"`
	NumericScale    *int        `json:"numeric_scale,omitempty"`
}

// TableData represents a page of table data
type TableData struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Total   int64           `json:"total"`
	Page    int             `json:"page"`
	PerPage int             `json:"per_page"`
}
