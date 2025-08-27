package weebase

import (
	"fmt"
	"net/url"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// connectDB establishes a database connection based on the provided configuration
func (w *App) connectDB(conn DatabaseConnection) (*gorm.DB, error) {
	switch strings.ToLower(conn.Driver) {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			url.QueryEscape(conn.Username),
			url.QueryEscape(conn.Password),
			conn.Host,
			conn.Port,
			conn.Database,
		)
		return gorm.Open(mysql.Open(dsn), &gorm.Config{})

	case "postgres", "postgresql":
		sslmode := conn.SSLMode
		if sslmode == "" {
			sslmode = "disable"
		}
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			conn.Host,
			conn.Username,
			conn.Password,
			conn.Database,
			conn.Port,
			sslmode,
		)
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})

	case "sqlite", "sqlite3":
		return gorm.Open(sqlite.Open(conn.Database), &gorm.Config{})

	case "sqlserver", "mssql":
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
			url.QueryEscape(conn.Username),
			url.QueryEscape(conn.Password),
			conn.Host,
			conn.Port,
			conn.Database,
		)
		return gorm.Open(sqlserver.Open(dsn), &gorm.Config{})

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", conn.Driver)
	}
}
