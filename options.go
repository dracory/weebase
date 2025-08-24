package weebase

// Options configures the embedded Adminer-like handler.
type Options struct {
	// EnabledDrivers lists enabled database drivers (e.g., postgres, mysql, sqlite, sqlserver)
	EnabledDrivers []string

	// SafeModeDefault turns on DDL/destructive guardrails by default
	SafeModeDefault bool

	// AllowAdHocConnections allows runtime connection entry via UI
	AllowAdHocConnections bool

	// ReadOnlyMode forces read-only operations regardless of DB grants
	ReadOnlyMode bool

	// ActionParam is the query param that selects behavior (default: "action")
	ActionParam string

	// BasePath is the mount path for the handler (for generating links), e.g. "/db"
	BasePath string
}

// withDefaults applies default values to Options.
func (o Options) withDefaults() Options {
	if o.ActionParam == "" {
		o.ActionParam = "action"
	}
	if o.BasePath == "" {
		o.BasePath = "/db"
	}
	return o
}
