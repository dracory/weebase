package weebase

import (
	"fmt"

	gormmysql "gorm.io/driver/mysql"
	gormpg "gorm.io/driver/postgres"
	gormsqlite "gorm.io/driver/sqlite"
	gormsqlserver "gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/dracory/weebase/shared/driver"
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

// ValidateDriver checks if a driver is valid and enabled
func (h *Handler) ValidateDriver(name string) error {
	validator := driver.NewValidator(&driverRegistryWrapper{h.drivers})
	return validator.Validate(name)
}

// driverRegistryWrapper wraps DriverRegistry to implement the driver.Registry interface
type driverRegistryWrapper struct {
	drivers *DriverRegistry
}

func (w *driverRegistryWrapper) IsEnabled(name string) bool {
	return w.drivers.IsEnabled(name)
}
