package weebase

import (
	"github.com/dracory/weebase/shared/types"
)

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

	// SessionSecret is used for CSRF token derivation and session-level secrets.
	SessionSecret string

	// DefaultConnection optionally provides a connection to auto-connect per session.
	DefaultConnection *DefaultConnection

	// PreconfiguredProfiles are loaded into the ConnectionStore at startup.
	PreconfiguredProfiles []types.ConnectionProfile
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

// Option is a functional option to configure Options.
type Option func(*Options)

// WithDrivers sets EnabledDrivers.
func WithDrivers(drivers []string) Option { return func(o *Options) { o.EnabledDrivers = drivers } }

// WithBasePath sets BasePath.
func WithBasePath(p string) Option { return func(o *Options) { o.BasePath = p } }

// WithSafeMode sets SafeModeDefault.
func WithSafeMode(enabled bool) Option { return func(o *Options) { o.SafeModeDefault = enabled } }

// WithAdHoc enables/disables ad-hoc connections.
func WithAdHoc(enabled bool) Option { return func(o *Options) { o.AllowAdHocConnections = enabled } }

// WithReadOnly forces read-only mode.
func WithReadOnly(enabled bool) Option { return func(o *Options) { o.ReadOnlyMode = enabled } }

// WithActionParam sets the action query param key.
func WithActionParam(key string) Option { return func(o *Options) { o.ActionParam = key } }

// WithSessionSecret sets the session secret.
func WithSessionSecret(secret string) Option { return func(o *Options) { o.SessionSecret = secret } }

// DefaultConnection specifies a driver+DSN for auto-connect.
type DefaultConnection struct {
	Driver string
	DSN    string
}
