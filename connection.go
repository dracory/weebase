package weebase

import "sync"

// ConnectionProfile stores user-saved connection details (placeholder; expand later).
type ConnectionProfile struct {
	ID     string
	Name   string
	Driver string
	DSN    string
}

// ConnectionStore is a pluggable store for persisted profiles.
type ConnectionStore interface {
	List() []ConnectionProfile
	Save(p ConnectionProfile) error
}

// MemoryConnectionStore is an in-memory implementation.
type MemoryConnectionStore struct {
	mu   sync.RWMutex
	list []ConnectionProfile
}

func NewMemoryConnectionStore() *MemoryConnectionStore { return &MemoryConnectionStore{} }

func (s *MemoryConnectionStore) List() []ConnectionProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ConnectionProfile, len(s.list))
	copy(out, s.list)
	return out
}

func (s *MemoryConnectionStore) Save(p ConnectionProfile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.list = append(s.list, p)
	return nil
}
