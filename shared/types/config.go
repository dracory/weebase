package types

// Config contains the configuration for web handlers
type Config struct {
	// BasePath is the base URL path for the application
	BasePath string
	// ActionParam is the query parameter used for actions
	ActionParam string
	// EnabledDrivers is the list of enabled database drivers
	EnabledDrivers []string
	// AllowAdHocConnections specifies if ad-hoc connections are allowed
	AllowAdHocConnections bool
	// SafeModeDefault specifies if safe mode is enabled by default
	SafeModeDefault bool
	// SessionSecret is the secret used for session management
	SessionSecret string
}
