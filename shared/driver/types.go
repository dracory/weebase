package driver

// Registry defines the interface for driver registry operations
type Registry interface {
	IsEnabled(name string) bool
}
