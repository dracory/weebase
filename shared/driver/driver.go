package driver

import (
	"errors"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// Validator provides functionality to validate database drivers
type Validator struct {
	drivers Registry
}

// NewValidator creates a new driver validator
func NewValidator(drivers Registry) *Validator {
	return &Validator{
		drivers: drivers,
	}
}

// Validate checks if a driver is valid and enabled
func (v *Validator) Validate(name string) error {
	if name == "" {
		return errors.New("driver is required")
	}
	if !v.drivers.IsEnabled(name) {
		return fmt.Errorf("driver not enabled: %s", name)
	}
	return nil
}

// OpenDBWithDSN opens a database connection using the specified driver and DSN
func OpenDBWithDSN(driver, dsn string) (*gorm.DB, error) {
	switch driver {
	case "postgres", "pg", "postgresql":
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case "mysql", "mariadb":
		return gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "sqlite", "sqlite3":
		return gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	case "sqlserver", "mssql":
		return gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
}
