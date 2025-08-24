package weebase

import "sort"

// DriverRegistry tracks enabled database drivers by name.
type DriverRegistry struct {
	enabled map[string]struct{}
}

// NewDriverRegistry constructs a registry from the provided enabled names.
func NewDriverRegistry(enabled []string) *DriverRegistry {
	m := make(map[string]struct{}, len(enabled))
	for _, n := range enabled {
		if n == "" {
			continue
		}
		m[n] = struct{}{}
	}
	return &DriverRegistry{enabled: m}
}

// IsEnabled returns true if the driver name is enabled.
func (r *DriverRegistry) IsEnabled(name string) bool {
	_, ok := r.enabled[name]
	return ok
}

// List returns a sorted list of enabled driver names.
func (r *DriverRegistry) List() []string {
	out := make([]string, 0, len(r.enabled))
	for n := range r.enabled {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
