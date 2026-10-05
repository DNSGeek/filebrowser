// Package token keeps the server-side record of revoked session tokens.
//
// Sessions are stateless JWTs, so a token cannot be invalidated by deleting
// it. Instead, every token carries a unique ID (the "jti" claim) and the IDs of
// tokens that must no longer be accepted are recorded here until the token
// would have expired anyway.
package token

import (
	"sync"
	"time"
)

// Backend is the interface to implement for a revoked-token storage.
type Backend interface {
	// Add records id as revoked until the expiry (unix seconds).
	Add(id string, expires int64) error
	// Has reports whether id is recorded as revoked.
	Has(id string) (bool, error)
	// Prune forgets every record that expires before the given unix time.
	Prune(before int64) error
}

// Storage is a revoked-token storage.
type Storage struct {
	back Backend

	// mux makes the check and the insert of Revoke a single step, so that
	// concurrent attempts to redeem the same token cannot both succeed.
	mux sync.Mutex
}

// NewStorage creates a revoked-token storage from a backend.
func NewStorage(back Backend) *Storage {
	return &Storage{back: back}
}

// IsRevoked reports whether the token with the given ID has been revoked.
func (s *Storage) IsRevoked(id string) (bool, error) {
	return s.back.Has(id)
}

// Revoke records the token as revoked until it expires. It reports whether this
// call revoked it, which is false when it had been revoked already.
func (s *Storage) Revoke(id string, expires time.Time) (bool, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	revoked, err := s.back.Has(id)
	if err != nil || revoked {
		return false, err
	}

	// A token that has expired is rejected on its own; dropping those records
	// keeps the list from growing without bound.
	if err := s.back.Prune(time.Now().Unix()); err != nil {
		return false, err
	}

	if err := s.back.Add(id, expires.Unix()); err != nil {
		return false, err
	}
	return true, nil
}
