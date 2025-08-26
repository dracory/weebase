package driver

import (
	"errors"
	"fmt"
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
