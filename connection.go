package weebase

import (
	"sync"

	"github.com/dracory/weebase/shared/types"
)

// MemoryConnectionStore is an in-memory implementation of types.ConnectionStore.
type MemoryConnectionStore struct {
	mu   sync.RWMutex
	list []types.ConnectionProfile
}

// NewMemoryConnectionStore creates a new in-memory connection store.
func NewMemoryConnectionStore() *MemoryConnectionStore { 
	return &MemoryConnectionStore{} 
}

// List returns all stored connection profiles.
func (s *MemoryConnectionStore) List() []types.ConnectionProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]types.ConnectionProfile, len(s.list))
	copy(out, s.list)
	return out
}

// Save stores a new connection profile.
func (s *MemoryConnectionStore) Save(p types.ConnectionProfile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.list = append(s.list, p)
	return nil
}

// Get retrieves a connection profile by ID.
func (s *MemoryConnectionStore) Get(id string) (types.ConnectionProfile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.list {
		if p.ID == id {
			return p, true
		}
	}
	return types.ConnectionProfile{}, false
}
