package weebase

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	gormmysql "gorm.io/driver/mysql"
	gormpg "gorm.io/driver/postgres"
	gormsqlite "gorm.io/driver/sqlite"
	gormsqlserver "gorm.io/driver/sqlserver"
)

// OpenGORM opens a GORM DB for the given driver and DSN.
// Supported drivers: postgres, mysql, sqlite, sqlserver.
func OpenGORM(driver, dsn string) (*gorm.DB, error) {
	switch driver {
	case "postgres", "pg", "postgresql":
		cfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
		return gorm.Open(gormpg.Open(dsn), cfg)
	case "mysql", "mariadb":
		cfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
		return gorm.Open(gormmysql.Open(dsn), cfg)
	case "sqlite", "sqlite3":
		cfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
		return gorm.Open(gormsqlite.Open(dsn), cfg)
	case "sqlserver", "mssql":
		cfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
		return gorm.Open(gormsqlserver.Open(dsn), cfg)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
}

// ValidateDriver checks if name is among enabled drivers in the registry.
func (h *Handler) ValidateDriver(name string) error {
	if name == "" {
		return errors.New("driver is required")
	}
	if !h.drivers.IsEnabled(name) {
		return fmt.Errorf("driver not enabled: %s", name)
	}
	return nil
}
